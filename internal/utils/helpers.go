package utils

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// IsValidGitURL は指定された文字列が有効なGit URLかどうかをチェックする
func IsValidGitURL(repoURL string) bool {
	_, err := url.ParseRequestURI(repoURL)
	if err != nil {
		return false
	}

	// 一般的なGitホスティングサービスのURLパターンをチェック
	return strings.HasSuffix(repoURL, ".git") ||
		strings.Contains(repoURL, "github.com/") ||
		strings.Contains(repoURL, "gitlab.com/") ||
		strings.Contains(repoURL, "bitbucket.org/")
}

// IsLocalRepository は指定されたパスがローカルのGitリポジトリかどうかをチェックする
func IsLocalRepository(repoPath string) bool {
	// .gitディレクトリの存在をチェック
	gitDir := filepath.Join(repoPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// SanitizePath はファイルパスを安全なファイル名形式に変換する
func SanitizePath(path string) string {
	// 移動やコピーを示す特殊なパスを処理（例: 'old => new'）
	if strings.Contains(path, "=>") {
		parts := strings.Split(path, "=>")
		if len(parts) > 1 {
			// 最新のパスを使用（通常は右側）
			path = strings.TrimSpace(parts[len(parts)-1])
		}
	}

	// ディレクトリ部分と名前部分を取得
	dir := filepath.Dir(path)
	filename := filepath.Base(path)

	// ディレクトリパスを安全な形式に変換
	if dir != "" && dir != "." {
		// 非アルファベット数字記号を安全な形式に変換
		dir = regexp.MustCompile(`[^a-zA-Z0-9/\\]`).ReplaceAllString(dir, "_")

		// WindowsとUnixのパス区切り文字を統一
		dir = strings.ReplaceAll(dir, "/", string(os.PathSeparator))
		dir = strings.ReplaceAll(dir, "\\", string(os.PathSeparator))
	} else {
		dir = "."
	}

	// ファイル名から拡張子を分離
	ext := filepath.Ext(filename)
	filenameWithoutExt := strings.TrimSuffix(filename, ext)

	// ファイル名部分を安全な文字列に変換
	// ファイルシステムで問題となる文字を置換
	safeFilename := regexp.MustCompile(`[^a-zA-Z0-9-_]`).ReplaceAllString(filenameWithoutExt, "_")

	// 拡張子も安全な形式に変換
	safeExt := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(ext, "_")

	// 安全なファイル名を生成
	safeFilename = safeFilename + safeExt

	// 空のファイル名は避ける
	if safeFilename == "" {
		safeFilename = "unknown_file"
	}

	// 新しいパスを作成
	if dir == "." {
		return safeFilename
	}
	return filepath.Join(dir, safeFilename)
}

// GetFileExtension はファイルの拡張子を取得する
func GetFileExtension(filePath string) string {
	return strings.TrimPrefix(filepath.Ext(filePath), ".")
}

// EnsureDirectoryExists は指定されたディレクトリが存在することを確認し、存在しなければ作成する
func EnsureDirectoryExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// PrintProgressBar はコンソールに進捗バーを表示する
func PrintProgressBar(current, total int, prefix string) {
	width := 50
	percent := float64(current) / float64(total)
	filled := int(float64(width) * percent)
	empty := width - filled

	bar := strings.Repeat("=", filled) + strings.Repeat(" ", empty)
	fmt.Printf("\r%s [%s] %.1f%% (%d/%d)", prefix, bar, percent*100, current, total)

	if current == total {
		fmt.Println()
	}
}

// GetNumCPUs はシステムで利用可能なCPUコア数を返す
// 最低1、最大は物理コア数
func GetNumCPUs() int {
	numCPU := runtime.NumCPU()
	if numCPU < 1 {
		return 1
	}
	return numCPU
}
