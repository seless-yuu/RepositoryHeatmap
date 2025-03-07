package git

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/repositoryheatmap/pkg/models"
)

// RepositoryAnalyzer はGitリポジトリ解析を担当する構造体
type RepositoryAnalyzer struct {
	repoPath string
	repo     *git.Repository
	stats    *models.RepositoryStats
}

// NewAnalyzer は新しいRepositoryAnalyzerインスタンスを作成する
func NewAnalyzer(repoPath string) (*RepositoryAnalyzer, error) {
	stats := &models.RepositoryStats{
		AnalyzedAt: time.Now(),
		Files:      make(map[string]models.FileChangeInfo),
	}

	return &RepositoryAnalyzer{
		repoPath: repoPath,
		stats:    stats,
	}, nil
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

	authors := make(map[string]models.AuthorStat)
	var firstCommit, lastCommit *object.Commit

	// 各コミットを処理
	err = commits.ForEach(func(c *object.Commit) error {
		// コミット数をカウント
		a.stats.CommitCount++

		// 最初と最後のコミットを記録
		if firstCommit == nil || c.Committer.When.Before(firstCommit.Committer.When) {
			firstCommit = c
		}
		if lastCommit == nil || c.Committer.When.After(lastCommit.Committer.When) {
			lastCommit = c
		}

		// 著者情報を更新
		authorName := c.Author.Name
		authorStat, exists := authors[authorName]
		if !exists {
			authorStat = models.AuthorStat{
				Name: authorName,
			}
		}
		authorStat.CommitCount++
		authors[authorName] = authorStat

		// ファイル変更を処理
		stats, err := c.Stats()
		if err != nil {
			return fmt.Errorf("コミット統計を取得できませんでした: %w", err)
		}

		for _, stat := range stats {
			filePath := stat.Name
			fileInfo, exists := a.stats.Files[filePath]
			if !exists {
				fileInfo = models.FileChangeInfo{
					FilePath:       filePath,
					Authors:        make(map[string]int),
					LineChanges:    make(map[int]int),
					CommitMessages: []string{},
					FileType:       filepath.Ext(filePath),
				}
				a.stats.FileCount++
			}

			// ファイル変更情報を更新
			fileInfo.ChangeCount++
			fileInfo.LastModified = c.Committer.When
			fileInfo.Authors[authorName]++

			// コミットメッセージを保存（重複を避けるためには別の処理が必要）
			fileInfo.CommitMessages = append(fileInfo.CommitMessages, c.Message)

			a.stats.Files[filePath] = fileInfo
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("コミット解析中にエラーが発生しました: %w", err)
	}

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
