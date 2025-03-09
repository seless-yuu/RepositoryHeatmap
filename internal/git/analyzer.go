package git

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/repositoryheatmap/pkg/models"
)

// RepositoryAnalyzer はGitリポジトリ解析を担当する構造体
type RepositoryAnalyzer struct {
	repoPath    string
	repo        *git.Repository
	stats       *models.RepositoryStats
	sinceDate   *time.Time
	workerCount int
	mutex       sync.Mutex
	filePattern string // ファイルパターン（ワイルドカード）
	fileType    string // ファイルタイプ（拡張子）
}

// NewAnalyzer は新しいRepositoryAnalyzerインスタンスを作成する
func NewAnalyzer(repoPath string) (*RepositoryAnalyzer, error) {
	stats := &models.RepositoryStats{
		AnalyzedAt: time.Now(),
		Files:      make(map[string]models.FileChangeInfo),
	}

	return &RepositoryAnalyzer{
		repoPath:    repoPath,
		stats:       stats,
		workerCount: 1, // デフォルトは1（シングルスレッド）
	}, nil
}

// SetSinceDate は指定した日付以降のコミットのみを解析対象に設定する
func (a *RepositoryAnalyzer) SetSinceDate(since time.Time) {
	a.sinceDate = &since
}

// SetWorkerCount は並列処理に使用するワーカー数を設定する
func (a *RepositoryAnalyzer) SetWorkerCount(count int) {
	if count < 1 {
		count = 1
	}
	a.workerCount = count
}

// SetFilePattern はファイルパターン（ワイルドカード）を設定する
func (a *RepositoryAnalyzer) SetFilePattern(pattern string) {
	a.filePattern = pattern
}

// SetFileType はファイルタイプ（拡張子）を設定する
func (a *RepositoryAnalyzer) SetFileType(fileType string) {
	a.fileType = fileType
}

// Open はリポジトリを開く
func (a *RepositoryAnalyzer) Open() error {
	repo, err := git.PlainOpen(a.repoPath)
	if err != nil {
		return fmt.Errorf("リポジトリを開けませんでした: %w", err)
	}
	a.repo = repo

	// リポジトリ名を設定
	a.stats.RepositoryName = filepath.Base(a.repoPath)
	return nil
}

// Clone はリポジトリをクローンする
func (a *RepositoryAnalyzer) Clone(url string) error {
	repo, err := git.PlainClone(a.repoPath, false, &git.CloneOptions{
		URL:      url,
		Progress: nil,
	})
	if err != nil {
		return fmt.Errorf("リポジトリをクローンできませんでした: %w", err)
	}
	a.repo = repo

	// リポジトリ名を設定
	a.stats.RepositoryName = filepath.Base(a.repoPath)
	return nil
}

// Analyze はリポジトリを解析し、統計情報を収集する
func (a *RepositoryAnalyzer) Analyze() (*models.RepositoryStats, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("リポジトリが開かれていません")
	}

	// コミット履歴を取得
	ref, err := a.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("HEAD参照を取得できませんでした: %w", err)
	}

	commits, err := a.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("コミット履歴を取得できませんでした: %w", err)
	}

	// 1. すべてのコミットを収集
	var commitList []*object.Commit
	err = commits.ForEach(func(c *object.Commit) error {
		// 日付フィルタが設定されている場合は日付をチェック
		if a.sinceDate != nil && c.Committer.When.Before(*a.sinceDate) {
			return nil // 指定日付より前のコミットはスキップ
		}
		commitList = append(commitList, c)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("コミット履歴の取得に失敗しました: %w", err)
	}

	// 最初と最後のコミットを特定
	var firstCommit, lastCommit *object.Commit
	if len(commitList) > 0 {
		firstCommit = commitList[len(commitList)-1] // 最も古いコミット
		lastCommit = commitList[0]                  // 最新のコミット
	}

	// 著者情報の共有マップ
	authors := make(map[string]models.AuthorStat)

	// 2. マルチスレッド処理でコミットを解析
	if a.workerCount > 1 && len(commitList) > 0 {
		// ワーカー数を調整（コミット数より多く設定しないように）
		workerCount := a.workerCount
		if workerCount > len(commitList) {
			workerCount = len(commitList)
		}

		// コミットをワーカー数に応じて分割
		chunkSize := (len(commitList) + workerCount - 1) / workerCount
		var wg sync.WaitGroup

		// 各ワーカーが担当するコミット範囲を処理
		for i := 0; i < workerCount; i++ {
			wg.Add(1)

			// 各ワーカーの処理範囲を計算
			startIdx := i * chunkSize
			endIdx := (i + 1) * chunkSize
			if endIdx > len(commitList) {
				endIdx = len(commitList)
			}

			// ゴルーチンでコミット解析を実行
			go func(commits []*object.Commit) {
				defer wg.Done()
				a.processCommits(commits, authors)
			}(commitList[startIdx:endIdx])
		}

		// すべてのワーカーの完了を待機
		wg.Wait()
	} else {
		// シングルスレッドで処理
		a.processCommits(commitList, authors)
	}

	// コミット数を設定
	a.stats.CommitCount = len(commitList)

	// 著者数を設定
	a.stats.AuthorCount = len(authors)

	// 最初と最後のコミット日時を設定
	if firstCommit != nil {
		a.stats.FirstCommitAt = firstCommit.Committer.When
	}
	if lastCommit != nil {
		a.stats.LastCommitAt = lastCommit.Committer.When
	}

	// TOPの著者を取得
	for _, stat := range authors {
		a.stats.TopAuthors = append(a.stats.TopAuthors, stat)
		// 実際の実装ではソートが必要
	}

	// ヒートレベルを計算
	a.calculateHeatLevels()

	return a.stats, nil
}

// processCommits は指定されたコミットのリストを処理する
// （並列処理用に抽出されたメソッド）
func (a *RepositoryAnalyzer) processCommits(commits []*object.Commit, authors map[string]models.AuthorStat) {
	for _, c := range commits {
		// 著者情報を更新
		authorName := c.Author.Name

		// 共有リソースへのアクセスはロックする
		a.mutex.Lock()

		authorStat, exists := authors[authorName]
		if !exists {
			authorStat = models.AuthorStat{
				Name: authorName,
			}
		}
		authorStat.CommitCount++
		authors[authorName] = authorStat

		a.mutex.Unlock()

		// ファイルの変更状況を解析
		// 各コミットで変更されたファイルを追跡
		stats, err := c.Stats()
		if err != nil {
			// エラー処理
			fmt.Printf("コミット %s の解析エラー: %v\n", c.Hash.String(), err)
			continue
		}

		// 各ファイルの統計情報を更新
		a.mutex.Lock()
		for _, stat := range stats {
			filePath := stat.Name
			if !a.isFileMatched(filePath) {
				continue
			}

			fileInfo, exists := a.stats.Files[filePath]
			if !exists {
				fileInfo = models.FileChangeInfo{
					FilePath:       filePath,
					ChangeCount:    0,
					LastModified:   c.Author.When,
					Authors:        make(map[string]int),
					LineChanges:    make(map[int]int),
					CommitMessages: []string{},
					FileType:       filepath.Ext(filePath),
				}
				a.stats.FileCount++
			}

			// 変更回数をカウント
			fileInfo.ChangeCount++
			fileInfo.LastModified = c.Author.When
			fileInfo.Authors[authorName]++

			// コミットメッセージを蓄積（最大5つまで）
			if len(fileInfo.CommitMessages) < 5 {
				if !containsString(fileInfo.CommitMessages, c.Message) {
					fileInfo.CommitMessages = append(fileInfo.CommitMessages, c.Message)
				}
			}

			a.stats.Files[filePath] = fileInfo
		}
		a.mutex.Unlock()
	}
}

// 指定されたファイルパスがパターンに一致するか確認する
func (a *RepositoryAnalyzer) isFileMatched(path string) bool {
	// パターンが指定されていない場合は常にマッチ
	if a.filePattern == "" && a.fileType == "" {
		return true
	}

	// ファイル種別（拡張子）によるフィルタリング
	if a.fileType != "" {
		ext := filepath.Ext(path)
		if ext == "" {
			return false
		}
		// 拡張子の先頭の「.」を取り除いて比較
		if ext[1:] != a.fileType {
			return false
		}
	}

	// ファイルパターンによるフィルタリング
	if a.filePattern != "" {
		matched, err := filepath.Match(a.filePattern, filepath.Base(path))
		if err != nil || !matched {
			return false
		}
	}

	return true
}

// containsString はスライスに特定の文字列が含まれているか確認する
func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// calculateHeatLevels は各ファイルの変更頻度に基づいてヒートレベルを計算する
func (a *RepositoryAnalyzer) calculateHeatLevels() {
	maxChanges := 0

	// 最大変更回数を見つける
	for _, fileInfo := range a.stats.Files {
		if fileInfo.ChangeCount > maxChanges {
			maxChanges = fileInfo.ChangeCount
		}
	}

	// 変更回数を0.0-1.0の範囲に正規化
	if maxChanges > 0 {
		for filePath, fileInfo := range a.stats.Files {
			fileInfo.HeatLevel = float64(fileInfo.ChangeCount) / float64(maxChanges)
			a.stats.Files[filePath] = fileInfo
		}
	}
}
