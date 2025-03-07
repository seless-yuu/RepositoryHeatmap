package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/repositoryheatmap/internal/git"
	"github.com/repositoryheatmap/internal/heatmap"
	"github.com/repositoryheatmap/internal/utils"
)

var (
	repoPath    string
	outputDir   string
	outputType  string
	skipClone   bool
	showVersion bool
	showHelp    bool
	debugLog    bool
	version     = "0.1.0" // アプリケーションバージョン
	logFile     *os.File
	logger      *log.Logger
)

func init() {
	// コマンドライン引数の設定
	flag.StringVar(&repoPath, "repo", "", "解析するGitリポジトリのパスまたはURL（必須）")
	flag.StringVar(&outputDir, "output", "output", "出力ディレクトリ")
	flag.StringVar(&outputType, "format", "svg", "出力形式 (svg または webp)")
	flag.BoolVar(&skipClone, "skip-clone", false, "リポジトリが既にクローンされている場合はスキップ")
	flag.BoolVar(&showVersion, "version", false, "バージョン情報を表示")
	flag.BoolVar(&showHelp, "help", false, "ヘルプメッセージを表示")
	flag.BoolVar(&debugLog, "debug", false, "デバッグログをファイルに出力")

	// ヘルプメッセージのカスタマイズ
	flag.Usage = func() {
		fmt.Println("Repository Heatmap - Gitリポジトリの変更頻度解析ツール")
		fmt.Println("=========================================================")
		fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
		fmt.Println("GitリポジトリのヒートマップデータとJSON出力を生成します。\n")
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("\n例:")
		fmt.Printf("  %s --repo=/path/to/local/repo\n", os.Args[0])
		fmt.Printf("  %s --repo=https://github.com/username/repo.git --format=webp\n", os.Args[0])
		fmt.Printf("  %s --repo=/path/to/repo --output=./results\n", os.Args[0])
	}
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

func main() {
	// フラグのパース
	flag.Parse()

	// デバッグログの設定
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

	// コマンドを表示
	logDebug("実行コマンド: %s", os.Args[0])

	// ヘルプ表示
	if showHelp || len(os.Args) == 1 {
		logDebug("ヘルプを表示します")
		flag.Usage()
		return
	}

	// バージョン情報表示
	if showVersion {
		logDebug("Repository Heatmap v%s", version)
		return
	}

	// 必須パラメータチェック
	if repoPath == "" {
		logDebug("エラー: リポジトリパスを指定してください（--repo オプション）\n")
		flag.Usage()
		os.Exit(1)
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

	// リポジトリの絶対パスを取得
	var absRepoPath string
	var localRepoPath string

	logDebug("リポジトリヒートマップ解析を開始します...")
	logDebug("リポジトリ: %s", repoPath)
	logDebug("出力ディレクトリ: %s", absOutputDir)
	logDebug("出力形式: %s", outputType)

	// リポジトリの解析
	analyzer, err := git.NewAnalyzer(repoPath)
	if err != nil {
		logDebug("リポジトリアナライザの作成に失敗しました: %v", err)
		os.Exit(1)
	}

	// リポジトリのオープンまたはクローン
	if utils.IsLocalRepository(repoPath) {
		logDebug("ローカルリポジトリを開きます: %s", repoPath)
		if err := analyzer.Open(); err != nil {
			logDebug("リポジトリを開けませんでした: %v", err)
			os.Exit(1)
		}
		// ローカルリポジトリパスを設定
		absRepoPath, err = filepath.Abs(repoPath)
		if err != nil {
			logDebug("リポジトリパスの解決に失敗しました: %v", err)
			os.Exit(1)
		}
		localRepoPath = absRepoPath
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
		// クローンされたリポジトリパスを設定
		absRepoPath, err = filepath.Abs(tempDir)
		if err != nil {
			logDebug("クローンパスの解決に失敗しました: %v", err)
			os.Exit(1)
		}
		localRepoPath = absRepoPath
	} else {
		logDebug("無効なリポジトリパスまたはURL: %s", repoPath)
		os.Exit(1)
	}

	// リポジトリの解析
	logDebug("コミット履歴を解析中...")
	stats, err := analyzer.Analyze()
	if err != nil {
		logDebug("リポジトリの解析に失敗しました: %v", err)
		os.Exit(1)
	}

	// 行単位での変更追跡
	logDebug("行単位の変更を追跡中...")
	lineTracker := git.NewLineTracker(analyzer)
	if err := lineTracker.TrackLineChanges(stats); err != nil {
		logDebug("行単位の変更追跡に失敗しました: %v", err)
		os.Exit(1)
	}
	lineTracker.CalculateLineHeatLevels(stats)

	// ヒートマップデータの生成
	logDebug("ヒートマップデータを生成中...")
	generator := heatmap.NewGenerator(absOutputDir, stats)
	if err := generator.Generate(); err != nil {
		logDebug("ヒートマップデータの生成に失敗しました: %v", err)
		os.Exit(1)
	}

	// ヒートマップの可視化
	logDebug("ヒートマップを可視化中...")
	visualizer := heatmap.NewVisualizer(absOutputDir, outputType, stats)
	// リポジトリのパスを設定（ファイル内容の読み込みに必要）
	visualizer.SetRepoPath(localRepoPath)
	if err := visualizer.Visualize(); err != nil {
		logDebug("ヒートマップの可視化に失敗しました: %v", err)
		os.Exit(1)
	}

	logDebug("解析が完了しました")
	logDebug("結果は %s ディレクトリに保存されました", absOutputDir)
}
