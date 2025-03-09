package utils

import (
	"bufio"
	"fmt"
	"os"
)

// ReadFileLines はファイルを行単位で読み込む
func ReadFileLines(file *os.File) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ファイル読み込み中にエラーが発生しました: %w", err)
	}

	return lines, nil
}
