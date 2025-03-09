package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	// グローバルフラグの設定
	globalFlags := pflag.NewFlagSet("global", pflag.ExitOnError)
	var (
		showVersion bool
		showHelp    bool
	)
	globalFlags.BoolVarP(&showVersion, "version", "v", false, "バージョン情報を表示")
	globalFlags.BoolVarP(&showHelp, "help", "h", false, "ヘルプを表示")

	// サブコマンドがない場合やヘルプフラグが指定された場合
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		showMainHelp()
		return
	}

	// バージョン表示
	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Println("サブコマンド例 v1.0.0")
		return
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

// analyzeコマンドの処理
func analyzeCommand(args []string) {
	// フラグセットの作成
	analyzeCmd := pflag.NewFlagSet("analyze", pflag.ExitOnError)

	// フラグの定義
	var (
		repoPath  string
		outputDir string
		debug     bool
		help      bool
	)

	// フラグの設定
	analyzeCmd.StringVarP(&repoPath, "repo", "r", "", "リポジトリのパスまたはURL")
	analyzeCmd.StringVarP(&outputDir, "output", "o", "output", "出力ディレクトリ")
	analyzeCmd.BoolVarP(&debug, "debug", "d", false, "デバッグモードを有効にする")
	analyzeCmd.BoolVarP(&help, "help", "h", false, "ヘルプを表示")

	// ヘルプメッセージのカスタマイズ
	analyzeCmd.Usage = func() {
		fmt.Println("使用方法: subcommand analyze [オプション]")
		fmt.Println("\nオプション:")
		analyzeCmd.PrintDefaults()
	}

	// 特殊なケース：明示的に-hや--helpが指定された場合に対応
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			analyzeCmd.Usage()
			return
		}
	}

	// フラグを解析
	analyzeCmd.Parse(args)

	// ヘルプフラグが設定された場合
	if help {
		analyzeCmd.Usage()
		return
	}

	// 実際の処理
	fmt.Println("analyzeコマンドを実行:")
	fmt.Printf("  リポジトリパス: %s\n", repoPath)
	fmt.Printf("  出力ディレクトリ: %s\n", outputDir)
	fmt.Printf("  デバッグモード: %v\n", debug)
}

// visualizeコマンドの処理
func visualizeCommand(args []string) {
	// フラグセットの作成
	visualizeCmd := pflag.NewFlagSet("visualize", pflag.ExitOnError)

	// フラグの定義
	var (
		inputFile string
		outputDir string
		debug     bool
		help      bool
	)

	// フラグの設定
	visualizeCmd.StringVarP(&inputFile, "input", "i", "", "入力ファイルのパス")
	visualizeCmd.StringVarP(&outputDir, "output", "o", "output", "出力ディレクトリ")
	visualizeCmd.BoolVarP(&debug, "debug", "d", false, "デバッグモードを有効にする")
	visualizeCmd.BoolVarP(&help, "help", "h", false, "ヘルプを表示")

	// ヘルプメッセージのカスタマイズ
	visualizeCmd.Usage = func() {
		fmt.Println("使用方法: subcommand visualize [オプション]")
		fmt.Println("\nオプション:")
		visualizeCmd.PrintDefaults()
	}

	// 特殊なケース：明示的に-hや--helpが指定された場合に対応
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			visualizeCmd.Usage()
			return
		}
	}

	// フラグを解析
	visualizeCmd.Parse(args)

	// ヘルプフラグが設定された場合
	if help {
		visualizeCmd.Usage()
		return
	}

	// 実際の処理
	fmt.Println("visualizeコマンドを実行:")
	fmt.Printf("  入力ファイル: %s\n", inputFile)
	fmt.Printf("  出力ディレクトリ: %s\n", outputDir)
	fmt.Printf("  デバッグモード: %v\n", debug)
}

// メインヘルプの表示
func showMainHelp() {
	fmt.Println("使用方法: subcommand <command> [オプション]")
	fmt.Println("\nコマンド:")
	fmt.Println("  analyze    リポジトリを解析")
	fmt.Println("  visualize  データを可視化")
	fmt.Println("\nグローバルオプション:")
	fmt.Println("  -h, --help     ヘルプを表示")
	fmt.Println("  -v, --version  バージョン情報を表示")
	fmt.Println("\n各コマンドのヘルプを表示するには:")
	fmt.Println("  subcommand analyze --help")
	fmt.Println("  subcommand visualize --help")
}
