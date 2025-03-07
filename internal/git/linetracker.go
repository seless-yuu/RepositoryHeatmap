package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/repositoryheatmap/pkg/models"
)

// LineTracker は行単位での変更追跡を担当する構造体
type LineTracker struct {
	repo *RepositoryAnalyzer
}

// NewLineTracker は新しいLineTrackerインスタンスを作成する
func NewLineTracker(repo *RepositoryAnalyzer) *LineTracker {
	return &LineTracker{
		repo: repo,
	}
}

// TrackLineChanges はリポジトリ内のファイルの行単位での変更を追跡する
func (lt *LineTracker) TrackLineChanges(stats *models.RepositoryStats) error {
	if lt.repo == nil || lt.repo.repo == nil {
		return fmt.Errorf("リポジトリが開かれていません")
	}

	// HEADの参照を取得
	ref, err := lt.repo.repo.Head()
	if err != nil {
		return fmt.Errorf("HEAD参照を取得できませんでした: %w", err)
	}

	// コミット履歴を取得
	commits, err := lt.repo.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return fmt.Errorf("コミット履歴を取得できませんでした: %w", err)
	}

	// 前のコミットを保持する変数
	var prevCommit *object.Commit

	// 各コミットを処理
	err = commits.ForEach(func(c *object.Commit) error {
		// 最初のコミットはスキップ（比較対象がないため）
		if prevCommit == nil {
			prevCommit = c
			return nil
		}

		// 2つのツリー間の差分を取得
		patch, err := prevCommit.Patch(c)
		if err != nil {
			return fmt.Errorf("パッチの作成に失敗しました: %w", err)
		}

		// 各ファイルのパッチを処理
		for _, filePatch := range patch.FilePatches() {
			from, to := filePatch.Files()
			var filePath string

			// ファイルパスを取得（削除されたファイルの場合はfromから、追加/変更されたファイルの場合はtoから）
			if from == nil {
				if to == nil {
					continue // ファイルが存在しない場合はスキップ
				}
				filePath = to.Path()
			} else {
				filePath = from.Path()
			}

			// ファイル情報を取得
			fileInfo, exists := stats.Files[filePath]
			if !exists {
				// 解析中に見つからないファイルはスキップ
				continue
			}

			// 各チャンクの変更を処理
			for _, chunk := range filePatch.Chunks() {
				lines := chunk.Lines()
				lineNum := chunk.Content()[0] // 現実の実装では正確な行番号計算が必要

				for i, line := range lines {
					if line.Type == chunk.Type() {
						lineNumber := lineNum + i
						// 行の変更回数を増やす
						fileInfo.LineChanges[lineNumber]++
					}
				}
			}

			// 更新された情報をマップに戻す
			stats.Files[filePath] = fileInfo
		}

		// 現在のコミットを前のコミットとして保存
		prevCommit = c
		return nil
	})

	if err != nil {
		return fmt.Errorf("行単位の変更追跡中にエラーが発生しました: %w", err)
	}

	return nil
}

// CalculateLineHeatLevels は各ファイルの行ごとのヒートレベルを計算する
func (lt *LineTracker) CalculateLineHeatLevels(stats *models.RepositoryStats) {
	// 各ファイルに対して処理
	for filePath, fileInfo := range stats.Files {
		// そのファイル内の最大変更回数を見つける
		maxLineChanges := 0
		for _, changes := range fileInfo.LineChanges {
			if changes > maxLineChanges {
				maxLineChanges = changes
			}
		}

		// 変更回数に基づいて各行のヒートレベルを計算
		if maxLineChanges > 0 {
			for lineNum, changes := range fileInfo.LineChanges {
				// ヒートレベルを0.0-1.0の範囲に正規化
				fileInfo.LineChanges[lineNum] = int(float64(changes) / float64(maxLineChanges) * 10)
			}
		}

		// 更新された情報をマップに戻す
		stats.Files[filePath] = fileInfo
	}
}
