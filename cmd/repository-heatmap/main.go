package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/repositoryheatmap/internal/git"
	"github.com/repositoryheatmap/internal/heatmap"
	"github.com/repositoryheatmap/internal/utils"
	"github.com/repositoryheatmap/pkg/models"
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
)

// コマンドラインフラグの設定
func setupCommonFlags() {
	flag.StringVar(&outputDir, "output", "output", "出力ディレクトリ")
	flag.BoolVar(&showVersion, "version", false, "バージョン情報を表示")
	flag.BoolVar(&showHelp, "help", false, "ヘルプメッセージを表示")
	flag.BoolVar(&debugLog, "debug", false, "デバッグログをファイルに出力")
}

// analyzeコマンド用のフラグ設定
func setupAnalyzeFlags(analyzeCmd *flag.FlagSet) {
	analyzeCmd.StringVar(&repoPath, "repo", "", "解析するGitリポジトリのパスまたはURL（必須）")
	analyzeCmd.StringVar(&outputDir, "output", "output", "出力ディレクトリ")
	analyzeCmd.BoolVar(&skipClone, "skip-clone", false, "リポジトリが既にクローンされている場合はスキップ")
	analyzeCmd.StringVar(&sinceDate, "since", "", "指定された日付以降のコミットのみを解析 (YYYY-MM-DD形式)")
	analyzeCmd.IntVar(&workers, "workers", 1, "並列処理に使用するワーカー数（0=CPUコア数に自動設定）")
	analyzeCmd.BoolVar(&debugLog, "debug", false, "デバッグログをファイルに出力")
}

// visualizeコマンド用のフラグ設定
func setupVisualizeFlags(visualizeCmd *flag.FlagSet) {
	visualizeCmd.StringVar(&outputDir, "output", "output", "JSONデータとヒートマップ画像の出力ディレクトリ")
	visualizeCmd.StringVar(&outputType, "format", "svg", "出力形式 (svg または webp)")
	visualizeCmd.StringVar(&repoPath, "repo", "", "ファイル内容表示のためのリポジトリパス（必須ではない）")
	visualizeCmd.IntVar(&maxFiles, "max-files", 100, "ヒートマップに表示する最大ファイル数")
	visualizeCmd.BoolVar(&debugLog, "debug", false, "デバッグログをファイルに出力")
}

// analyzeコマンドのヘルプメッセージ
func showAnalyzeHelp(analyzeCmd *flag.FlagSet) {
	fmt.Println("Repository Heatmap - リポジトリ解析コマンド")
	fmt.Println("=========================================================")
	fmt.Printf("Usage: %s analyze [options]\n\n", os.Args[0])
	fmt.Println("GitリポジトリのコミットデータとJSON出力を生成します。\n")
	fmt.Println("Options:")
	analyzeCmd.PrintDefaults()
	fmt.Println("\n例:")
	fmt.Printf("  %s analyze --repo=/path/to/local/repo\n", os.Args[0])
	fmt.Printf("  %s analyze --repo=https://github.com/username/repo.git\n", os.Args[0])
	fmt.Printf("  %s analyze --repo=/path/to/repo --output=./results --workers=4\n", os.Args[0])
	fmt.Printf("  %s analyze --repo=/path/to/repo --since=2023-02-01\n", os.Args[0])
}

// visualizeコマンドのヘルプメッセージ
func showVisualizeHelp(visualizeCmd *flag.FlagSet) {
	fmt.Println("Repository Heatmap - ヒートマップ可視化コマンド")
	fmt.Println("=========================================================")
	fmt.Printf("Usage: %s visualize [options]\n\n", os.Args[0])
	fmt.Println("解析済みJSON出力からヒートマップを生成します。\n")
	fmt.Println("Options:")
	visualizeCmd.PrintDefaults()
	fmt.Println("\n例:")
	fmt.Printf("  %s visualize --output=./results\n", os.Args[0])
	fmt.Printf("  %s visualize --output=./results --format=webp\n", os.Args[0])
	fmt.Printf("  %s visualize --output=./results --repo=/path/to/repo\n", os.Args[0])
	fmt.Printf("  %s visualize --output=./results --max-files=200\n", os.Args[0])
}

// 全体のヘルプメッセージ
func showMainHelp() {
	fmt.Println("Repository Heatmap - Gitリポジトリの変更頻度解析ツール")
	fmt.Println("=========================================================")
	fmt.Printf("Usage: %s <command> [options]\n\n", os.Args[0])
	fmt.Println("Commands:")
	fmt.Println("  analyze    リポジトリを解析しJSONデータを生成")
	fmt.Println("  visualize  JSONデータからヒートマップを生成")
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
	analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
	setupAnalyzeFlags(analyzeCmd)

	// ヘルプメッセージのカスタマイズ
	analyzeCmd.Usage = func() {
		showAnalyzeHelp(analyzeCmd)
	}

	analyzeCmd.Parse(args)

	// デバッグログの設定
	setupDebugLog()

	// コマンドを表示
	logDebug("実行コマンド: %s analyze", os.Args[0])

	// ヘルプ表示
	if showHelp || len(args) == 0 {
		showAnalyzeHelp(analyzeCmd)
		return
	}

	// 必須パラメータチェック
	if repoPath == "" {
		logDebug("エラー: リポジトリパスを指定してください（--repo オプション）\n")
		showAnalyzeHelp(analyzeCmd)
		os.Exit(1)
	}

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
	jsonFilePath := filepath.Join(absOutputDir, fmt.Sprintf("%s-heatmap.json", repoName))

	// ヒートマップデータのJSONへの保存
	logDebug("ヒートマップデータを生成中...")
	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		logDebug("ヒートマップデータのJSON変換に失敗しました: %v", err)
		os.Exit(1)
	}

	if err := ioutil.WriteFile(jsonFilePath, jsonData, 0644); err != nil {
		logDebug("ヒートマップデータのJSON保存に失敗しました: %v", err)
		os.Exit(1)
	}

	logDebug("ヒートマップデータをJSONに保存しました: %s", jsonFilePath)
	logDebug("解析が完了しました")
}

// visualizeCommand はJSON出力からヒートマップを生成する
func visualizeCommand(args []string) {
	visualizeCmd := flag.NewFlagSet("visualize", flag.ExitOnError)
	setupVisualizeFlags(visualizeCmd)

	// ヘルプメッセージのカスタマイズ
	visualizeCmd.Usage = func() {
		showVisualizeHelp(visualizeCmd)
	}

	visualizeCmd.Parse(args)

	// デバッグログの設定
	setupDebugLog()

	// コマンドを表示
	logDebug("実行コマンド: %s visualize", os.Args[0])

	// ヘルプ表示
	if showHelp || len(args) == 0 {
		showVisualizeHelp(visualizeCmd)
		return
	}

	// 出力ディレクトリのチェック
	if outputDir == "" {
		logDebug("エラー: 出力ディレクトリを指定してください（--output オプション）\n")
		showVisualizeHelp(visualizeCmd)
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

	// 出力ディレクトリにJSONファイルがあるか検索
	files, err := ioutil.ReadDir(absOutputDir)
	if err != nil {
		logDebug("出力ディレクトリの読み取りに失敗しました: %v", err)
		os.Exit(1)
	}

	// 最新のJSONファイルを探す
	var jsonFilePath string
	var latestTime time.Time

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			// JSONファイルの中で最新のものを選択
			if file.ModTime().After(latestTime) {
				latestTime = file.ModTime()
				jsonFilePath = filepath.Join(absOutputDir, file.Name())
			}
		}
	}

	if jsonFilePath == "" {
		logDebug("エラー: 出力ディレクトリにJSONファイルが見つかりませんでした")
		os.Exit(1)
	}

	logDebug("JSONファイルを読み込みます: %s", jsonFilePath)

	// JSONファイルからデータを読み込む
	jsonData, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		logDebug("JSONファイルの読み込みに失敗しました: %v", err)
		os.Exit(1)
	}

	var stats models.RepositoryStats
	if err := json.Unmarshal(jsonData, &stats); err != nil {
		logDebug("JSONデータのパースに失敗しました: %v", err)
		os.Exit(1)
	}

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
	logDebug("出力形式: %s", outputType)

	// ヒートマップの可視化
	visualizer := heatmap.NewVisualizer(absOutputDir, outputType, &stats)

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

func main() {
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
	case "--version":
		fmt.Printf("Repository Heatmap v%s\n", version)
	case "--help", "-h":
		showMainHelp()
	default:
		fmt.Printf("エラー: 不明なコマンド '%s'\n", os.Args[1])
		showMainHelp()
		os.Exit(1)
	}
}
