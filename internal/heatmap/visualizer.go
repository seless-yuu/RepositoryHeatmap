package heatmap

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/repositoryheatmap/internal/utils"
	"github.com/repositoryheatmap/pkg/models"
)

// Visualizer はヒートマップの可視化を担当する構造体
type Visualizer struct {
	outputDir  string
	outputType string
	stats      *models.RepositoryStats
	repoPath   string // リポジトリのルートパス
}

// NewVisualizer は新しいVisualizerインスタンスを作成する
func NewVisualizer(outputDir, outputType string, stats *models.RepositoryStats) *Visualizer {
	return &Visualizer{
		outputDir:  outputDir,
		outputType: strings.ToLower(outputType),
		stats:      stats,
		repoPath:   "", // デフォルトは空文字
	}
}

// SetRepoPath はリポジトリパスを設定する
func (v *Visualizer) SetRepoPath(repoPath string) {
	v.repoPath = repoPath
}

// Visualize はヒートマップデータを可視化する
func (v *Visualizer) Visualize() error {
	// 出力タイプのチェック
	if v.outputType != "svg" && v.outputType != "webp" {
		return fmt.Errorf("サポートされていない出力形式です: %s (svg または webp を使用してください)", v.outputType)
	}

	// リポジトリ全体のヒートマップを生成
	if err := v.generateRepositoryHeatmap(); err != nil {
		return fmt.Errorf("リポジトリヒートマップの生成に失敗しました: %w", err)
	}

	// ファイル単位のヒートマップを生成
	if err := v.generateFileHeatmaps(); err != nil {
		return fmt.Errorf("ファイルヒートマップの生成に失敗しました: %w", err)
	}

	return nil
}

// generateRepositoryHeatmap はリポジトリ全体のヒートマップを生成する
func (v *Visualizer) generateRepositoryHeatmap() error {
	// 出力ファイルパスを作成
	outputPath := filepath.Join(v.outputDir, fmt.Sprintf("%s-repository-heatmap.%s", v.stats.RepositoryName, v.outputType))

	// SVG形式の場合
	if v.outputType == "svg" {
		return v.generateSVGRepositoryHeatmap(outputPath)
	}

	// WebP形式の場合
	return v.generateWebPRepositoryHeatmap(outputPath)
}

// generateSVGRepositoryHeatmap はSVG形式のリポジトリヒートマップを生成する
func (v *Visualizer) generateSVGRepositoryHeatmap(outputPath string) error {
	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("SVGファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// ヒートマップの作成に必要なファイルデータを準備
	files := getSortedFilesByHeat(v.stats)

	// 描画パラメータ - ファイル数に応じて調整
	const (
		marginTop        = 80
		marginLeft       = 50
		marginRight      = 50
		marginBottom     = 50
		barHeight        = 20
		barGap           = 5
		maxBarWidth      = 1100               // 最大バー幅は固定
		fileHeaderHeight = 100                // ヘッダー部分の高さ
		fileEntryHeight  = barHeight + barGap // 1ファイルあたりの高さ
	)

	// 表示するファイル数を決定（最大で100ファイルまで）
	maxFiles := 100
	if len(files) > maxFiles {
		files = files[:maxFiles]
	}

	// キャンバスのサイズをファイル数に応じて調整
	canvasWidth := 1200
	// ファイル数 × (バーの高さ + 間隔) + ヘッダーと余白
	canvasHeight := len(files)*fileEntryHeight + fileHeaderHeight + marginTop + marginBottom

	// 最小高さを確保
	if canvasHeight < 800 {
		canvasHeight = 800
	}

	// SVGヘッダーを書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <text x="%d" y="40" font-size="24" font-family="Arial">%s リポジトリヒートマップ</text>
  <text x="%d" y="70" font-size="14" font-family="Arial">コミット数: %d, ファイル数: %d, 期間: %s 〜 %s</text>
  <text x="%d" y="90" font-size="12" font-family="Arial">※ 変更頻度上位 %d ファイルを表示（全 %d ファイル中）</text>
  
  <!-- 凡例 -->
  <g transform="translate(%d, 30)">
    <text x="0" y="0" font-size="12" font-family="Arial">変更頻度:</text>
    <rect x="80" y="-10" width="20" height="12" fill="#0000FF" />
    <text x="105" y="0" font-size="10" font-family="Arial">低</text>
    <rect x="130" y="-10" width="20" height="12" fill="#0066FF" />
    <rect x="180" y="-10" width="20" height="12" fill="#00FF00" />
    <rect x="230" y="-10" width="20" height="12" fill="#FFCC00" />
    <rect x="280" y="-10" width="20" height="12" fill="#FF0000" />
    <text x="305" y="0" font-size="10" font-family="Arial">高</text>
  </g>

  <!-- スクロール可能なヒートマップ -->
  <g transform="translate(%d, %d)">
`,
		canvasWidth, canvasHeight,
		canvasWidth/2, v.stats.RepositoryName,
		canvasWidth/2, v.stats.CommitCount, v.stats.FileCount,
		v.stats.FirstCommitAt.Format("2006/01/02"), v.stats.LastCommitAt.Format("2006/01/02"),
		canvasWidth/2, len(files), v.stats.FileCount,
		marginLeft, marginLeft, marginTop)

	// 各ファイルのヒートマップバーを描画
	for i, fileInfo := range files {
		y := i * (barHeight + barGap)
		barWidth := int(fileInfo.HeatLevel * float64(maxBarWidth))
		if barWidth < 10 {
			barWidth = 10 // 最小幅を設定
		}

		// 色を取得
		color := GetHeatColor(fileInfo.HeatLevel)

		// パスを短縮（長すぎる場合）
		displayPath := fileInfo.FilePath
		if len(displayPath) > 50 {
			displayPath = "..." + displayPath[len(displayPath)-47:]
		}

		// バーを描画
		fmt.Fprintf(file, `    <!-- %s -->
    <text x="-10" y="%d" text-anchor="end" font-size="12" font-family="monospace">%d</text>
    <rect x="0" y="%d" width="%d" height="%d" fill="%s" />
    <text x="%d" y="%d" font-size="12" font-family="monospace">%s (%d)</text>
`,
			fileInfo.FilePath,
			y+barHeight/2+4, fileInfo.ChangeCount,
			y, barWidth, barHeight, color,
			barWidth+5, y+barHeight/2+4, displayPath, fileInfo.ChangeCount)
	}

	// SVGフッターを書き込む
	fmt.Fprintf(file, `  </g>
</svg>`)

	fmt.Printf("リポジトリヒートマップをSVGに保存しました: %s\n", outputPath)
	return nil
}

// generateWebPRepositoryHeatmap はWebP形式のリポジトリヒートマップを生成する
func (v *Visualizer) generateWebPRepositoryHeatmap(outputPath string) error {
	// TODO: 実際のWebP生成処理の実装
	// Go言語でWebP形式を生成するには、別のライブラリが必要になる可能性があります

	// ダミーのファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("WebPファイルの作成に失敗しました: %w", err)
	}
	file.Close()

	fmt.Printf("リポジトリヒートマップをWebPに保存しました: %s\n", outputPath)
	return nil
}

// generateFileHeatmaps は各ファイルのヒートマップを生成する
func (v *Visualizer) generateFileHeatmaps() error {
	// ファイルヒートマップの出力ディレクトリを作成
	fileHeatmapDir := filepath.Join(v.outputDir, "file-heatmaps")
	if err := utils.EnsureDirectoryExists(fileHeatmapDir); err != nil {
		return fmt.Errorf("ファイルヒートマップディレクトリの作成に失敗しました: %w", err)
	}

	// 各ファイルのヒートマップを生成
	for filePath, fileInfo := range v.stats.Files {
		// ファイル名のハッシュを生成して一意のファイル名にする
		safeFileName := utils.SanitizePath(filePath)

		// 出力ファイルパスを作成 - フラットな構造を使用
		outputPath := filepath.Join(fileHeatmapDir, safeFileName+"."+v.outputType)

		// ファイルの変更頻度に基づく色を取得
		color := GetHeatColor(fileInfo.HeatLevel)

		// SVG形式の場合
		if v.outputType == "svg" {
			if err := v.generateSVGFileHeatmap(outputPath, filePath, fileInfo, color); err != nil {
				return fmt.Errorf("ファイル '%s' のSVGヒートマップ生成に失敗しました: %w", filePath, err)
			}
		} else {
			// WebP形式の場合
			if err := v.generateWebPFileHeatmap(outputPath, filePath, fileInfo, color); err != nil {
				return fmt.Errorf("ファイル '%s' のWebPヒートマップ生成に失敗しました: %w", filePath, err)
			}
		}
	}

	fmt.Printf("ファイルヒートマップを保存しました: %s\n", fileHeatmapDir)
	return nil
}

// readFileContent はファイルの内容を読み込む
func (v *Visualizer) readFileContent(filePath string) ([]string, error) {
	// リポジトリパスが設定されていない場合はエラー
	if v.repoPath == "" {
		return nil, fmt.Errorf("リポジトリパスが設定されていません")
	}

	// ファイルのフルパスを作成
	fullPath := filepath.Join(v.repoPath, filePath)

	// ファイルを開く
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開けませんでした: %w", err)
	}
	defer file.Close()

	// ファイル内容を読み込む
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ファイル読み込み中にエラーが発生しました: %w", err)
	}

	return lines, nil
}

// escapeSpecialChars はSVG内で特殊文字をエスケープする
func escapeSpecialChars(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&apos;")
	return text
}

// generateSVGFileHeatmap はSVG形式のファイルヒートマップを生成する
func (v *Visualizer) generateSVGFileHeatmap(outputPath, filePath string, fileInfo models.FileChangeInfo, color string) error {
	// 必要に応じてディレクトリを作成
	if err := utils.EnsureDirectoryExists(filepath.Dir(outputPath)); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %s: %w", filepath.Dir(outputPath), err)
	}

	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("SVGファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// 行変更頻度データを取得
	lineChanges := make([]struct {
		LineNum int
		Count   int
	}, 0, len(fileInfo.LineChanges))

	for lineNum, count := range fileInfo.LineChanges {
		lineChanges = append(lineChanges, struct {
			LineNum int
			Count   int
		}{
			LineNum: lineNum,
			Count:   count,
		})
	}

	// 行番号でソート
	sort.Slice(lineChanges, func(i, j int) bool {
		return lineChanges[i].LineNum < lineChanges[j].LineNum
	})

	// 著者情報を取得
	authors := make([]struct {
		Name  string
		Count int
	}, 0, len(fileInfo.Authors))

	for name, count := range fileInfo.Authors {
		authors = append(authors, struct {
			Name  string
			Count int
		}{
			Name:  name,
			Count: count,
		})
	}

	// 貢献度でソート
	sort.Slice(authors, func(i, j int) bool {
		return authors[i].Count > authors[j].Count
	})

	// ファイルの内容を読み込む（失敗しても続行）
	fileContent, _ := v.readFileContent(filePath)

	// 縦幅の調整（ファイル内容に応じて）
	height := 400 // 基本高さ

	// ファイル内容がある場合は高さを増やす
	if len(fileContent) > 0 {
		// 各行18pxとして計算（最大50行まで）
		contentHeight := len(fileContent)
		if contentHeight > 50 {
			contentHeight = 50
		}
		height += contentHeight*18 + 60 // 説明用のヘッダーなどを考慮して+60px
	}

	// SVGの基本構造を書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="1000" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <text x="10" y="30" font-size="16" font-family="Arial">ファイル: %s</text>
  <text x="10" y="50" font-size="14" font-family="Arial">変更回数: %d / 最終変更: %s</text>
  
  <!-- ヒートマップバー -->
  <rect x="10" y="70" width="980" height="30" fill="%s" />
  <text x="500" y="90" text-anchor="middle" font-size="14" fill="white">変更頻度レベル: %.2f</text>
  
  <!-- 著者情報 -->
  <text x="10" y="120" font-size="14" font-family="Arial">主な貢献者:</text>`,
		height, filePath, fileInfo.ChangeCount, fileInfo.LastModified.Format("2006/01/02"),
		color, fileInfo.HeatLevel)

	// 著者リストを表示（最大5人まで）
	authorLimit := 5
	if len(authors) < authorLimit {
		authorLimit = len(authors)
	}

	for i := 0; i < authorLimit; i++ {
		fmt.Fprintf(file, `
  <text x="20" y="%d" font-size="12" font-family="Arial">%s (%d)</text>`,
			140+i*20, authors[i].Name, authors[i].Count)
	}

	// 最大変更回数を計算（色の濃淡用）
	maxCount := 0
	lineCountMap := make(map[int]int) // 行番号→変更回数のマップ

	for _, lc := range lineChanges {
		if lc.Count > maxCount {
			maxCount = lc.Count
		}
		lineCountMap[lc.LineNum] = lc.Count
	}

	// ファイル内容の表示（ある場合）
	if len(fileContent) > 0 {
		fmt.Fprint(file, `
  
  <!-- ファイル内容 -->
  <text x="10" y="200" font-size="14" font-family="Arial">ファイル内容（色は変更頻度を表します）:</text>
  <rect x="10" y="210" width="980" height="30" fill="#ddd" />
  <text x="20" y="230" font-size="12" font-family="monospace">行番号</text>
  <text x="100" y="230" font-size="12" font-family="monospace">内容</text>
  <text x="900" y="230" font-size="12" font-family="monospace">変更頻度</text>
  
  <g transform="translate(10, 250)">`)

		// 表示する行数を制限（最大50行まで）
		lineLimit := 50
		if len(fileContent) < lineLimit {
			lineLimit = len(fileContent)
		}

		for i := 0; i < lineLimit; i++ {
			lineNum := i + 1 // 1-indexed
			lineHeatLevel := 0.0
			lineCount := lineCountMap[lineNum]

			if maxCount > 0 && lineCount > 0 {
				lineHeatLevel = float64(lineCount) / float64(maxCount)
			}

			// 行の背景色を決定
			bgColor := "#fff" // デフォルトは白
			if lineHeatLevel > 0 {
				bgColor = GetHeatColor(lineHeatLevel)
			}

			// 行の内容をエスケープ
			lineContent := ""
			if i < len(fileContent) {
				lineContent = escapeSpecialChars(fileContent[i])
				// 長すぎる行は省略
				if len(lineContent) > 100 {
					lineContent = lineContent[:97] + "..."
				}
			}

			// 行を描画
			fmt.Fprintf(file, `
    <g>
      <rect x="0" y="%d" width="980" height="18" fill="%s" opacity="0.7" />
      <text x="20" y="%d" font-size="12" font-family="monospace">%d</text>
      <text x="100" y="%d" font-size="12" font-family="monospace">%s</text>
      <text x="900" y="%d" font-size="12" font-family="monospace">%d</text>
    </g>`,
				i*18, bgColor, i*18+14, lineNum, i*18+14, lineContent, i*18+14, lineCount)
		}

		fmt.Fprint(file, `
  </g>`)
	}

	// SVG終了タグ
	fmt.Fprint(file, `
</svg>`)

	return nil
}

// generateWebPFileHeatmap はWebP形式のファイルヒートマップを生成する
func (v *Visualizer) generateWebPFileHeatmap(outputPath, filePath string, fileInfo models.FileChangeInfo, color string) error {
	// 必要に応じてディレクトリを作成
	if err := utils.EnsureDirectoryExists(filepath.Dir(outputPath)); err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %s: %w", filepath.Dir(outputPath), err)
	}

	// TODO: 実際のWebP生成処理の実装
	// Go言語でWebP形式を生成するには、別のライブラリが必要になる可能性があります

	// ダミーのファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("WebPファイルの作成に失敗しました: %w", err)
	}
	file.Close()

	return nil
}
