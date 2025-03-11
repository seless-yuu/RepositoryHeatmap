package heatmap

import (
	"bufio"
	"fmt"
	"html"
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

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// generateRepositoryHeatmap はリポジトリ全体のヒートマップを生成する
func (v *Visualizer) generateRepositoryHeatmap() error {
	// リポジトリ名の検証
	if v.stats.RepositoryName == "" {
		return fmt.Errorf("リポジトリ名が設定されていません")
	}

	// 出力ディレクトリのチェックと作成
	if err := utils.EnsureDirectoryExists(v.outputDir); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗しました: %w", err)
	}

	// 出力ファイルパスを決定
	outputFileName := fmt.Sprintf("%s-repository-heatmap.%s", v.stats.RepositoryName, v.outputType)
	outputPath := filepath.Join(v.outputDir, outputFileName)

	// ファイルヒートマップ用のディレクトリを作成
	fileHeatmapDir := filepath.Join(v.outputDir, "file-heatmaps")
	if err := utils.EnsureDirectoryExists(fileHeatmapDir); err != nil {
		return fmt.Errorf("ファイルヒートマップディレクトリの作成に失敗しました: %w", err)
	}

	// 出力タイプに基づいてヒートマップを生成
	var err error
	if v.outputType == "svg" {
		err = v.generateSVGRepositoryHeatmap(outputPath)
	} else {
		err = v.generateWebPRepositoryHeatmap(outputPath)
	}

	if err != nil {
		return err
	}

	// ファイルごとのヒートマップを生成
	if err := v.generateFileHeatmaps(); err != nil {
		return err
	}

	return nil
}

// generateSVGRepositoryHeatmap はSVG形式のリポジトリヒートマップを生成すめE
func (v *Visualizer) generateSVGRepositoryHeatmap(outputPath string) error {
	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("SVGファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

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

	// ファイルを変更頻度でソートし、表示数を制限する
	files := getSortedFilesByHeat(v.stats)
	if v.maxFilesToShow > 0 {
		if v.maxFilesToShow < len(files) {
			files = files[:v.maxFilesToShow]
		}
	} else {
		v.maxFilesToShow = len(files) // maxFilesToShowが0の場合は全ファイルを表示
	}

	// ツリー構造を構築
	root := v.buildFileTreeFromFiles(files)

	// チエーマップレイアウトを計箁E
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
  <title>%s Repository Heatmap (%s)</title>
  <text x="%d" y="30" text-anchor="middle" font-size="20" font-family="Segoe UI, Helvetica, Arial" fill="#333333">%s</text>
  <text x="%d" y="50" text-anchor="middle" font-size="12" font-family="Segoe UI, Helvetica, Arial" fill="#666666">コミット数: %d, ファイル数: %d</text>
  <text x="%d" y="65" text-anchor="middle" font-size="12" font-family="Segoe UI, Helvetica, Arial" fill="#666666">期間: %s 〜 %s</text>
  <text x="%d" y="80" text-anchor="middle" font-size="12" font-family="Segoe UI, Helvetica, Arial" fill="#666666">表示: %d / %d ファイル</text>
  <g transform="translate(10, 100)">
    <text x="10" y="0" font-size="10" font-family="Segoe UI, Helvetica, Arial" fill="#333333">低</text>
    <rect x="30" y="-10" width="20" height="12" fill="#FFFFFF" stroke="#CCCCCC" />
    <rect x="80" y="-10" width="20" height="12" fill="#FFE5CC" />
    <rect x="130" y="-10" width="20" height="12" fill="#FFD199" />
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
		canvasWidth/2, v.stats.FirstCommitAt.Format("2006/01/02"), v.stats.LastCommitAt.Format("2006/01/02"),
		canvasWidth/2, len(files), len(v.stats.Files))

	// チエーマップを描画
	v.renderTreeMap(file, root)

	// SVGフッターを書き込む
	fmt.Fprintf(file, `</svg>`)

	fmt.Printf("リポジトリヒートマップをSVGに保存しました: %s\n", outputPath)
	return nil
}

// buildFileTreeFromFiles はファイル情報からチエーツリー構造を構築すめE
func (v *Visualizer) buildFileTreeFromFiles(files []models.FileChangeInfo) *TreeNode {
	root := &TreeNode{
		Name:     "root",
		Path:     "",
		IsDir:    true,
		Children: []*TreeNode{},
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

// renderTreeMap はツリーマップを描画する
func (v *Visualizer) renderTreeMap(file *os.File, node *TreeNode) {
	// ボーダーと背景色を設定
	border := "#CCCCCC"
	bgColor := "#FFFFFF"

	// ヒートレベルに基づく色をディレクトリにも割り当てる
	if node.IsDir {
		bgColor = "#F8F8F8" // ディレクトリはわずかにグレー
	} else {
		// ヒートレベルに基づいて色を取得
		bgColor = GetHeatColor(node.HeatLevel)
	}

	// ノードの矩形を描画
	fmt.Fprintf(file, `  <rect x="%f" y="%f" width="%f" height="%f" fill="%s" stroke="%s">
    <title>%s (変更: %d)</title>
  </rect>`,
		node.Rect.X, node.Rect.Y, node.Rect.Width, node.Rect.Height,
		bgColor, border, node.Path, int(node.Size))

	// ファイルノードの場合、リンクを追加
	if !node.IsDir {
		// ファイル名をサニタイズ
		safeFileName := utils.SanitizePath(node.Path)

		// デバッグ情報を表示
		fmt.Printf("リンク生成: %s -> %s\n", node.Path, safeFileName)

		// リンク先はfile-heatmapsディレクトリ内のファイル
		// 常に相対パスで指定（リポジトリヒートマップSVGからの相対パス）
		linkPath := fmt.Sprintf("./file-heatmaps/%s.%s", safeFileName, v.outputType)

		// リンクタグを開始（エスケープ処理を追加）
		fmt.Fprintf(file, `  <a href="%s" target="_blank">`, html.EscapeString(linkPath))
	}

	// ノード名が表示できる十分なスペースがある場合のみ表示（幅 > 40px および 高さ > 14px）
	if node.Rect.Width > 40 && node.Rect.Height > 14 {
		// ファイル/ディレクトリ名を表示
		name := node.Name
		// 名前が長い場合は省略（表示幅に合わせる）
		if float64(len(name)*7) > node.Rect.Width {
			maxChars := int(node.Rect.Width / 7)
			if maxChars > 3 {
				name = name[:maxChars-3] + "..."
			} else if maxChars > 0 {
				name = name[:maxChars]
			} else {
				name = ""
			}
		}

		if name != "" {
			// テキストは白または黒（背景色に基づいて）
			textColor := "#000000" // デフォルトは黒
			if node.HeatLevel > 0.6 {
				textColor = "#FFFFFF" // 暗い背景色の場合は白
			}

			fmt.Fprintf(file, `  <text x="%f" y="%f" font-size="12" font-family="Segoe UI, Helvetica, Arial" fill="%s">%s</text>`,
				node.Rect.X+5, node.Rect.Y+15, textColor, name)
		}
	}

	// ファイルノードの場合、リンクタグを閉じる
	if !node.IsDir {
		fmt.Fprintf(file, `  </a>`)
	}

	// 子ノードを再帰的に描画
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

// generateFileHeatmaps はファイルごとのヒートマップを生成する
func (v *Visualizer) generateFileHeatmaps() error {
	// 出力ディレクトリを確認
	fileHeatmapDir := filepath.Join(v.outputDir, "file-heatmaps")
	if err := utils.EnsureDirectoryExists(fileHeatmapDir); err != nil {
		return fmt.Errorf("ファイルヒートマップディレクトリの作成に失敗しました: %w", err)
	}

	// ファイルを変更頻度でソート
	files := getSortedFilesByHeat(v.stats)

	// maxFilesToShowの値を確認
	maxFiles := v.maxFilesToShow
	if maxFiles <= 0 || maxFiles > len(files) {
		maxFiles = len(files)
	}

	// 選択されたファイルのみヒートマップを生成
	selectedFiles := files[:maxFiles]

	// 選択されたファイルのヒートマップを生成
	for _, fileInfo := range selectedFiles {
		filePath := fileInfo.FilePath
		// ファイルパスを安全な形式に変換
		safeFileName := utils.SanitizePath(filePath)

		// 出力ファイルパスを作成
		outputPath := filepath.Join(fileHeatmapDir, safeFileName+"."+v.outputType)

		// 親ディレクトリが存在することを確認
		parentDir := filepath.Dir(outputPath)
		if err := utils.EnsureDirectoryExists(parentDir); err != nil {
			return fmt.Errorf("ディレクトリの作成に失敗しました: %s: %w", parentDir, err)
		}

		// ファイルの変更頻度に基づく色を取得
		color := GetHeatColor(fileInfo.HeatLevel)

		// SVG形式の場合
		var err error
		if v.outputType == "svg" {
			err = v.generateSVGFileHeatmap(outputPath, filePath, fileInfo, color)
		} else {
			// WebP形式の場合
			err = v.generateWebPFileHeatmap(outputPath, filePath, fileInfo, color)
		}

		// エラーが発生した場合は処理を中断
		if err != nil {
			return fmt.Errorf("ファイル '%s' のヒートマップ生成に失敗しました: %w", filePath, err)
		}
	}

	return nil
}

// readFileContent はファイルの内容を読み込む
func (v *Visualizer) readFileContent(filePath string) ([]string, error) {
	// 1. まずJSONデータに保存されたLineContentsを確認
	fileInfo, exists := v.stats.Files[filePath]
	if exists && len(fileInfo.LineContents) > 0 {
		// すでに読み込んだ内容があれば、それを使用
		lines := make([]string, len(fileInfo.LineContents))
		for lineNum, content := range fileInfo.LineContents {
			lines[lineNum-1] = content // LineContentsは1始まりのインデックス
		}
		return lines, nil
	}

	// 2. 分析対象のリポジトリパスが設定されている場合
	if v.stats.RepositoryPath != "" {
		fullPath := filepath.Join(v.stats.RepositoryPath, filePath)
		// パスが実際に存在するか確認
		if _, err := os.Stat(fullPath); err == nil {
			return v.readFile(fullPath)
		}
		// ファイルが見つからない場合はエラーメッセージを返す
		return []string{fmt.Sprintf("// ファイル %s が見つかりません（%s）", filePath, fullPath)}, nil
	}

	// 3. リポジトリパスが設定されていない場合
	return []string{"// リポジトリパスが設定されていないため、ファイル内容を読み込めません"}, nil
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

// generateSVGFileHeatmap はSVG形式のファイルヒートマップを生成する
func (v *Visualizer) generateSVGFileHeatmap(outputPath, filePath string, fileInfo models.FileChangeInfo, color string) error {
	// デバッグ情報
	fmt.Printf("SVGファイルヒートマップを生成: %s -> %s\n", filePath, outputPath)

	// 出力ディレクトリの作成を確認
	outputDir := filepath.Dir(outputPath)
	if err := utils.EnsureDirectoryExists(outputDir); err != nil {
		return fmt.Errorf("出力ディレクトリの作成に失敗しました: %s: %w", outputDir, err)
	}

	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("SVGファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// ファイル内容の取得
	var fileContent []string

	// JSONに行内容が保存されている場合はそれを使用
	if len(fileInfo.LineContents) > 0 {
		// 最大の行番号を取得
		maxLine := 0
		for lineNum := range fileInfo.LineContents {
			if lineNum > maxLine {
				maxLine = lineNum
			}
		}

		// LineContentsから行内容を取得
		fileContent = make([]string, maxLine)
		for i := range fileContent {
			lineNum := i + 1
			if content, exists := fileInfo.LineContents[lineNum]; exists {
				fileContent[i] = content
			}
		}

		fmt.Printf("JSONから '%s' の内容を読み込みました (%d行)\n", filePath, maxLine)
	} else {
		// JSONに行内容がない場合、リポジトリからファイル内容を読み込む（後方互換性）
		var readErr error
		fileContent, readErr = v.readFileContent(filePath)
		if readErr != nil {
			// エラーの場合は空の内容で続行
			fileContent = []string{"// ファイルの内容を読み込めませんでした"}
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

	// HTMLエスケープ処理
	escapedTitle := html.EscapeString(title)

	// SVGヘッダーを書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <text x="600" y="40" font-size="24" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle" fill="#333333">ファイルヒートマップ</text>
  <text x="600" y="70" font-size="14" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle" fill="#333333">%s</text>
  <text x="600" y="90" font-size="12" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle" fill="#666666">変更回数: %d, 最終更新: %s</text>

  <!-- スクロール可能なコードビュー -->
  <g transform="translate(%d, %d)">
    <!-- 行番号とコード-->
`,
		maxWidth, canvasHeight,
		escapedTitle, fileInfo.ChangeCount, fileInfo.LastModified.Format("2006/01/02"),
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
			if normalizedHeat > 1.0 {
				normalizedHeat = 1.0
			}
			lineColor = GetHeatColor(normalizedHeat)
		}

		// テキストをエスケープ
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
