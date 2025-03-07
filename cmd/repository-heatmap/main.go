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
	version     = "0.1.0" // アプリケーションバージョン
)

func init() {
	// コマンドライン引数の設定
	flag.StringVar(&repoPath, "repo", "", "解析するGitリポジトリのパスまたはURL（必須）")
	flag.StringVar(&outputDir, "output", "output", "出力ディレクトリ")
	flag.StringVar(&outputType, "format", "svg", "出力形式 (svg または webp)")
	flag.BoolVar(&skipClone, "skip-clone", false, "リポジトリが既にクローンされている場合はスキップ")
	flag.BoolVar(&showVersion, "version", false, "バージョン情報を表示")
}

func main() {
	flag.Parse()

	// バージョン情報表示
	if showVersion {
		fmt.Printf("Repository Heatmap v%s\n", version)
		os.Exit(0)
	}

	// 必須パラメータチェック
	if repoPath == "" {
		log.Fatal("リポジトリパスを指定してください（--repo オプション）")
	}

	// 出力ディレクトリ作成
	if err := utils.EnsureDirectoryExists(outputDir); err != nil {
		log.Fatalf("出力ディレクトリの作成に失敗しました: %v", err)
	}

	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		log.Fatalf("出力パスの解決に失敗しました: %v", err)
	}

	fmt.Printf("リポジトリヒートマップ解析を開始します...\n")
	fmt.Printf("リポジトリ: %s\n", repoPath)
	fmt.Printf("出力ディレクトリ: %s\n", absOutputDir)
	fmt.Printf("出力形式: %s\n", outputType)

	// リポジトリの解析
	analyzer, err := git.NewAnalyzer(repoPath)
	if err != nil {
		log.Fatalf("リポジトリアナライザの作成に失敗しました: %v", err)
	}

	// リポジトリのオープンまたはクローン
	var isLocalRepo bool
	if utils.IsLocalRepository(repoPath) {
		isLocalRepo = true
		fmt.Printf("ローカルリポジトリを開きます: %s\n", repoPath)
		if err := analyzer.Open(); err != nil {
			log.Fatalf("リポジトリを開けませんでした: %v", err)
		}
	} else if utils.IsValidGitURL(repoPath) {
		isLocalRepo = false
		// クローン用の一時ディレクトリを作成
		tempDir := filepath.Join(outputDir, "repo-clone")
		if err := utils.EnsureDirectoryExists(tempDir); err != nil {
			log.Fatalf("クローン用ディレクトリの作成に失敗しました: %v", err)
		}

		fmt.Printf("リポジトリをクローンします: %s -> %s\n", repoPath, tempDir)
		if err := analyzer.Clone(repoPath); err != nil {
			log.Fatalf("リポジトリのクローンに失敗しました: %v", err)
		}
	} else {
		log.Fatalf("無効なリポジトリパスまたはURL: %s", repoPath)
	}

	// リポジトリの解析
	fmt.Println("コミット履歴を解析中...")
	stats, err := analyzer.Analyze()
	if err != nil {
		log.Fatalf("リポジトリの解析に失敗しました: %v", err)
	}

	// 行単位での変更追跡
	fmt.Println("行単位の変更を追跡中...")
	lineTracker := git.NewLineTracker(analyzer)
	if err := lineTracker.TrackLineChanges(stats); err != nil {
		log.Fatalf("行単位の変更追跡に失敗しました: %v", err)
	}
	lineTracker.CalculateLineHeatLevels(stats)

	// ヒートマップデータの生成
	fmt.Println("ヒートマップデータを生成中...")
	generator := heatmap.NewGenerator(absOutputDir, stats)
	if err := generator.Generate(); err != nil {
		log.Fatalf("ヒートマップデータの生成に失敗しました: %v", err)
	}

	// ヒートマップの可視化
	fmt.Println("ヒートマップを可視化中...")
	visualizer := heatmap.NewVisualizer(absOutputDir, outputType, stats)
	if err := visualizer.Visualize(); err != nil {
		log.Fatalf("ヒートマップの可視化に失敗しました: %v", err)
	}

	fmt.Println("解析が完了しました")
	fmt.Printf("結果は %s ディレクトリに保存されました\n", absOutputDir)
}
