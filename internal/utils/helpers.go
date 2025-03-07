package utils

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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

// SanitizePath はファイルパスを安全な形式に変換する
func SanitizePath(path string) string {
	// ファイルシステムで問題となる文字を置換
	sanitized := strings.ReplaceAll(path, ":", "_")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	sanitized = strings.ReplaceAll(sanitized, "*", "_")
	sanitized = strings.ReplaceAll(sanitized, "?", "_")
	sanitized = strings.ReplaceAll(sanitized, "\"", "_")
	sanitized = strings.ReplaceAll(sanitized, "<", "_")
	sanitized = strings.ReplaceAll(sanitized, ">", "_")
	sanitized = strings.ReplaceAll(sanitized, "|", "_")
	return sanitized
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
