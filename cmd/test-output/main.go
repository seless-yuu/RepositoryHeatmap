package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("これはテスト出力です")
	fmt.Fprintf(os.Stdout, "標準出力へのテスト\n")
	fmt.Fprintf(os.Stderr, "標準エラー出力へのテスト\n")

	// 手動フラッシュを試みる（通常は不要）
	os.Stdout.Sync()
	os.Stderr.Sync()
}
