package git

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/repositoryheatmap/pkg/models"
)

// LineTracker は行単位での変更追跡を担当する構造体
type LineTracker struct {
	repo        *RepositoryAnalyzer
	workerCount int
	mutex       sync.Mutex
}

// NewLineTracker は新しいLineTrackerインスタンスを作成する
func NewLineTracker(repo *RepositoryAnalyzer) *LineTracker {
	return &LineTracker{
		repo:        repo,
		workerCount: 1, // デフォルトは1（シングルスレッド）
	}
}

// SetWorkerCount は並列処理に使用するワーカー数を設定する
func (lt *LineTracker) SetWorkerCount(count int) {
	if count < 1 {
		count = 1
	}
	lt.workerCount = count
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

	// 1. すべてのコミットを収集
	var commitList []*object.Commit
	err = commits.ForEach(func(c *object.Commit) error {
		// 日付フィルタが設定されている場合は日付をチェック
		if lt.repo.sinceDate != nil && c.Committer.When.Before(*lt.repo.sinceDate) {
			return nil // 指定日付より前のコミットはスキップ
		}
		commitList = append(commitList, c)
		return nil
	})

	if err != nil {
		return fmt.Errorf("コミット履歴の取得に失敗しました: %w", err)
	}

	// コミットが2つ未満の場合は差分を計算できないのでスキップ
	if len(commitList) < 2 {
		return nil
	}

	// 2. マルチスレッド処理でコミットペアを解析
	if lt.workerCount > 1 && len(commitList) > 1 {
		// ワーカー数を調整（コミット数より多く設定しないように）
		workerCount := lt.workerCount
		if workerCount > len(commitList)-1 {
			workerCount = len(commitList) - 1
		}

		// コミットペアをワーカー数で分割
		pairCount := len(commitList) - 1
		pairsPerWorker := pairCount / workerCount
		if pairsPerWorker < 1 {
			pairsPerWorker = 1
		}

		// ワーカーを起動
		var wg sync.WaitGroup
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			startIdx := i * pairsPerWorker
			endIdx := (i + 1) * pairsPerWorker
			if i == workerCount-1 {
				endIdx = pairCount
			}

			go func(start, end int) {
				defer wg.Done()
				for j := start; j < end; j++ {
					lt.processCommitPair(commitList[j], commitList[j+1], stats)
				}
			}(startIdx, endIdx)
		}

		// すべてのワーカーの完了を待機
		wg.Wait()
	} else {
		// シングルスレッド処理
		for i := 0; i < len(commitList)-1; i++ {
			lt.processCommitPair(commitList[i], commitList[i+1], stats)
		}
	}

	return nil
}

var chunkHeaderRegex = regexp.MustCompile(`@@ -(\d+),?(\d*) \+(\d+),?(\d*) @@`)

// processCommitPair は2つのコミット間の差分を処理する
func (lt *LineTracker) processCommitPair(current, next *object.Commit, stats *models.RepositoryStats) {
	// 2つのツリー間の差分を取得
	patch, err := current.Patch(next)
	if err != nil {
		// エラーがあっても続行
		return
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

		// 共有リソースへのアクセスはロックする
		lt.mutex.Lock()

		// ファイル情報を取得
		fileInfo, exists := stats.Files[filePath]
		if !exists {
			// 解析中に見つからないファイルはスキップ
			lt.mutex.Unlock()
			continue
		}

		// 各チャンクの変更を処理
		currentLine := 1
		for _, chunk := range filePatch.Chunks() {
			// チャンクの内容を行に分割
			lines := strings.Split(chunk.Content(), "\n")
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1] // 最後の空行を除外
			}

			// チャンクのタイプに基づいて処理
			switch chunk.Type() {
			case diff.Add:
				// 追加された行を処理
				for _, line := range lines {
					if line == "" {
						continue
					}
					fileInfo.LineChanges[currentLine]++
					currentLine++
				}
			case diff.Delete:
				// 削除された行を処理
				for _, line := range lines {
					if line == "" {
						continue
					}
					fileInfo.LineChanges[currentLine]++
					currentLine++
				}
			default:
				// 変更されていない行を処理
				for range lines {
					currentLine++
				}
			}
		}

		// 更新された情報をマップに戻す
		stats.Files[filePath] = fileInfo

		// ロックを解放
		lt.mutex.Unlock()
	}
}

// CalculateLineHeatLevels は各ファイルの行ごとのヒートレベルを計算する
func (lt *LineTracker) CalculateLineHeatLevels(stats *models.RepositoryStats) {
	// 並列処理用のチャネル
	type fileJob struct {
		path string
		info models.FileChangeInfo
	}

	// 1. マルチスレッド処理を使用する場合
	if lt.workerCount > 1 && len(stats.Files) > 1 {
		// ワーカー数を調整（ファイル数より多く設定しないように）
		workerCount := lt.workerCount
		if workerCount > len(stats.Files) {
			workerCount = len(stats.Files)
		}

		// ジョブチャネルを作成
		jobs := make(chan fileJob, len(stats.Files))
		results := make(chan fileJob, len(stats.Files))

		// ワーカーを起動
		var wg sync.WaitGroup
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					// そのファイル内の最大変更回数を見つける
					maxLineChanges := 0
					for _, changes := range job.info.LineChanges {
						if changes > maxLineChanges {
							maxLineChanges = changes
						}
					}

					// 変更回数に基づいて各行のヒートレベルを計算
					if maxLineChanges > 0 {
						for lineNum, changes := range job.info.LineChanges {
							// ヒートレベルを0.0-1.0の範囲に正規化
							job.info.LineChanges[lineNum] = int(float64(changes) / float64(maxLineChanges) * 10)
						}
					}

					// 結果を送信
					results <- job
				}
			}()
		}

		// ジョブを送信
		for path, info := range stats.Files {
			jobs <- fileJob{path: path, info: info}
		}
		close(jobs)

		// すべてのワーカーの終了を待機
		go func() {
			wg.Wait()
			close(results)
		}()

		// 結果を収集
		for job := range results {
			stats.Files[job.path] = job.info
		}
	} else {
		// 2. シングルスレッド処理
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
}
