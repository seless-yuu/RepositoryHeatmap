package heatmap

import (
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
}

// NewVisualizer は新しいVisualizerインスタンスを作成する
func NewVisualizer(outputDir, outputType string, stats *models.RepositoryStats) *Visualizer {
	return &Visualizer{
		outputDir:  outputDir,
		outputType: strings.ToLower(outputType),
		stats:      stats,
	}
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

	// 描画パラメータ
	const (
		canvasWidth  = 1200
		canvasHeight = 800
		marginTop    = 80
		marginLeft   = 50
		marginRight  = 50
		marginBottom = 50
		maxBarWidth  = canvasWidth - marginLeft - marginRight
		barHeight    = 20
		barGap       = 5
	)

	// 表示するファイル数を制限（最大で25ファイル）
	maxFiles := 25
	if len(files) > maxFiles {
		files = files[:maxFiles]
	}

	// SVGヘッダーを書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <text x="%d" y="40" font-size="24" font-family="Arial">%s リポジトリヒートマップ</text>
  <text x="%d" y="70" font-size="14" font-family="Arial">コミット数: %d, ファイル数: %d, 期間: %s 〜 %s</text>
  
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

  <!-- ヒートマップ -->
  <g transform="translate(%d, %d)">
`,
		canvasWidth, canvasHeight,
		canvasWidth/2, v.stats.RepositoryName,
		canvasWidth/2, v.stats.CommitCount, v.stats.FileCount,
		v.stats.FirstCommitAt.Format("2006/01/02"), v.stats.LastCommitAt.Format("2006/01/02"),
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

	// SVGの基本構造を書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="800" height="400" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <text x="10" y="30" font-size="16" font-family="Arial">ファイル: %s</text>
  <text x="10" y="50" font-size="14" font-family="Arial">変更回数: %d / 最終変更: %s</text>
  
  <!-- ヒートマップバー -->
  <rect x="10" y="70" width="780" height="30" fill="%s" />
  <text x="400" y="90" text-anchor="middle" font-size="14" fill="white">変更頻度レベル: %.2f</text>
  
  <!-- 著者情報 -->
  <text x="10" y="120" font-size="14" font-family="Arial">主な貢献者:</text>`,
		filePath, fileInfo.ChangeCount, fileInfo.LastModified.Format("2006/01/02"),
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

	// 行変更ヒートマップの描画（最大100行まで）
	if len(lineChanges) > 0 {
		fmt.Fprint(file, `
  
  <!-- 行変更ヒートマップ -->
  <text x="10" y="240" font-size="14" font-family="Arial">行変更ヒートマップ:</text>
  <g transform="translate(10, 250)">`)

		// 最大変更行数を計算
		maxCount := 0
		for _, lc := range lineChanges {
			if lc.Count > maxCount {
				maxCount = lc.Count
			}
		}

		// 表示する行数を制限
		lineLimit := 100
		if len(lineChanges) < lineLimit {
			lineLimit = len(lineChanges)
		}

		for i := 0; i < lineLimit; i++ {
			lc := lineChanges[i]
			lineHeatLevel := float64(lc.Count) / float64(maxCount)
			lineColor := GetHeatColor(lineHeatLevel)
			barWidth := 50 + int(lineHeatLevel*650) // 最小幅50、最大幅700

			fmt.Fprintf(file, `
    <rect x="0" y="%d" width="%d" height="4" fill="%s" />
    <text x="%d" y="%d" font-size="10" font-family="monospace">行%d: %d回</text>`,
				i*6, barWidth, lineColor, barWidth+5, i*6+3, lc.LineNum, lc.Count)
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
