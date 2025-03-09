package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

// 問題のあるフラグ設定を再現する関数
func problematicFlagSetup() {
	// サブコマンドフラグの設定
	cmd := pflag.NewFlagSet("cmd", pflag.ExitOnError)

	// フラグを直接設定する（変数に割り当てない）
	cmd.BoolP("help", "h", false, "ヘルプを表示")

	// 引数を解析
	cmd.Parse(os.Args[1:])

	// フラグの値を取得して変数に代入
	showHelp, _ := cmd.GetBool("help")

	// ヘルプフラグを確認
	if showHelp {
		fmt.Println("ヘルプを表示（正しい方法）")
	} else {
		// ここでは--helpを指定しても実行されない可能性がある
		fmt.Println("ヘルプフラグは設定されていません")
	}
}

// 正しいフラグ設定を示す関数
func correctFlagSetup() {
	// サブコマンドフラグの設定
	cmd := pflag.NewFlagSet("cmd", pflag.ExitOnError)

	// フラグを変数に割り当てる
	var showHelp bool
	cmd.BoolVarP(&showHelp, "help", "h", false, "ヘルプを表示")

	// 引数を解析
	cmd.Parse(os.Args[1:])

	// ヘルプフラグを確認（変数を直接使用）
	if showHelp {
		fmt.Println("ヘルプを表示（正しい方法）")
	} else {
		fmt.Println("ヘルプフラグは設定されていません")
	}
}

// もう一つの方法：解析前に明示的にヘルプフラグをチェック
func explicitHelpCheck() {
	cmd := pflag.NewFlagSet("cmd", pflag.ExitOnError)
	cmd.BoolP("help", "h", false, "ヘルプを表示")

	// 解析前に明示的にヘルプフラグをチェック
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			fmt.Println("ヘルプを表示（明示的なチェック）")
			return
		}
	}

	// 引数を解析
	cmd.Parse(os.Args[1:])

	// ここでは通常の処理を行う
	fmt.Println("通常の処理を実行")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: pflag_debug [problematic|correct|explicit]")
		return
	}

	switch os.Args[1] {
	case "problematic":
		os.Args = os.Args[1:] // 最初の引数を削除
		problematicFlagSetup()
	case "correct":
		os.Args = os.Args[1:] // 最初の引数を削除
		correctFlagSetup()
	case "explicit":
		os.Args = os.Args[1:] // 最初の引数を削除
		explicitHelpCheck()
	default:
		fmt.Println("不明なモード。problematic、correct、またはexplicitを指定してください。")
	}
}
