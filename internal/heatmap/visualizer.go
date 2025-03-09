package heatmap

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/repositoryheatmap/internal/utils"
	"github.com/repositoryheatmap/internal/web_visualizer"
	"github.com/repositoryheatmap/pkg/models"
)

// Visualizer はヒートマップの可視化を担当する構造体
type Visualizer struct {
	outputDir      string
	outputType     string
	stats          *models.RepositoryStats
	repoPath       string // リポジトリのルートパス
	maxFilesToShow int    // 表示する最大ファイル数
}

// TreeNode はツリーマップのノードを表す構造体
type TreeNode struct {
	Name      string
	Path      string
	Size      float64 // ファイルの変更回数（ヒートレベル）
	HeatLevel float64 // 0.0-1.0の範囲のヒートレベル
	Children  []*TreeNode
	IsDir     bool
	Rect      Rectangle // 描画用の矩形領域
}

// Rectangle はツリーマップの矩形領域を表す構造体
type Rectangle struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// NewVisualizer は新しいVisualizerインスタンスを作成する
func NewVisualizer(outputDir, outputType string, stats *models.RepositoryStats) *Visualizer {
	return &Visualizer{
		outputDir:      outputDir,
		outputType:     strings.ToLower(outputType),
		stats:          stats,
		repoPath:       "",  // デフォルトは空文字
		maxFilesToShow: 100, // デフォルトは100ファイル
	}
}

// SetRepoPath はリポジトリパスを設定する
func (v *Visualizer) SetRepoPath(repoPath string) {
	v.repoPath = repoPath
}

// SetMaxFilesToShow は表示する最大ファイル数を設定する
func (v *Visualizer) SetMaxFilesToShow(maxFiles int) {
	if maxFiles < 1 {
		maxFiles = 1 // 最低1ファイルは表示する
	}
	v.maxFilesToShow = maxFiles
}

// Visualize はヒートマップデータを可視化する
func (v *Visualizer) Visualize() error {
	// 出力タイプのチェック
	if v.outputType != "svg" && v.outputType != "webp" && v.outputType != "html" {
		return fmt.Errorf("サポートされていない出力形式です: %s (svg, html または webp を使用してください)", v.outputType)
	}

	// 出力ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(v.outputDir, 0755); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗しました: %w", err)
	}

	// 出力形式に応じて処理を分岐
	switch v.outputType {
	case "svg":
		// リポジトリ全体のヒートマップを生成
		if err := v.generateRepositoryHeatmap(); err != nil {
			return fmt.Errorf("リポジトリヒートマップの生成に失敗しました: %w", err)
		}

		// 出力ファイルパスを作成
		repoOutputPath := filepath.Join(v.outputDir, fmt.Sprintf("%s-repository-heatmap.%s", v.stats.RepositoryName, v.outputType))
		fmt.Printf("リポジトリヒートマップをSVGに保存しました: %s\n", repoOutputPath)

		// 個別ファイルのヒートマップを生成
		if err := v.generateFileHeatmaps(); err != nil {
			return fmt.Errorf("ファイルヒートマップの生成に失敗しました: %w", err)
		}
		fmt.Printf("ファイルヒートマップを保存しました: %s\n", filepath.Join(v.outputDir, "file-heatmaps"))

	case "webp":
		// WebP形式の出力（将来の拡張用）
		return fmt.Errorf("WebP形式の出力はまだ実装されていません")

	case "html":
		// HTML形式の出力
		htmlOutputPath := filepath.Join(v.outputDir, fmt.Sprintf("%s-heatmap.html", v.stats.RepositoryName))
		if err := v.GenerateHTMLHeatmap(htmlOutputPath); err != nil {
			return fmt.Errorf("HTMLヒートマップの生成に失敗しました: %w", err)
		}
		fmt.Printf("リポジトリヒートマップをHTMLに保存しました: %s\n", htmlOutputPath)
	}

	return nil
}

// GenerateHTMLHeatmap はHTML形式のヒートマップを生成します
func (v *Visualizer) GenerateHTMLHeatmap(outputPath string) error {
	// HTMLテンプレートを読み込む
	htmlTemplate, err := web_visualizer.GetHTMLTemplate()
	if err != nil {
		return fmt.Errorf("HTMLテンプレートの読み込みに失敗しました: %w", err)
	}

	// HTMLファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("HTMLファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// テンプレートを実行
	data := struct {
		RepoName    string
		MaxFiles    int
		JSONData    *models.RepositoryStats
		FirstCommit string
		LastCommit  string
	}{
		RepoName:    v.stats.RepositoryName,
		MaxFiles:    v.maxFilesToShow,
		JSONData:    v.stats,
		FirstCommit: v.stats.FirstCommitAt.Format("2006/01/02"),
		LastCommit:  v.stats.LastCommitAt.Format("2006/01/02"),
	}

	if err := htmlTemplate.Execute(file, data); err != nil {
		return fmt.Errorf("HTMLテンプレートの実行に失敗しました: %w", err)
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

	// ツリーマップのルートノードを作成
	root := v.buildFileTree()

	// 描画パラメータ
	canvasWidth := 1200
	canvasHeight := 900
	if len(v.stats.Files) > 20 {
		// ファイル数が多い場合は高さを調整
		canvasHeight = 900 + (len(v.stats.Files)/20)*100
	}
	marginTop := 100
	marginRight := 50
	marginBottom := 50
	marginLeft := 50
	treeMapWidth := canvasWidth - marginLeft - marginRight
	treeMapHeight := canvasHeight - marginTop - marginBottom

	// ツリーマップレイアウトを計算
	rootRect := Rectangle{
		X:      float64(marginLeft),
		Y:      float64(marginTop),
		Width:  float64(treeMapWidth),
		Height: float64(treeMapHeight),
	}
	v.layoutTreeMap(root, rootRect)

	// SVGヘッダーを書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <text x="%d" y="40" font-size="24" font-family="Arial" text-anchor="middle">%s リポジトリヒートマップ</text>
  <text x="%d" y="70" font-size="14" font-family="Arial" text-anchor="middle">コミット数: %d, ファイル数: %d, 期間: %s 〜 %s</text>
  <text x="%d" y="90" font-size="12" font-family="Arial" text-anchor="middle">※ 変更頻度上位 %d ファイルを表示（全 %d ファイル中）</text>
  
  <!-- 凡例 -->
  <g transform="translate(50, 30)">
    <text x="0" y="0" font-size="12" font-family="Arial">変更頻度:</text>
    <rect x="80" y="-10" width="20" height="12" fill="#0000FF" />
    <text x="105" y="0" font-size="10" font-family="Arial">低</text>
    <rect x="130" y="-10" width="20" height="12" fill="#0066FF" />
    <rect x="180" y="-10" width="20" height="12" fill="#00FF00" />
    <rect x="230" y="-10" width="20" height="12" fill="#FFCC00" />
    <rect x="280" y="-10" width="20" height="12" fill="#FF0000" />
    <text x="305" y="0" font-size="10" font-family="Arial">高</text>
  </g>
`,
		canvasWidth, canvasHeight,
		canvasWidth/2, v.stats.RepositoryName,
		canvasWidth/2, v.stats.CommitCount, v.stats.FileCount,
		v.stats.FirstCommitAt.Format("2006/01/02"), v.stats.LastCommitAt.Format("2006/01/02"),
		canvasWidth/2, len(getSortedFilesByHeat(v.stats)[:v.maxFilesToShow]), len(v.stats.Files))

	// ツリーマップを描画
	v.renderTreeMap(file, root)

	// SVGフッターを書き込む
	fmt.Fprintf(file, `</svg>`)

	fmt.Printf("リポジトリヒートマップをSVGに保存しました: %s\n", outputPath)
	return nil
}

// buildFileTree はファイル情報からツリー構造を構築する
func (v *Visualizer) buildFileTree() *TreeNode {
	root := &TreeNode{
		Name:     "root",
		Path:     "",
		IsDir:    true,
		Children: []*TreeNode{},
	}

	// ファイルのソートと制限
	files := getSortedFilesByHeat(v.stats)
	maxFiles := v.maxFilesToShow
	if len(files) > maxFiles {
		files = files[:maxFiles]
	}

	// 各ファイルをツリーに追加
	for _, fileInfo := range files {
		path := fileInfo.FilePath
		pathParts := strings.Split(path, "/")

		// パスがバックスラッシュで区切られている場合（Windows）
		if len(pathParts) == 1 {
			pathParts = strings.Split(path, "\\")
		}

		currentNode := root

		// ディレクトリ階層を構築
		for i := 0; i < len(pathParts)-1; i++ {
			dirName := pathParts[i]
			if dirName == "" {
				continue
			}

			// 既存のディレクトリノードを検索
			found := false
			for _, child := range currentNode.Children {
				if child.Name == dirName && child.IsDir {
					currentNode = child
					found = true
					break
				}
			}

			// 見つからなければ新しいディレクトリノードを作成
			if !found {
				dirNode := &TreeNode{
					Name:     dirName,
					Path:     strings.Join(pathParts[:i+1], "/"),
					IsDir:    true,
					Children: []*TreeNode{},
				}
				currentNode.Children = append(currentNode.Children, dirNode)
				currentNode = dirNode
			}
		}

		// ファイルノードを追加
		fileName := pathParts[len(pathParts)-1]
		fileNode := &TreeNode{
			Name:      fileName,
			Path:      fileInfo.FilePath,
			Size:      float64(fileInfo.ChangeCount),
			HeatLevel: fileInfo.HeatLevel,
			IsDir:     false,
			Children:  []*TreeNode{},
		}
		currentNode.Children = append(currentNode.Children, fileNode)
	}

	// ディレクトリサイズを計算（子ノードの合計）
	calculateDirSize(root)

	return root
}

// calculateDirSize はディレクトリノードのサイズを計算する（再帰的に）
func calculateDirSize(node *TreeNode) float64 {
	if !node.IsDir {
		return node.Size
	}

	var total float64
	for _, child := range node.Children {
		total += calculateDirSize(child)
	}

	node.Size = total
	node.HeatLevel = 0 // ディレクトリのヒートレベルは子ノードの平均値を使用することも可能
	return total
}

// layoutTreeMap はツリーマップのレイアウトを計算する
func (v *Visualizer) layoutTreeMap(node *TreeNode, rect Rectangle) {
	node.Rect = rect

	if !node.IsDir || len(node.Children) == 0 {
		return
	}

	// 子ノードの合計サイズを計算
	var totalSize float64
	for _, child := range node.Children {
		totalSize += child.Size
	}

	// サイズが0の場合は処理しない
	if totalSize <= 0 {
		return
	}

	// スクエアリファインド（Squarified）アルゴリズムの簡易版
	// 横長の矩形なら縦に分割、縦長の矩形なら横に分割
	isHorizontal := rect.Width > rect.Height

	// 現在の位置
	x := rect.X
	y := rect.Y

	// 残りの幅と高さ
	remainingWidth := rect.Width
	remainingHeight := rect.Height

	for _, child := range node.Children {
		// 子ノードが占める割合
		ratio := child.Size / totalSize

		// 子ノードの矩形を計算
		var childRect Rectangle
		if isHorizontal {
			// 横方向に分割
			childRect = Rectangle{
				X:      x,
				Y:      y,
				Width:  remainingWidth * ratio,
				Height: rect.Height,
			}
			x += childRect.Width
		} else {
			// 縦方向に分割
			childRect = Rectangle{
				X:      x,
				Y:      y,
				Width:  rect.Width,
				Height: remainingHeight * ratio,
			}
			y += childRect.Height
		}

		// 再帰的にレイアウト
		v.layoutTreeMap(child, childRect)
	}
}

// renderTreeMap はツリーマップをSVGに描画する
func (v *Visualizer) renderTreeMap(file *os.File, node *TreeNode) {
	// ルートノードは描画しない
	if node.Name == "root" {
		// 子ノードを描画
		for _, child := range node.Children {
			v.renderTreeMap(file, child)
		}
		return
	}

	// ノードの矩形を描画
	rect := node.Rect

	// 最小サイズの確認（あまりに小さい矩形は描画しない）
	if rect.Width < 2 || rect.Height < 2 {
		return
	}

	// 色を決定
	var color string
	if node.IsDir {
		// ディレクトリは薄いグレー
		color = "#e0e0e0"
	} else {
		// ファイルはヒートレベルに応じた色
		color = GetHeatColor(node.HeatLevel)
	}

	// ファイルノードの場合は、リンクを追加
	if !node.IsDir {
		// ファイル名をサニタイズ
		safeFileName := utils.SanitizePath(node.Path)
		// リンク先のファイルパス（相対パス）
		linkPath := fmt.Sprintf("file-heatmaps/%s.%s", safeFileName, v.outputType)

		// リンクタグを開始
		fmt.Fprintf(file, `  <a href="%s" target="_blank">`, linkPath)
	}

	// 矩形を描画
	fmt.Fprintf(file, `  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="#ffffff" stroke-width="1">
    <title>%s (変更回数: %d)</title>
  </rect>`,
		rect.X, rect.Y, rect.Width, rect.Height, color, node.Path, int(node.Size))

	// テキストを表示（十分な大きさの場合のみ）
	if rect.Width > 40 && rect.Height > 14 {
		// 表示名を短縮
		displayName := node.Name
		if len(displayName) > 20 {
			displayName = displayName[:17] + "..."
		}

		fmt.Fprintf(file, `
  <text x="%.1f" y="%.1f" font-size="12" font-family="Arial" fill="#000000" pointer-events="none">%s</text>`,
			rect.X+4, rect.Y+15, displayName)
	}

	// ファイルノードの場合は、リンクタグを閉じる
	if !node.IsDir {
		fmt.Fprintf(file, `  </a>`)
	}

	// 子ノードを再帰的に描画
	for _, child := range node.Children {
		v.renderTreeMap(file, child)
	}
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
	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("SVGファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// ファイルの内容を読み込む
	fileContent, err := v.readFileContent(filePath)
	// エラーの場合は空の内容で続行
	if err != nil {
		fileContent = []string{"ファイルの内容を読み込めませんでした。"}
	}

	// 描画パラメータ
	const (
		lineHeight   = 20
		marginTop    = 100
		marginLeft   = 50
		marginRight  = 50
		marginBottom = 50
		maxWidth     = 1200
	)

	// コード行数に基づいてキャンバスの高さを決定
	contentHeight := len(fileContent) * lineHeight
	canvasHeight := contentHeight + marginTop + marginBottom
	if canvasHeight < 400 {
		canvasHeight = 400 // 最小高さ
	}

	// ヒートマップに表示するタイトル（ファイルパス）
	title := filePath
	if len(title) > 80 {
		title = "..." + title[len(title)-77:]
	}

	// SVGヘッダーを書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <a href="../%s-repository-heatmap.%s" title="リポジトリ全体のヒートマップに戻る">
    <text x="50" y="30" font-size="16" font-family="Arial" fill="#0066cc">← リポジトリ全体に戻る</text>
  </a>
  <text x="600" y="40" font-size="24" font-family="Arial" text-anchor="middle">ファイルヒートマップ</text>
  <text x="600" y="70" font-size="14" font-family="Arial" text-anchor="middle">%s</text>
  <text x="600" y="90" font-size="12" font-family="Arial" text-anchor="middle">変更回数: %d, 最終更新: %s</text>

  <!-- スクロール可能なコードビュー -->
  <g transform="translate(%d, %d)">
    <!-- 行番号とコード -->
`,
		maxWidth, canvasHeight,
		v.stats.RepositoryName, v.outputType,
		title, fileInfo.ChangeCount, fileInfo.LastModified.Format("2006/01/02"),
		marginLeft, marginTop)

	// 各行のコードを描画
	for i, line := range fileContent {
		// 行番号を計算
		lineNum := i + 1

		// この行のヒートレベル（デフォルトは0）
		heatLevel := 0
		if level, exists := fileInfo.LineChanges[lineNum]; exists {
			heatLevel = level
		}

		// 色を取得
		lineColor := "#ffffff" // デフォルトは白
		if heatLevel > 0 {
			// ヒートレベルを0.0-1.0の範囲に変換
			normalizedHeat := float64(heatLevel) / 10.0
			lineColor = GetHeatColor(normalizedHeat)
		}

		// テキストをエスケープ
		escapedText := escapeSpecialChars(line)

		// 行を描画
		fmt.Fprintf(file, `    <g>
      <rect x="0" y="%d" width="%d" height="%d" fill="%s" />
      <text x="-10" y="%d" text-anchor="end" font-size="12" font-family="monospace">%d</text>
      <text x="5" y="%d" font-size="12" font-family="monospace">%s</text>
    </g>
`,
			i*lineHeight, maxWidth-marginLeft-marginRight, lineHeight, lineColor,
			i*lineHeight+14, lineNum,
			i*lineHeight+14, escapedText)
	}

	// SVGフッターを書き込む
	fmt.Fprintf(file, `  </g>
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
