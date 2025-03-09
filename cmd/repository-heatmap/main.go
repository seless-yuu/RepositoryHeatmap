package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/repositoryheatmap/internal/git"
	"github.com/repositoryheatmap/internal/heatmap"
	"github.com/repositoryheatmap/internal/utils"
	"github.com/repositoryheatmap/pkg/models"
	"github.com/spf13/pflag"
)

var (
	repoPath    string
	outputDir   string
	outputType  string
	skipClone   bool
	showVersion bool
	showHelp    bool
	debugLog    bool
	sinceDate   string
	workers     int
	maxFiles    int
	version     = "0.1.0" // アプリケーションバージョン
	logFile     *os.File
	logger      *log.Logger
	inputFile   string
	filePattern string
	fileType    string
)

// analyzeコマンド用のフラグ設定
func setupAnalyzeFlags(analyzeCmd *pflag.FlagSet) {
	analyzeCmd.StringVarP(&repoPath, "repo", "r", "", "解析するGitリポジトリのパスまたはURL")
	analyzeCmd.StringVarP(&outputDir, "output", "o", "output", "解析結果の出力ディレクトリ")
	analyzeCmd.BoolVarP(&skipClone, "skip-clone", "s", false, "リポジトリが既にクローンされている場合はスキップ")
	analyzeCmd.StringVar(&sinceDate, "since", "", "指定日付以降のコミットのみを解析 (YYYY-MM-DD形式)")
	analyzeCmd.IntVarP(&workers, "workers", "w", 0, "並列ワーカー数（0: 自動、CPU数に基づく）")
	analyzeCmd.BoolVarP(&showHelp, "help", "h", false, "ヘルプを表示")
	analyzeCmd.BoolVarP(&debugLog, "debug", "d", false, "デバッグログをファイルに出力")
	analyzeCmd.StringVarP(&filePattern, "file-pattern", "p", "", "解析対象のファイルパターン（ワイルドカード対応、例: '*.go'）")
	analyzeCmd.StringVarP(&fileType, "file-type", "t", "", "解析対象のファイル種別（拡張子のみ、例: 'go'）")
}

// visualizeコマンド用のフラグ設定
func setupVisualizeFlags(visualizeCmd *pflag.FlagSet) {
	visualizeCmd.StringVarP(&outputDir, "output", "o", "output", "ヒートマップ出力ディレクトリ")
	visualizeCmd.StringVarP(&outputType, "format", "f", "svg", "出力形式 (svg または webp)")
	visualizeCmd.StringVarP(&repoPath, "repo", "r", "", "ファイル内容表示のためのリポジトリパス（必須ではない）")
	visualizeCmd.IntVarP(&maxFiles, "max-files", "m", 100, "ヒートマップに表示する最大ファイル数")
	visualizeCmd.BoolVarP(&debugLog, "debug", "d", false, "デバッグログをファイルに出力")
	visualizeCmd.StringVarP(&inputFile, "input", "i", "", "入力JSONファイルのパス（指定しない場合は出力ディレクトリから自動選択）")
	visualizeCmd.BoolVarP(&showHelp, "help", "h", false, "ヘルプを表示")
}

// analyzeコマンドのヘルプメッセージ
func showAnalyzeHelp(analyzeCmd *pflag.FlagSet) {
	fmt.Println("Repository Heatmap - リポジトリ解析コマンド")
	fmt.Println("=========================================================")
	fmt.Printf("Usage: %s analyze [options]\n\n", os.Args[0])
	fmt.Println("Gitリポジトリを解析し、ヒートマップデータを生成します。")
	fmt.Println()
	fmt.Println("Options:")

	// pflagのデフォルトでは短いオプションがデフォルトのPrintDefaultsで表示されないため、
	// カスタマイズしたヘルプ表示を使用
	fmt.Printf("  -d, --debug\n")
	fmt.Printf("        デバッグログをファイルに出力\n")
	fmt.Printf("  -p, --file-pattern string\n")
	fmt.Printf("        解析対象のファイルパターン（ワイルドカード対応、例: '*.go'）\n")
	fmt.Printf("  -t, --file-type string\n")
	fmt.Printf("        解析対象のファイル種別（拡張子のみ、例: 'go'）\n")
	fmt.Printf("  -h, --help\n")
	fmt.Printf("        ヘルプを表示\n")
	fmt.Printf("  -o, --output string (default \"output\")\n")
	fmt.Printf("        解析結果の出力ディレクトリ\n")
	fmt.Printf("  -r, --repo string\n")
	fmt.Printf("        解析するGitリポジトリのパスまたはURL\n")
	fmt.Printf("  --since string\n")
	fmt.Printf("        指定日付以降のコミットのみを解析 (YYYY-MM-DD形式)\n")
	fmt.Printf("  -s, --skip-clone\n")
	fmt.Printf("        リポジトリが既にクローンされている場合はスキップ\n")
	fmt.Printf("  -w, --workers int\n")
	fmt.Printf("        並列ワーカー数（0: 自動、CPU数に基づく）\n")

	fmt.Println()
	fmt.Println("例:")
	fmt.Printf("  %s analyze --repo=./myrepo --output=./results\n", os.Args[0])
	fmt.Printf("  %s analyze -r ./myrepo -o ./results\n", os.Args[0])
	fmt.Printf("  %s analyze --repo=https://github.com/username/repo.git\n", os.Args[0])
	fmt.Printf("  %s analyze --repo=./myrepo --since=2023-01-01\n", os.Args[0])
	fmt.Printf("  %s analyze --repo=./myrepo --workers=4\n", os.Args[0])
	fmt.Printf("  %s analyze -r ./myrepo -w 4\n", os.Args[0])
	fmt.Printf("  %s analyze --repo=./myrepo --file-pattern=\"*.go\" --output=./results\n", os.Args[0])
	fmt.Printf("  %s analyze -r ./myrepo -p \"*.go\" -o ./results\n", os.Args[0])
	fmt.Printf("  %s analyze --repo=./myrepo --file-type=js --output=./results\n", os.Args[0])
	fmt.Printf("  %s analyze -r ./myrepo -t js -o ./results\n", os.Args[0])
}

// visualizeコマンドのヘルプメッセージ
func showVisualizeHelp(visualizeCmd *pflag.FlagSet) {
	fmt.Println("Repository Heatmap - ヒートマップ可視化コマンド")
	fmt.Println("=========================================================")
	fmt.Printf("Usage: %s visualize [options]\n\n", os.Args[0])
	fmt.Println("リポジトリ解析データからヒートマップを生成します。")
	fmt.Println()
	fmt.Println("Options:")

	// pflagのデフォルトでは短いオプションがデフォルトのPrintDefaultsで表示されないため、
	// カスタマイズしたヘルプ表示を使用
	fmt.Printf("  -d, --debug\n")
	fmt.Printf("        デバッグログをファイルに出力\n")
	fmt.Printf("  -f, --format string (default \"svg\")\n")
	fmt.Printf("        出力形式 (svg または webp)\n")
	fmt.Printf("  -h, --help\n")
	fmt.Printf("        ヘルプを表示\n")
	fmt.Printf("  -i, --input string\n")
	fmt.Printf("        入力JSONファイルのパス（指定しない場合は出力ディレクトリから自動選択）\n")
	fmt.Printf("  -m, --max-files int (default 100)\n")
	fmt.Printf("        ヒートマップに表示する最大ファイル数\n")
	fmt.Printf("  -o, --output string (default \"output\")\n")
	fmt.Printf("        ヒートマップ出力ディレクトリ\n")
	fmt.Printf("  -r, --repo string\n")
	fmt.Printf("        ファイル内容表示のためのリポジトリパス（必須ではない）\n")

	fmt.Println()
	fmt.Println("出力形式:")
	fmt.Println("  svg  - SVG形式のベクターグラフィックス（リポジトリ全体と個別ファイルのヒートマップを生成）")
	fmt.Println("  webp - WebP形式の画像（実験的機能）")
	fmt.Println()
	fmt.Println("例:")
	fmt.Printf("  %s visualize --input=./results/repo-heatmap.json --output=./results\n", os.Args[0])
	fmt.Printf("  %s visualize -i ./results/repo-heatmap.json -o ./results\n", os.Args[0])
	fmt.Printf("  %s visualize --input=./results/repo-heatmap.json --format=svg\n", os.Args[0])
	fmt.Printf("  %s visualize -i ./results/repo-heatmap.json -f svg\n", os.Args[0])
	fmt.Printf("  %s visualize --output=./results --format=svg\n", os.Args[0])
	fmt.Printf("  %s visualize -o ./results -f svg\n", os.Args[0])
	fmt.Printf("  %s visualize --output=./results --repo=/path/to/repo\n", os.Args[0])
	fmt.Printf("  %s visualize -o ./results -r /path/to/repo\n", os.Args[0])
	fmt.Printf("  %s visualize --input=./results/repo-heatmap.json --max-files=200\n", os.Args[0])
	fmt.Printf("  %s visualize -i ./results/repo-heatmap.json -m 200\n", os.Args[0])
}

// 全体のヘルプメッセージ
func showMainHelp() {
	fmt.Println("Repository Heatmap - Gitリポジトリの変更頻度解析ツール")
	fmt.Println("=========================================================")
	fmt.Printf("Usage: %s <command> [options]\n\n", os.Args[0])
	fmt.Println("Commands:")
	fmt.Println("  analyze    リポジトリを解析しJSONデータを生成")
	fmt.Println("  visualize  JSONデータからヒートマップを生成")
	fmt.Println()
	fmt.Println("バージョン情報:")
	fmt.Printf("  %s --version\n", os.Args[0])
	fmt.Println("\nオプションを確認するには:")
	fmt.Printf("  %s analyze --help\n", os.Args[0])
	fmt.Printf("  %s visualize --help\n", os.Args[0])
}

// logDebug はデバッグメッセージをログとコンソールに出力
func logDebug(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)

	// コンソールに出力
	fmt.Println(message)

	// ロガーに出力（設定されている場合）
	if logger != nil {
		logger.Println(message)
	}
}

// analyzeCommand はリポジトリの解析を行い、JSONデータを生成する
func analyzeCommand(args []string) {
	// pflagを使用してフラグを設定
	analyzeCmd := pflag.NewFlagSet("analyze", pflag.ExitOnError)

	// フラグの設定
	setupAnalyzeFlags(analyzeCmd)

	// ヘルプメッセージのカスタマイズ
	analyzeCmd.Usage = func() {
		showAnalyzeHelp(analyzeCmd)
	}

	// 特殊なケース：明示的に-hや--helpが指定された場合に対応
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			analyzeCmd.Usage()
			return
		}
	}

	// フラグを解析
	if err := analyzeCmd.Parse(args); err != nil {
		fmt.Printf("フラグの解析エラー: %v\n", err)
		analyzeCmd.Usage()
		os.Exit(1)
	}

	// showHelpフラグが直接設定された場合
	if showHelp {
		analyzeCmd.Usage()
		return
	}

	// デバッグログの設定
	setupDebugLog()

	// コマンドを表示
	logDebug("実行コマンド: %s analyze", os.Args[0])

	// 日付パラメータのパース
	var since *time.Time
	if sinceDate != "" {
		parsedTime, err := time.Parse("2006-01-02", sinceDate)
		if err != nil {
			logDebug("エラー: 日付形式が正しくありません。YYYY-MM-DD形式で指定してください: %v", err)
			os.Exit(1)
		}
		since = &parsedTime
		logDebug("指定された日付以降のコミットのみを解析します: %s", parsedTime.Format("2006-01-02"))
	}

	// 並列ワーカー数の設定
	if workers < 0 {
		logDebug("エラー: ワーカー数は0以上を指定してください")
		os.Exit(1)
	} else if workers == 0 {
		workers = utils.GetNumCPUs()
		logDebug("利用可能なCPUコア数に基づいてワーカー数を設定: %d", workers)
	} else {
		logDebug("ワーカー数: %d", workers)
	}

	// 出力ディレクトリ作成
	if err := utils.EnsureDirectoryExists(outputDir); err != nil {
		logDebug("出力ディレクトリの作成に失敗しました: %v", err)
		os.Exit(1)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		logDebug("出力パスの解決に失敗しました: %v", err)
		os.Exit(1)
	}

	logDebug("リポジトリヒートマップ解析を開始します...")
	logDebug("リポジトリ: %s", repoPath)
	logDebug("出力ディレクトリ: %s", absOutputDir)

	// リポジトリの解析
	analyzer, err := git.NewAnalyzer(repoPath)
	if err != nil {
		logDebug("リポジトリアナライザの作成に失敗しました: %v", err)
		os.Exit(1)
	}

	// 並列ワーカー数を設定
	analyzer.SetWorkerCount(workers)

	// 日付フィルタを設定
	if since != nil {
		analyzer.SetSinceDate(*since)
	}

	// ファイルフィルタを設定
	if filePattern != "" {
		logDebug("ファイルパターンでフィルタリング: %s", filePattern)
		analyzer.SetFilePattern(filePattern)
	}

	if fileType != "" {
		logDebug("ファイル種別でフィルタリング: %s", fileType)
		analyzer.SetFileType(fileType)
	}

	// リポジトリのオープンまたはクローン
	if utils.IsLocalRepository(repoPath) {
		logDebug("ローカルリポジトリを開きます: %s", repoPath)
		if err := analyzer.Open(); err != nil {
			logDebug("リポジトリを開けませんでした: %v", err)
			os.Exit(1)
		}
	} else if utils.IsValidGitURL(repoPath) {
		// クローン用の一時ディレクトリを作成
		tempDir := filepath.Join(outputDir, "repo-clone")
		if err := utils.EnsureDirectoryExists(tempDir); err != nil {
			logDebug("クローン用ディレクトリの作成に失敗しました: %v", err)
			os.Exit(1)
		}

		logDebug("リポジトリをクローンします: %s -> %s", repoPath, tempDir)
		if err := analyzer.Clone(repoPath); err != nil {
			logDebug("リポジトリのクローンに失敗しました: %v", err)
			os.Exit(1)
		}
	} else {
		logDebug("無効なリポジトリパスまたはURL: %s", repoPath)
		os.Exit(1)
	}

	// リポジトリの解析（マルチスレッド対応）
	logDebug("コミット履歴を解析中...")
	stats, err := analyzer.Analyze()
	if err != nil {
		logDebug("リポジトリの解析に失敗しました: %v", err)
		os.Exit(1)
	}

	// 行単位での変更追跡
	logDebug("行単位の変更を追跡中...")
	lineTracker := git.NewLineTracker(analyzer)
	lineTracker.SetWorkerCount(workers) // 行変更追跡も並列化
	if err := lineTracker.TrackLineChanges(stats); err != nil {
		logDebug("行単位の変更追跡に失敗しました: %v", err)
		os.Exit(1)
	}
	lineTracker.CalculateLineHeatLevels(stats)

	// リポジトリ名を含むJSON出力ファイル名を生成
	repoName := stats.RepositoryName
	jsonFilePath, err := writeJSONResult(stats, absOutputDir, repoName)
	if err != nil {
		logDebug("JSONファイルの書き込みに失敗しました: %v", err)
		os.Exit(1)
	}

	logDebug("ヒートマップデータをJSONに保存しました: %s", jsonFilePath)
	logDebug("解析が完了しました")
}

// visualizeCommand はJSON出力からヒートマップを生成する
func visualizeCommand(args []string) {
	// pflagを使用してフラグを設定
	visualizeCmd := pflag.NewFlagSet("visualize", pflag.ExitOnError)

	// フラグの設定
	setupVisualizeFlags(visualizeCmd)

	// ヘルプメッセージのカスタマイズ
	visualizeCmd.Usage = func() {
		showVisualizeHelp(visualizeCmd)
	}

	// 特殊なケース：明示的に-hや--helpが指定された場合に対応
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			visualizeCmd.Usage()
			return
		}
	}

	// フラグを解析
	if err := visualizeCmd.Parse(args); err != nil {
		fmt.Printf("フラグの解析エラー: %v\n", err)
		visualizeCmd.Usage()
		os.Exit(1)
	}

	// showHelpフラグが直接設定された場合
	if showHelp {
		visualizeCmd.Usage()
		return
	}

	// デバッグログの設定
	setupDebugLog()

	// コマンドを表示
	logDebug("実行コマンド: %s visualize", os.Args[0])

	// 出力ディレクトリのチェック
	if outputDir == "" {
		logDebug("エラー: 出力ディレクトリを指定してください（--output または -o オプション）\n")
		visualizeCmd.Usage()
		os.Exit(1)
	}

	// 最大ファイル数のチェック
	if maxFiles < 1 {
		logDebug("警告: 最大ファイル数は1以上である必要があります。1に設定します")
		maxFiles = 1
	} else {
		logDebug("表示する最大ファイル数: %d", maxFiles)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		logDebug("出力パスの解決に失敗しました: %v", err)
		os.Exit(1)
	}

	// 入力JSONファイルを決定
	var jsonFilePath string

	if inputFile != "" {
		// 1. コマンドラインオプションで指定されたJSONファイル
		absInputFile, err := filepath.Abs(inputFile)
		if err != nil {
			logDebug("入力ファイルパスの解決に失敗しました: %v", err)
			os.Exit(1)
		}

		// ファイルの存在確認
		if _, err := os.Stat(absInputFile); err != nil {
			logDebug("指定された入力ファイルが見つかりません: %s", absInputFile)
			os.Exit(1)
		}

		jsonFilePath = absInputFile
		logDebug("指定された入力JSONファイルを使用します: %s", jsonFilePath)
	} else {
		// 2. 最新のJSONファイルを検索
		latestFile, err := findLatestJSONFile(absOutputDir)
		if err != nil {
			logDebug("JSONファイルの検索に失敗しました: %v", err)
			logDebug("エラー: 入力JSONファイルが見つかりませんでした。--input (-i) オプションで指定するか、出力ディレクトリに有効なJSONファイルを配置してください")
			os.Exit(1)
		}
		jsonFilePath = latestFile
		logDebug("最新のJSONファイルを使用します: %s", jsonFilePath)
	}

	logDebug("JSONファイルを読み込みます: %s", jsonFilePath)

	// JSONファイルからデータを読み込む
	stats, err := readJSONFile(jsonFilePath)
	if err != nil {
		logDebug("JSONファイルの読み込みに失敗しました: %v", err)
		os.Exit(1)
	}

	// 特定のディレクトリ（.github, testdata）を除外するフィルタリング処理
	filteredFiles := make(map[string]models.FileChangeInfo)
	originalFileCount := len(stats.Files)
	for path, fileInfo := range stats.Files {
		// .githubディレクトリは除外する（標準的な設定ファイルのため）
		if !strings.HasPrefix(path, ".github/") {
			filteredFiles[path] = fileInfo
		}
	}
	// フィルタリングされたファイルリストで置き換え
	stats.Files = filteredFiles
	logDebug("ファイル数: %d (フィルタリング前) -> %d (フィルタリング後)", originalFileCount, len(filteredFiles))

	// リポジトリパスが指定されていれば、その情報を使用
	var localRepoPath string
	if repoPath != "" {
		if !utils.IsLocalRepository(repoPath) {
			logDebug("警告: 指定されたリポジトリパスが見つかりません。ファイル内容表示が無効になります")
		} else {
			absRepoPath, err := filepath.Abs(repoPath)
			if err != nil {
				logDebug("リポジトリパスの解決に失敗しました: %v", err)
			} else {
				localRepoPath = absRepoPath
			}
		}
	}

	logDebug("ヒートマップを可視化中...")
	logDebug("出力ディレクトリ: %s", absOutputDir)

	// 通常の静的可視化
	logDebug("出力形式: %s", outputType)

	// ヒートマップの可視化
	visualizer := heatmap.NewVisualizer(absOutputDir, outputType, stats)

	// リポジトリのパスが設定されていれば、それを使用（ファイル内容の読み込みに必要）
	if localRepoPath != "" {
		visualizer.SetRepoPath(localRepoPath)
	}

	// 最大表示ファイル数を設定
	visualizer.SetMaxFilesToShow(maxFiles)

	if err := visualizer.Visualize(); err != nil {
		logDebug("ヒートマップの可視化に失敗しました: %v", err)
		os.Exit(1)
	}

	logDebug("可視化が完了しました")
	logDebug("結果は %s ディレクトリに保存されました", absOutputDir)
}

// デバッグログの設定
func setupDebugLog() {
	if debugLog {
		var err error
		logFile, err = os.Create("repository-heatmap-debug.log")
		if err != nil {
			fmt.Printf("デバッグログファイルを作成できませんでした: %v\n", err)
		} else {
			defer logFile.Close()
			logger = log.New(logFile, "", log.LstdFlags)
			logger.Println("Repository Heatmap デバッグログ開始")
		}
	}
}

// JSONファイルに解析結果を書き込む
func writeJSONResult(data *models.RepositoryStats, outputDir, repoName string) (string, error) {
	// 出力ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("出力ディレクトリの作成に失敗しました: %v", err)
	}

	// JSONファイルパスの生成
	jsonFilePath := filepath.Join(outputDir, fmt.Sprintf("%s-heatmap.json", repoName))

	// JSONデータの生成
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSONエンコードに失敗しました: %v", err)
	}

	// ファイルに書き込み
	if err := os.WriteFile(jsonFilePath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("JSONファイルの書き込みに失敗しました: %v", err)
	}

	return jsonFilePath, nil
}

// JSONファイルを読み込む
func readJSONFile(jsonFilePath string) (*models.RepositoryStats, error) {
	// JSONファイルの読み込み
	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("JSONファイルの読み込みに失敗しました: %v", err)
	}

	// JSONデータのデコード
	var data models.RepositoryStats
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("JSONデコードに失敗しました: %v", err)
	}

	return &data, nil
}

// 最新のJSONファイルを検索
func findLatestJSONFile(outputDir string) (string, error) {
	// 出力ディレクトリの絶対パスを取得
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("出力ディレクトリの絶対パス取得に失敗しました: %v", err)
	}

	// ディレクトリ内のファイル一覧を取得
	files, err := os.ReadDir(absOutputDir)
	if err != nil {
		return "", fmt.Errorf("ディレクトリの読み込みに失敗しました: %v", err)
	}

	var latestTime time.Time
	var latestFile string

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		if !info.IsDir() && filepath.Ext(file.Name()) == ".json" {
			// JSONファイルの中で最新のものを選択
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestFile = filepath.Join(absOutputDir, file.Name())
			}
		}
	}

	if latestFile == "" {
		return "", fmt.Errorf("JSONファイルが見つかりませんでした")
	}

	return latestFile, nil
}

func main() {
	// グローバルフラグの設定
	globalFlags := pflag.NewFlagSet("global", pflag.ExitOnError)
	globalFlags.BoolVarP(&showVersion, "version", "v", false, "バージョン情報を表示")
	globalFlags.BoolVarP(&showHelp, "help", "h", false, "ヘルプを表示")

	// グローバルフラグを解析（先頭の引数だけを見る）
	if len(os.Args) > 1 {
		// --versionフラグのチェック
		if os.Args[1] == "--version" || os.Args[1] == "-v" {
			fmt.Printf("Repository Heatmap v%s\n", version)
			return
		}

		// --helpフラグのチェック
		if os.Args[1] == "--help" || os.Args[1] == "-h" {
			showMainHelp()
			return
		}
	}

	// サブコマンドがない場合はヘルプを表示
	if len(os.Args) < 2 {
		showMainHelp()
		os.Exit(1)
	}

	// サブコマンドの処理
	switch os.Args[1] {
	case "analyze":
		analyzeCommand(os.Args[2:])
	case "visualize":
		visualizeCommand(os.Args[2:])
	default:
		fmt.Printf("エラー: 不明なコマンド '%s'\n", os.Args[1])
		showMainHelp()
		os.Exit(1)
	}
}
