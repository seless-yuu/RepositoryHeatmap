package heatmap

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/repositoryheatmap/internal/utils"
	"github.com/repositoryheatmap/pkg/models"
)

// Visualizer はヒートマップの可視化を担当する構造体
type Visualizer struct {
	outputDir      string
	outputType     string
	stats          *models.RepositoryStats
	maxFilesToShow int
	inputJSONPath  string
}

// TreeNode はチエーマップのノードを表す構造体
type TreeNode struct {
	Name      string
	Path      string
	Size      float64 // ファイルの変更回数�E�ヒートレベル�E�E
	HeatLevel float64 // 0.0-1.0の篁E��のヒートレベル
	Children  []*TreeNode
	IsDir     bool
	Rect      Rectangle // 描画用の矩形領域
}

// Rectangle はチエーマップの矩形領域を表す構造体
type Rectangle struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// NewVisualizer は新しいVisualizerインスタンスを作成する
func NewVisualizer(outputDir, outputType string, stats *models.RepositoryStats, maxFilesToShow int, inputJSONPath string) *Visualizer {
	return &Visualizer{
		outputDir:      outputDir,
		outputType:     outputType,
		stats:          stats,
		maxFilesToShow: maxFilesToShow,
		inputJSONPath:  inputJSONPath,
	}
}

// SetRepoPath はリポジトリパスを設定すめE
func (v *Visualizer) SetRepoPath(repoPath string) {
	// こ関数は後方互換性のために残してぁE��すが、実際には何もしません
}

// SetMaxFilesToShow は表示する最大ファイル数を設定すめE
func (v *Visualizer) SetMaxFilesToShow(maxFiles int) {
	if maxFiles < 1 {
		maxFiles = 1 // 最佁Eファイルは表示する
	}
	v.maxFilesToShow = maxFiles
}

// Visualize はヒートマップの生成すめE
func (v *Visualizer) Visualize() error {
	// 出力ディレクトリが存在するか碁E
	if err := utils.EnsureDirectoryExists(v.outputDir); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗しました: %w", err)
	}

	// ヒートマップの種類に応じて生成処理を行います
	if v.outputType == "svg" || v.outputType == "webp" {
		// リポジトリ全体のヒートマップを生成
		if err := v.generateRepositoryHeatmap(); err != nil {
			return fmt.Errorf("リポジトリヒートマップの生成に失敗しました: %w", err)
		}

		// 個別のファイルヒートマップを生成
		if err := v.generateFileHeatmaps(); err != nil {
			return fmt.Errorf("ファイルヒートマップの生成に失敗しました: %w", err)
		}
	} else {
		return fmt.Errorf("未サポートの出力形式でぁE %s", v.outputType)
	}

	fmt.Printf("可視化が完了しました\n")
	fmt.Printf("結果は %s チエーマップディレクトリに保存されました\n", v.outputDir)

	return nil
}

// GetHeatColor は熱レベルに応じた色を返す
func GetHeatColor(heatLevel float64) string {
	// 目に優しい暖色系のグラチエションE低温から高温へEE
	if heatLevel < 0.2 {
		return "#FFF5E6" // 非常に薁E黁EE低！E
	} else if heatLevel < 0.4 {
		return "#FFE0B2" // 薁E橙色EやめE！E
	} else if heatLevel < 0.6 {
		return "#FFAB66" // 中程度の橙色�E�中�E�E
	} else if heatLevel < 0.8 {
		return "#FF8533" // めE��濁E��橙色�E�やめE��！E
	} else {
		return "#E65C00" // 濁E��橙色�E�高！E
	}
}

// generateRepositoryHeatmap はリポジトリ全体のヒートマップを生成すめE
func (v *Visualizer) generateRepositoryHeatmap() error {
	// 出力ファイルパスを作成
	outputPath := filepath.Join(v.outputDir, fmt.Sprintf("%s-repository-heatmap.%s", v.stats.RepositoryName, v.outputType))

	// SVG形式の場合吁E
	if v.outputType == "svg" {
		return v.generateSVGRepositoryHeatmap(outputPath)
	}

	// WebP形式の場合吁E
	return v.generateWebPRepositoryHeatmap(outputPath)
}

// generateSVGRepositoryHeatmap はSVG形式のリポジトリヒートマップを生成すめE
func (v *Visualizer) generateSVGRepositoryHeatmap(outputPath string) error {
	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("SVGファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// チエーマップのルートノードを作成
	root := v.buildFileTree()

	// 描画パラメータ
	canvasWidth := 1200
	canvasHeight := 900
	if len(v.stats.Files) > 20 {
		// ファイル数が多い場合E高さを調整
		canvasHeight = 900 + (len(v.stats.Files)/20)*100
	}
	marginTop := 100
	marginRight := 50
	marginBottom := 50
	marginLeft := 50
	treeMapWidth := canvasWidth - marginLeft - marginRight
	treeMapHeight := canvasHeight - marginTop - marginBottom

	// チエーマップレイアウトを計箁E
	rootRect := Rectangle{
		X:      float64(marginLeft),
		Y:      float64(marginTop),
		Width:  float64(treeMapWidth),
		Height: float64(treeMapHeight),
	}
	v.layoutTreeMap(root, rootRect)

	// リポジトリ全体に戻るパスを計箁E
	// ファイル個別SVGは常にfile-heatmapsチエーマップディレクトリ冁E生され、E
	// リポジトリ全体SVGはそエ親チエーマップディレクトリにあるので、常に "../" で戻れる

	// SVGヘッダーを書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <a href="../../%s-repository-heatmap.%s" title="リポジトリ全体のヒートマップに戻る">
    <text x="50" y="30" font-size="16" font-family="Segoe UI, Helvetica, Arial" fill="#333333">← リポジトリ全体に戻る</text>
  </a>
  <text x="%d" y="40" font-size="24" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle">%s リポジトリヒートマップ</text>
  <text x="%d" y="70" font-size="14" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle">コミット数: %d, ファイル数: %d, 期間: %s 〜 %s</text>
  <text x="%d" y="90" font-size="12" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle">※ 変更頻度上位 %d ファイルを表示（全 %d ファイル中）</text>
  
  <!-- 凡例 -->
  <g transform="translate(50, 30)">
    <text x="0" y="0" font-size="12" font-family="Segoe UI, Helvetica, Arial" fill="#333333">変更頻度:</text>
    <text x="105" y="0" font-size="10" font-family="Segoe UI, Helvetica, Arial" fill="#333333">低</text>
    <rect x="130" y="-10" width="20" height="12" fill="#FFE0B2" />
    <rect x="180" y="-10" width="20" height="12" fill="#FFAB66" />
    <rect x="230" y="-10" width="20" height="12" fill="#FF8533" />
    <rect x="280" y="-10" width="20" height="12" fill="#E65C00" />
    <text x="305" y="0" font-size="10" font-family="Segoe UI, Helvetica, Arial" fill="#333333">高</text>
  </g>
`,
		canvasWidth, canvasHeight,
		v.stats.RepositoryName, v.outputType,
		canvasWidth/2, v.stats.RepositoryName,
		canvasWidth/2, v.stats.CommitCount, len(v.stats.Files),
		v.stats.FirstCommitAt.Format("2006/01/02"), v.stats.LastCommitAt.Format("2006/01/02"),
		canvasWidth/2, len(getSortedFilesByHeat(v.stats)[:v.maxFilesToShow]), len(v.stats.Files))

	// チエーマップを描画
	v.renderTreeMap(file, root)

	// SVGフッターを書き込む
	fmt.Fprintf(file, `</svg>`)

	fmt.Printf("リポジトリヒートマップをSVGに保存しました: %s\n", outputPath)
	return nil
}

// buildFileTree はファイル情報からチエーツリー構造を構築すめE
func (v *Visualizer) buildFileTree() *TreeNode {
	root := &TreeNode{
		Name:     "root",
		Path:     "",
		IsDir:    true,
		Children: []*TreeNode{},
	}

	// ファイルのソートと制陁E
	files := getSortedFilesByHeat(v.stats)
	maxFiles := v.maxFilesToShow
	if len(files) > maxFiles {
		files = files[:maxFiles]
	}

	// 吁Eァイルをツリーに追加
	for _, fileInfo := range files {
		path := fileInfo.FilePath
		pathParts := strings.Split(path, "/")

		// パスがバチエスラチゥで区刁EれてぁE場合！EindowsEE
		if len(pathParts) == 1 {
			pathParts = strings.Split(path, "\\")
		}

		currentNode := root

		// チエーマップ階層を構篁E
		for i := 0; i < len(pathParts)-1; i++ {
			dirName := pathParts[i]
			if dirName == "" {
				continue
			}

			// 既存チエーマップノードを検索
			found := false
			for _, child := range currentNode.Children {
				if child.Name == dirName && child.IsDir {
					currentNode = child
					found = true
					break
				}
			}

			// 見つからなければ新しいチエーマップノードを作成
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

	// チエーマップサイズを計箁E
	calculateDirSize(root)

	return root
}

// calculateDirSize はチエーマップノードのサイズを計算する（再帰皁EEE
func calculateDirSize(node *TreeNode) float64 {
	if !node.IsDir {
		return node.Size
	}

	var total float64
	for _, child := range node.Children {
		total += calculateDirSize(child)
	}

	node.Size = total
	node.HeatLevel = 0 // チエーマップのヒートレベルは子ノード平均を使用することも可能
	return total
}

// layoutTreeMap はチエーマップレイアウトを計箁E
func (v *Visualizer) layoutTreeMap(node *TreeNode, rect Rectangle) {
	node.Rect = rect

	if !node.IsDir || len(node.Children) == 0 {
		return
	}

	// 子ノード合計サイズを計箁E
	var totalSize float64
	for _, child := range node.Children {
		totalSize += child.Size
	}

	// サイズぁEの場合E処琁EなぁE
	if totalSize <= 0 {
		return
	}

	// スクエアリファインド！EquarifiedEアルゴリズムの簡易版
	// 横長の矩形なら縦に刁E、縦長の矩形なら横に刁E
	isHorizontal := rect.Width > rect.Height

	// 現在の位置
	x := rect.X
	y := rect.Y

	// 残りの幁E高さ
	remainingWidth := rect.Width
	remainingHeight := rect.Height

	for _, child := range node.Children {
		// 子ノードが占める割吁E
		ratio := child.Size / totalSize

		// 子ノードの矩形を計箁E
		var childRect Rectangle
		if isHorizontal {
			// 横方向に刁E
			childRect = Rectangle{
				X:      x,
				Y:      y,
				Width:  remainingWidth * ratio,
				Height: rect.Height,
			}
			x += childRect.Width
		} else {
			// 縦方向に刁E
			childRect = Rectangle{
				X:      x,
				Y:      y,
				Width:  rect.Width,
				Height: remainingHeight * ratio,
			}
			y += childRect.Height
		}

		// 再帰皁EレイアウチE
		v.layoutTreeMap(child, childRect)
	}
}

// renderTreeMap はチエーマップをSVGに描画する
func (v *Visualizer) renderTreeMap(file *os.File, node *TreeNode) {
	// ルートノードを描画しなぁE
	if node.Name == "root" {
		// 子ノードを描画
		for _, child := range node.Children {
			v.renderTreeMap(file, child)
		}
		return
	}

	// ノードの矩形を描画
	rect := node.Rect

	// 最小サイズの確認（あまりに小さぁE形は描画しなぁEE
	if rect.Width < 2 || rect.Height < 2 {
		return
	}

	// 色を決宁E
	var color string
	if node.IsDir {
		// チエーマップは薁Eグレー
		color = "#e0e0e0"
	} else {
		// ファイルはヒートレベルに応じた色
		color = GetHeatColor(node.HeatLevel)
	}

	// ファイルノードの場合E、リンクを追加
	if !node.IsDir {
		// ファイル名をサニタイズ
		safeFileName := utils.SanitizePath(node.Path)
		// リンク先はファイルパスE相対パスEE
		linkPath := fmt.Sprintf("file-heatmaps/%s.%s", safeFileName, v.outputType)

		// リンクタグを開姁E
		fmt.Fprintf(file, `  <a href="%s" target="_blank">`, linkPath)
	}

	// 矩形を描画
	fmt.Fprintf(file, `  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" stroke="#ffffff" stroke-width="1">
    <title>%s (変更回数: %d)</title>
  </rect>`,
		rect.X, rect.Y, rect.Width, rect.Height, color, node.Path, int(node.Size))

	// チエードを表示E十刁E大きさの場合EみEE
	if rect.Width > 40 && rect.Height > 14 {
		// 表示名を短縮
		displayName := node.Name
		if len(displayName) > 20 {
			displayName = displayName[:17] + "..."
		}

		fmt.Fprintf(file, `
  <text x="%.1f" y="%.1f" font-size="14" font-family="Segoe UI, Helvetica, Arial" fill="#000000" pointer-events="none">%s</text>`,
			rect.X+4, rect.Y+16, displayName)
	}

	// ファイルノードの場合E、リンクタグを閉じる
	if !node.IsDir {
		fmt.Fprintf(file, `  </a>`)
	}

	// 子ノードを再帰皁E描画
	for _, child := range node.Children {
		v.renderTreeMap(file, child)
	}
}

// generateWebPRepositoryHeatmap はWebP形式のリポジトリヒートマップを生成すめE
func (v *Visualizer) generateWebPRepositoryHeatmap(outputPath string) error {
	// TODO: 実際のWebP生成処理琁EE実裁E
	// Go言語でWebP形式を生成するには、別のライブラリが忁Eになる可能性がありまぁE

	// ダミEのファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("WebPファイルの作成に失敗しました: %w", err)
	}
	file.Close()

	fmt.Printf("リポジトリヒートマップをWebPに保存しました: %s\n", outputPath)
	return nil
}

// generateFileHeatmaps は吁Eァイルのヒートマップを生成すめE
func (v *Visualizer) generateFileHeatmaps() error {
	// ファイルヒートマップの出力ディレクトリを作成
	fileHeatmapDir := filepath.Join(v.outputDir, "file-heatmaps")
	if err := utils.EnsureDirectoryExists(fileHeatmapDir); err != nil {
		return fmt.Errorf("ファイルヒートマップディレクトリの作成に失敗しました: %w", err)
	}

	// 吁Eァイルのヒートマップを生成
	for filePath, fileInfo := range v.stats.Files {
		// ファイルパスを安Eな形式に変換
		safeFileName := utils.SanitizePath(filePath)

		// HTMLのリンクではスラチゥを使用する忁Eがあるため、E
		// 格納する実際のファイルシスチEパスに変換
		fsPath := filepath.FromSlash(safeFileName)

		// 出力ファイルパスを作成
		outputPath := filepath.Join(fileHeatmapDir, fsPath+"."+v.outputType)

		// ファイルの階層の深さを計算して適切な相対パスを生成
		// 例: "foo/bar/baz.go" -> "../../../gin-repository-heatmap.svg"
		depth := len(strings.Split(fsPath, string(filepath.Separator)))
		relativePathPrefix := strings.Repeat("../", depth+1) // +1 for file-heatmaps directory
		repoSVGRelPath := fmt.Sprintf("%s%s-repository-heatmap.%s", relativePathPrefix, v.stats.RepositoryName, v.outputType)

		// 親チエーマップディレクトリが存在することを碁E
		parentDir := filepath.Dir(outputPath)
		if err := utils.EnsureDirectoryExists(parentDir); err != nil {
			return fmt.Errorf("チエーマップディレクトリの作成に失敗しました: %s: %w", parentDir, err)
		}

		// ファイルの変更頻度に基づく色を取征E
		color := GetHeatColor(fileInfo.HeatLevel)

		// SVG形式の場合吁E
		var err error
		if v.outputType == "svg" {
			err = v.generateSVGFileHeatmap(outputPath, filePath, fileInfo, color, repoSVGRelPath)
		} else {
			// WebP形式の場合吁E
			err = v.generateWebPFileHeatmap(outputPath, filePath, fileInfo, color)
		}

		// エラーが発生した場合E処琁E中断
		if err != nil {
			return fmt.Errorf("ファイル '%s' のヒートマップ生成に失敗しました: %w", filePath, err)
		}
	}

	fmt.Printf("ファイルヒートマップを保存しました: %s\n", fileHeatmapDir)
	return nil
}

// readFileContent はファイルの冁Eを読み込む
func (v *Visualizer) readFileContent(filePath string) ([]string, error) {
	var attemptedPaths []string

	// 1. JSONファイルの場所を基準にした相対パス
	if v.inputJSONPath != "" {
		jsonDir := filepath.Dir(v.inputJSONPath)
		repoName := v.stats.RepositoryName

		// パターン1: JSONファイルと同じチエーマップディレクトリ
		fullPath := filepath.Join(jsonDir, filePath)
		attemptedPaths = append(attemptedPaths, fullPath)
		if _, err := os.Stat(fullPath); err == nil {
			return v.readFile(fullPath)
		}

		// パターン2: JSONファイルの親チエーマップディレクトリのtest-reposチエーマップディレクトリ + リポジトリ吁E
		testRepoPath := filepath.Join(jsonDir, "..", "test-repos", repoName)
		fullPath = filepath.Join(testRepoPath, filePath)
		attemptedPaths = append(attemptedPaths, fullPath)
		if _, err := os.Stat(fullPath); err == nil {
			return v.readFile(fullPath)
		}

		// パターン2.1: test-reposチエーマップディレクトリ直丁E(repoNameなぁE
		testRepoPath = filepath.Join(jsonDir, "..", "test-repos")
		fullPath = filepath.Join(testRepoPath, filePath)
		attemptedPaths = append(attemptedPaths, fullPath)
		if _, err := os.Stat(fullPath); err == nil {
			return v.readFile(fullPath)
		}

		// パターン3: カレントディレクトリのtest-reposチエーマップディレクトリ + リポジトリ吁E
		workingDir, err := os.Getwd()
		if err == nil {
			testPath := filepath.Join(workingDir, "test-repos", repoName, filePath)
			attemptedPaths = append(attemptedPaths, testPath)
			if _, err := os.Stat(testPath); err == nil {
				return v.readFile(testPath)
			}

			// パターン3.1: カレントディレクトリのtest-reposチエーマップディレクトリ直丁E(repoNameなぁE
			testPath = filepath.Join(workingDir, "test-repos", filePath)
			attemptedPaths = append(attemptedPaths, testPath)
			if _, err := os.Stat(testPath); err == nil {
				return v.readFile(testPath)
			}
		}
	}

	// 2. 相対パスをそのまま試す（カレントディレクトリからの相対パスEE
	workingDir, err := os.Getwd()
	if err == nil {
		fullPath := filepath.Join(workingDir, filePath)
		attemptedPaths = append(attemptedPaths, fullPath)
		if _, err := os.Stat(fullPath); err == nil {
			return v.readFile(fullPath)
		}
	}

	// 3. 絶対パスとして試ぁE
	attemptedPaths = append(attemptedPaths, filePath)
	if _, err := os.Stat(filePath); err == nil {
		return v.readFile(filePath)
	}

	// すべての試行が失敗した場吁E
	errMsg := fmt.Sprintf("ファイルの冁Eを読み込めませんでした: %s\n以下Eパスを試行しました:\n", filePath)
	for _, path := range attemptedPaths {
		errMsg += fmt.Sprintf("- %s\n", path)
	}

	return nil, errors.New(errMsg)
}

func (v *Visualizer) readFile(fullPath string) ([]string, error) {
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// escapeSpecialChars はSVG冁E特殊文字をエスケープすめE
func escapeSpecialChars(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&apos;")
	return text
}

// generateSVGFileHeatmap はSVG形式のファイルヒートマップを生成すめE
func (v *Visualizer) generateSVGFileHeatmap(outputPath, filePath string, fileInfo models.FileChangeInfo, color string, repoSVGRelPath string) error {
	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("SVGファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// ファイル冁Eの取征E
	var fileContent []string

	// JSONに行E容が保存されてぁE場合Eそれを使用
	if len(fileInfo.LineContents) > 0 {
		// 最大の行番号を取征E
		maxLine := 0
		for lineNum := range fileInfo.LineContents {
			if lineNum > maxLine {
				maxLine = lineNum
			}
		}

		// LineContentsから行E容を取征E
		fileContent = make([]string, maxLine)
		for i := range fileContent {
			lineNum := i + 1
			if content, exists := fileInfo.LineContents[lineNum]; exists {
				fileContent[i] = content
			}
		}

		fmt.Printf("JSONから '%s' の冁Eを読み込みました (%d衁E\n", filePath, maxLine)
	} else {
		// JSONに行E容がなぁE合E、リポジトリからファイル冁Eを読み込むE後方互換性EE
		var readErr error
		fileContent, readErr = v.readFileContent(filePath)
		if readErr != nil {
			// エラーの場合は空の内容で続行
			fileContent = []string{"ファイルの内容を読み込めませんでした"}
			fmt.Printf("警告: '%s' の内容を読み込めませんでした: %v\n", filePath, readErr)
		}
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

	// コード行数に基づぁE��キャンバスの高さを決宁E
	contentHeight := len(fileContent) * lineHeight
	canvasHeight := contentHeight + marginTop + marginBottom
	if canvasHeight < 400 {
		canvasHeight = 400 // 最小高さ
	}

	// ヒートマップに表示するタイトルEファイルパスEE
	title := filePath
	if len(title) > 80 {
		title = "..." + title[len(title)-77:]
	}

	// リポジトリ全体に戻るパスを計箁E
	// ファイル個別SVGは常にfile-heatmapsチエーマップディレクトリ冁E生され、E
	// リポジトリ全体SVGはそエ親チエーマップディレクトリにあるので、常に "../" で戻れる

	// SVGヘッダーを書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <a href="%s" title="リポジトリ全体のヒートマップに戻る">
    <text x="50" y="30" font-size="16" font-family="Segoe UI, Helvetica, Arial" fill="#333333">← リポジトリ全体に戻る</text>
  </a>
  <text x="600" y="40" font-size="24" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle" fill="#333333">ファイルヒートマップ</text>
  <text x="600" y="70" font-size="14" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle" fill="#333333">%s</text>
  <text x="600" y="90" font-size="12" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle" fill="#666666">変更回数: %d, 最終更新: %s</text>

  <!-- スクロール可能なコードビュー -->
  <g transform="translate(%d, %d)">
    <!-- 行番号とコード-->
`,
		maxWidth, canvasHeight,
		repoSVGRelPath,
		title, fileInfo.ChangeCount, fileInfo.LastModified.Format("2006/01/02"),
		marginLeft, marginTop)

	// 吁Eコードを描画
	for i, line := range fileContent {
		// 行番号を計箁E
		lineNum := i + 1

		// こE行のヒートレベルEデフォルトE0EE
		heatLevel := 0
		if level, exists := fileInfo.LineChanges[lineNum]; exists {
			heatLevel = level
		}

		// 色を取征E
		lineColor := "#ffffff" // チォルトE白
		if heatLevel > 0 {
			// ヒートレベルめE.0-1.0の篁Eに変換
			normalizedHeat := float64(heatLevel) / 10.0
			lineColor = GetHeatColor(normalizedHeat)
		}

		// チエードをエスケーチE
		escapedText := escapeSpecialChars(line)

		// 行を描画
		fmt.Fprintf(file, `    <g>
      <rect x="0" y="%d" width="%d" height="%d" fill="%s" />
      <text x="-10" y="%d" text-anchor="end" font-size="12" font-family="Consolas, Menlo, monospace" fill="#666666">%d</text>
      <text x="5" y="%d" font-size="12" font-family="Consolas, Menlo, monospace" fill="#333333">%s</text>
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

// generateWebPFileHeatmap はWebP形式のファイルヒートマップを生成すめE
func (v *Visualizer) generateWebPFileHeatmap(outputPath, filePath string, fileInfo models.FileChangeInfo, color string) error {
	// 忁Eに応じてチエーマップディレクトリを作成
	if err := utils.EnsureDirectoryExists(filepath.Dir(outputPath)); err != nil {
		return fmt.Errorf("チエーマップディレクトリの作成に失敗しました: %s: %w", filepath.Dir(outputPath), err)
	}

	// TODO: 実際のWebP生成処理琁EE実裁E
	// Go言語でWebP形式を生成するには、別のライブラリが忁Eになる可能性がありまぁE

	// ダミEのファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("WebPファイルの作成に失敗しました: %w", err)
	}
	file.Close()

	return nil
}
