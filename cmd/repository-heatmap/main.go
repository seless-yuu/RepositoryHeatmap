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
	version     = "1.0.0" // Application version
	logFile     *os.File
	logger      *log.Logger
	inputFile   string
	filePattern string
	fileType    string
)

// analyzeコマンド用のフラグ設定
func setupAnalyzeFlags(analyzeCmd *pflag.FlagSet) {
	analyzeCmd.StringVarP(&repoPath, "repo", "r", "", "Git repository path or URL to analyze")
	analyzeCmd.StringVarP(&outputDir, "output", "o", "output", "Output directory for analysis results")
	analyzeCmd.BoolVarP(&skipClone, "skip-clone", "s", false, "Skip cloning if repository already exists")
	analyzeCmd.StringVar(&sinceDate, "since", "", "Analyze commits only after this date (YYYY-MM-DD format)")
	analyzeCmd.IntVarP(&workers, "workers", "w", 0, "Number of parallel workers (0: automatic, based on CPU count)")
	analyzeCmd.BoolVarP(&showHelp, "help", "h", false, "Show help")
	analyzeCmd.BoolVarP(&debugLog, "debug", "d", false, "Output debug logs to file")
	analyzeCmd.StringVarP(&filePattern, "file-pattern", "p", "", "File pattern to analyze (wildcard support, e.g., '*.go')")
	analyzeCmd.StringVarP(&fileType, "file-type", "t", "", "File type to analyze (extension only, e.g., 'go')")
}

// visualizeコマンド用のフラグ設定
func setupVisualizeFlags(visualizeCmd *pflag.FlagSet) {
	visualizeCmd.StringVarP(&outputDir, "output", "o", "output", "Output directory for heatmap")
	visualizeCmd.StringVarP(&outputType, "format", "f", "svg", "Output format (svg or webp)")
	visualizeCmd.StringVarP(&repoPath, "repo", "r", "", "Repository path for file content display (not required)")
	visualizeCmd.IntVarP(&maxFiles, "max-files", "m", 100, "Maximum number of files to display in the heatmap")
	visualizeCmd.BoolVarP(&debugLog, "debug", "d", false, "Output debug logs to file")
	visualizeCmd.StringVarP(&inputFile, "input", "i", "", "Input JSON file path (if not specified, automatically selected from output directory)")
	visualizeCmd.BoolVarP(&showHelp, "help", "h", false, "Show help")
}

// Show help message for analyze command
func showAnalyzeHelp(analyzeCmd *pflag.FlagSet) {
	fmt.Println("Repository Heatmap - Repository Analysis Command")
	fmt.Println("=========================================================")
	fmt.Printf("Usage: %s analyze [options]\n\n", os.Args[0])
	fmt.Println("Analyzes a Git repository and generates heatmap data.")
	fmt.Println()
	fmt.Println("Options:")

	// pflagのデフォルトでは短いオプションがデフォルトのPrintDefaultsで表示されないため、
	// カスタマイズしたヘルプ表示を使用
	fmt.Printf("  -d, --debug\n")
	fmt.Printf("        Output debug logs to file\n")
	fmt.Printf("  -p, --file-pattern string\n")
	fmt.Printf("        File pattern to analyze (wildcard support, e.g., '*.go')\n")
	fmt.Printf("  -t, --file-type string\n")
	fmt.Printf("        File type to analyze (extension only, e.g., 'go')\n")
	fmt.Printf("  -h, --help\n")
	fmt.Printf("        Show help\n")
	fmt.Printf("  -o, --output string (default \"output\")\n")
	fmt.Printf("        Output directory for analysis results\n")
	fmt.Printf("  -r, --repo string\n")
	fmt.Printf("        Git repository path or URL to analyze\n")
	fmt.Printf("  --since string\n")
	fmt.Printf("        Analyze commits only after this date (YYYY-MM-DD format)\n")
	fmt.Printf("  -s, --skip-clone\n")
	fmt.Printf("        Skip cloning if repository already exists\n")
	fmt.Printf("  -w, --workers int\n")
	fmt.Printf("        Number of parallel workers (0: automatic, based on CPU count)\n")

	fmt.Println()
	fmt.Println("Examples:")
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

// Show help message for visualize command
func showVisualizeHelp(visualizeCmd *pflag.FlagSet) {
	fmt.Println("Repository Heatmap - Heatmap Visualization Command")
	fmt.Println("=========================================================")
	fmt.Printf("Usage: %s visualize [options]\n\n", os.Args[0])
	fmt.Println("Generates a heatmap from repository analysis data.")
	fmt.Println()
	fmt.Println("Options:")

	// pflagのデフォルトでは短いオプションがデフォルトのPrintDefaultsで表示されないため、
	// カスタマイズしたヘルプ表示を使用
	fmt.Printf("  -d, --debug\n")
	fmt.Printf("        Output debug logs to file\n")
	fmt.Printf("  -f, --format string (default \"svg\")\n")
	fmt.Printf("        Output format (svg or webp)\n")
	fmt.Printf("  -h, --help\n")
	fmt.Printf("        Show help\n")
	fmt.Printf("  -i, --input string\n")
	fmt.Printf("        Input JSON file path (if not specified, automatically selected from output directory)\n")
	fmt.Printf("  -m, --max-files int (default 100)\n")
	fmt.Printf("        Maximum number of files to display in the heatmap\n")
	fmt.Printf("  -o, --output string (default \"output\")\n")
	fmt.Printf("        Output directory for heatmap\n")
	fmt.Printf("  -r, --repo string\n")
	fmt.Printf("        Repository path for file content display (not required)\n")

	fmt.Println()
	fmt.Println("Output Formats:")
	fmt.Println("  svg  - SVG vector graphics (generates heatmaps for the entire repository and individual files)")
	fmt.Println("  webp - WebP image format (experimental feature)")
	fmt.Println()
	fmt.Println("Examples:")
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
	fmt.Println("Repository Heatmap - Change Frequency Analysis Tool for Git Repositories")
	fmt.Println("=========================================================")
	fmt.Printf("Usage: %s <command> [options]\n\n", os.Args[0])
	fmt.Println("Commands:")
	fmt.Println("  analyze    Analyze repository and generate JSON data")
	fmt.Println("  visualize  Generate heatmap from JSON data")
	fmt.Println()
	fmt.Println("Version Information:")
	fmt.Printf("  %s --version\n", os.Args[0])
	fmt.Println("\nTo check options:")
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
		fmt.Printf("Flag parsing error: %v\n", err)
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
	logDebug("Command executed: %s analyze", os.Args[0])

	// 日付パラメータのパース
	var since *time.Time
	if sinceDate != "" {
		parsedTime, err := time.Parse("2006-01-02", sinceDate)
		if err != nil {
			logDebug("Error: Date format is incorrect. Please specify in YYYY-MM-DD format: %v", err)
			os.Exit(1)
		}
		since = &parsedTime
		logDebug("Analyzing commits only after specified date: %s", parsedTime.Format("2006-01-02"))
	}

	// 並列ワーカー数の設定
	if workers < 0 {
		logDebug("Error: Worker count must be 0 or greater")
		os.Exit(1)
	} else if workers == 0 {
		workers = utils.GetNumCPUs()
		logDebug("Setting worker count based on available CPU cores: %d", workers)
	} else {
		logDebug("Worker count: %d", workers)
	}

	// 出力ディレクトリ作成
	if err := utils.EnsureDirectoryExists(outputDir); err != nil {
		logDebug("Failed to create output directory: %v", err)
		os.Exit(1)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		logDebug("Failed to resolve output path: %v", err)
		os.Exit(1)
	}

	logDebug("Starting repository heatmap analysis...")
	logDebug("Repository: %s", repoPath)
	logDebug("Output directory: %s", absOutputDir)

	// リポジトリの解析
	analyzer, err := git.NewAnalyzer(repoPath)
	if err != nil {
		logDebug("Failed to create repository analyzer: %v", err)
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
		logDebug("Filtering by file pattern: %s", filePattern)
		analyzer.SetFilePattern(filePattern)
	}

	if fileType != "" {
		logDebug("Filtering by file type: %s", fileType)
		analyzer.SetFileType(fileType)
	}

	// リポジトリのオープンまたはクローン
	if utils.IsLocalRepository(repoPath) {
		logDebug("Opening local repository: %s", repoPath)
		if err := analyzer.Open(); err != nil {
			logDebug("Failed to open repository: %v", err)
			os.Exit(1)
		}
	} else if utils.IsValidGitURL(repoPath) {
		// クローン用の一時ディレクトリを作成
		tempDir := filepath.Join(outputDir, "repo-clone")
		if err := utils.EnsureDirectoryExists(tempDir); err != nil {
			logDebug("Failed to create directory for cloning: %v", err)
			os.Exit(1)
		}

		logDebug("Cloning repository: %s -> %s", repoPath, tempDir)
		if err := analyzer.Clone(repoPath); err != nil {
			logDebug("Failed to clone repository: %v", err)
			os.Exit(1)
		}
	} else {
		logDebug("Invalid repository path or URL: %s", repoPath)
		os.Exit(1)
	}

	// リポジトリの解析（マルチスレッド対応）
	logDebug("Analyzing commit history...")
	stats, err := analyzer.Analyze()
	if err != nil {
		logDebug("Failed to analyze repository: %v", err)
		os.Exit(1)
	}

	// 行単位での変更追跡
	logDebug("Tracking line-level changes...")
	lineTracker := git.NewLineTracker(analyzer)
	lineTracker.SetWorkerCount(workers) // 行変更追跡も並列化
	if err := lineTracker.TrackLineChanges(stats); err != nil {
		logDebug("Failed to track line-level changes: %v", err)
		os.Exit(1)
	}
	lineTracker.CalculateLineHeatLevels(stats)

	// リポジトリ名を含むJSON出力ファイル名を生成
	repoName := stats.RepositoryName
	jsonFilePath, err := writeJSONResult(stats, absOutputDir, repoName)
	if err != nil {
		logDebug("Failed to write JSON file: %v", err)
		os.Exit(1)
	}

	logDebug("Heatmap data saved to JSON: %s", jsonFilePath)
	logDebug("Analysis completed")
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
		fmt.Printf("Flag parsing error: %v\n", err)
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
	logDebug("Command executed: %s visualize", os.Args[0])

	// 出力ディレクトリのチェック
	if outputDir == "" {
		logDebug("Error: Please specify an output directory (--output or -o option)\n")
		visualizeCmd.Usage()
		os.Exit(1)
	}

	// 最大ファイル数のチェック
	if maxFiles < 1 {
		logDebug("Warning: Maximum file count must be greater than or equal to 1. Setting to 1")
		maxFiles = 1
	} else {
		logDebug("Maximum number of files to display: %d", maxFiles)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		logDebug("Failed to resolve output path: %v", err)
		os.Exit(1)
	}

	// 入力JSONファイルを決定
	var jsonFilePath string

	if inputFile != "" {
		// 1. コマンドラインオプションで指定されたJSONファイル
		absInputFile, err := filepath.Abs(inputFile)
		if err != nil {
			logDebug("Failed to resolve input file path: %v", err)
			os.Exit(1)
		}

		// ファイルの存在確認
		if _, err := os.Stat(absInputFile); err != nil {
			logDebug("Specified input file not found: %s", absInputFile)
			os.Exit(1)
		}

		jsonFilePath = absInputFile
		logDebug("Using specified input JSON file: %s", jsonFilePath)
	} else {
		// 2. 最新のJSONファイルを検索
		latestFile, err := findLatestJSONFile(outputDir)
		if err != nil {
			logDebug("Failed to search JSON file: %v", err)
			logDebug("Error: Input JSON file not found. Please specify with --input (-i) option or place a valid JSON file in the output directory")
			os.Exit(1)
		}
		jsonFilePath = latestFile
		logDebug("Using latest JSON file: %s", jsonFilePath)
	}

	logDebug("Reading JSON file: %s", jsonFilePath)

	// JSONファイルからデータを読み込む
	stats, err := readJSONFile(jsonFilePath)
	if err != nil {
		logDebug("Failed to read JSON file: %v", err)
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
	logDebug("File count: %d (before filtering) -> %d (after filtering)", originalFileCount, len(filteredFiles))

	// リポジトリパスが指定されていれば、その情報を使用
	var localRepoPath string
	if repoPath != "" {
		if !utils.IsLocalRepository(repoPath) {
			logDebug("Warning: Specified repository path not found. File content display will be disabled")
		} else {
			absRepoPath, err := filepath.Abs(repoPath)
			if err != nil {
				logDebug("Failed to resolve repository path: %v", err)
			} else {
				localRepoPath = absRepoPath
			}
		}
	}

	logDebug("Visualizing heatmap...")
	logDebug("Output directory: %s", absOutputDir)

	// 通常の静的可視化
	logDebug("Output format: %s", outputType)

	// ヒートマップの可視化
	visualizer := heatmap.NewVisualizer(absOutputDir, outputType, stats, maxFiles, inputFile)

	// リポジトリのパスが設定されていれば、それを使用（ファイル内容の読み込みに必要）
	if localRepoPath != "" {
		visualizer.SetRepoPath(localRepoPath)
	}

	if err := visualizer.Visualize(); err != nil {
		logDebug("Failed to visualize heatmap: %v", err)
		os.Exit(1)
	}

	logDebug("Visualization completed")
	logDebug("Results saved in %s directory", absOutputDir)
}

// デバッグログの設定
func setupDebugLog() {
	if debugLog {
		var err error
		logFile, err = os.Create("repository-heatmap-debug.log")
		if err != nil {
			fmt.Printf("Failed to create debug log file: %v\n", err)
		} else {
			defer logFile.Close()
			logger = log.New(logFile, "", log.LstdFlags)
			logger.Println("Repository Heatmap debug log started")
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

// findLatestJSONFile は最新のJSONファイルを検索する
func findLatestJSONFile(outputDir string) (string, error) {
	// 検索対象ディレクトリ
	searchDirs := []string{
		outputDir,     // 指定された出力ディレクトリ
		".",           // カレントディレクトリ
		"test-output", // test-outputディレクトリ
		"output",      // outputディレクトリ
	}

	var latestTime time.Time
	var latestFile string
	var searchErrors []string

	// 各ディレクトリを検索
	for _, dir := range searchDirs {
		// ディレクトリの絶対パスを取得
		absDir, err := filepath.Abs(dir)
		if err != nil {
			searchErrors = append(searchErrors, fmt.Sprintf("%s: %v", dir, err))
			continue
		}

		// ディレクトリ内のファイル一覧を取得
		files, err := os.ReadDir(absDir)
		if err != nil {
			searchErrors = append(searchErrors, fmt.Sprintf("%s: %v", absDir, err))
			continue
		}

		// JSONファイルを検索
		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				continue
			}

			if !info.IsDir() && filepath.Ext(file.Name()) == ".json" {
				// JSONファイルの中で最新のものを選択
				if info.ModTime().After(latestTime) {
					latestTime = info.ModTime()
					latestFile = filepath.Join(absDir, file.Name())
				}
			}
		}
	}

	if latestFile == "" {
		errorMsg := "JSON file not found.\nSearched directories:\n"
		for _, dir := range searchDirs {
			errorMsg += fmt.Sprintf("- %s\n", dir)
		}
		if len(searchErrors) > 0 {
			errorMsg += "\nErrors during search:\n"
			for _, err := range searchErrors {
				errorMsg += fmt.Sprintf("- %s\n", err)
			}
		}
		return "", fmt.Errorf(errorMsg)
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
		fmt.Printf("Error: Unknown command '%s'\n", os.Args[1])
		showMainHelp()
		os.Exit(1)
	}
}
