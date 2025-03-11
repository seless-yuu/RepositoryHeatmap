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
	// NULLパスのチェック
	if path == "" {
		return "unknown_file"
	}

	// パス区切り文字を '/' に正規化
	normalizedPath := filepath.ToSlash(path)

	// 安全に処理するために、ディレクトリとファイル名を分ける
	dir, fileName := filepath.Split(normalizedPath)

	// ファイル名を単純化（最初に特定の特殊ファイルを確認）
	isSpecialFile := false

	// READMEファイルの特別処理
	if strings.HasPrefix(strings.ToLower(fileName), "readme") {
		isSpecialFile = true
	}

	// Pythonファイルの特別処理
	if strings.HasSuffix(strings.ToLower(fileName), ".py") && strings.Contains(fileName, "_") {
		isSpecialFile = true
	}

	// 特別処理が必要なファイル名
	var safeFileName string

	if isSpecialFile {
		// 特別なファイルはアンダースコアと拡張子を保持
		// ただし、ファイルシステムやURLで問題になりうる文字だけを置換
		safeFileName = replaceUnsafeChars(fileName)
	} else {
		// 一般的なファイル名の処理
		ext := filepath.Ext(fileName)
		fileNameWithoutExt := strings.TrimSuffix(fileName, ext)

		// ファイル名だけアルファベット、数字、ハイフン、アンダースコアに制限
		safeBase := replaceUnsafeChars(fileNameWithoutExt)

		// 拡張子があれば保持（先頭の「.」を含む）
		if ext != "" {
			// 拡張子の始めのドットは維持し、残りの部分を処理
			safeExt := "." + replaceUnsafeChars(ext[1:])
			safeFileName = safeBase + safeExt
		} else {
			safeFileName = safeBase
		}
	}

	// ディレクトリパスを安全に変換（ディレクトリ構造を維持）
	var safeDirPath string
	if dir != "" {
		// パスセパレータ（/）で分割
		dirParts := strings.Split(strings.Trim(dir, "/"), "/")
		safeDirParts := make([]string, len(dirParts))

		for i, part := range dirParts {
			// 各ディレクトリ名を安全に変換
			safeDirParts[i] = replaceUnsafeChars(part)
		}

		// パスを再構築
		safeDirPath = strings.Join(safeDirParts, "/")
		if safeDirPath != "" {
			safeDirPath += "/"
		}
	}

	// 最終的なパスを構築
	return safeDirPath + safeFileName
}

// replaceUnsafeChars は安全でない文字を置き換える
func replaceUnsafeChars(s string) string {
	// URLやファイルシステム上で問題になる特殊文字のみを置換
	// アンダースコア、ハイフン、ドットは保持
	return regexp.MustCompile(`[^a-zA-Z0-9_\-.]`).ReplaceAllString(s, "_")
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
