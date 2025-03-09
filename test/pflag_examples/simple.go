package main

import (
	"fmt"

	"github.com/spf13/pflag"
)

func main() {
	// フラグの定義
	var (
		repoPath  string
		outputDir string
		debug     bool
		help      bool
	)

	// フラグの設定（グローバルフラグを使用）
	pflag.StringVarP(&repoPath, "repo", "r", "", "リポジトリのパスまたはURL")
	pflag.StringVarP(&outputDir, "output", "o", "output", "出力ディレクトリ")
	pflag.BoolVarP(&debug, "debug", "d", false, "デバッグモードを有効にする")
	pflag.BoolVarP(&help, "help", "h", false, "ヘルプを表示")

	// フラグの説明をカスタマイズ
	pflag.Usage = func() {
		fmt.Println("使用方法: simple [オプション]")
		fmt.Println("\nオプション:")
		pflag.PrintDefaults()
	}

	// フラグを解析
	pflag.Parse()

	// ヘルプフラグが指定された場合
	if help {
		pflag.Usage()
		return
	}

	// フラグの値を表示
	fmt.Printf("リポジトリパス: %s\n", repoPath)
	fmt.Printf("出力ディレクトリ: %s\n", outputDir)
	fmt.Printf("デバッグモード: %v\n", debug)
}
