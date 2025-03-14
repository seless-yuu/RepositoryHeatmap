package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	// 指定されたパスが存在することを確認
	absPath, err := filepath.Abs(a.repoPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// パスが存在し、ディレクトリであることを確認
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("specified path is not a directory: %s", absPath)
	}

	// .gitディレクトリの存在を確認
	gitDirPath := filepath.Join(absPath, ".git")
	if _, err := os.Stat(gitDirPath); err != nil {
		return fmt.Errorf("specified directory is not a Git repository (no .git directory found): %s", absPath)
	}

	// リポジトリを開く
	a.repoPath = absPath
	repo, err := git.PlainOpen(a.repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	a.repo = repo

	// 新しいRepositoryStatsオブジェクトを作成
	a.stats = &models.RepositoryStats{
		Files:          make(map[string]models.FileChangeInfo),
		RepositoryPath: a.repoPath,
		RepositoryName: filepath.Base(a.repoPath),
	}

	return nil
}

// Clone はリポジトリをクローンする
func (a *RepositoryAnalyzer) Clone(url string) error {
	repo, err := git.PlainClone(a.repoPath, false, &git.CloneOptions{
		URL:      url,
		Progress: nil,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	a.repo = repo

	// リポジトリ名を設定
	a.stats.RepositoryName = filepath.Base(a.repoPath)
	return nil
}

// Analyze はGitリポジトリを解析してヒートマップ情報を生成する
func (a *RepositoryAnalyzer) Analyze() (*models.RepositoryStats, error) {
	// リポジトリが開かれていることを確認
	if a.repo == nil {
		return nil, fmt.Errorf("repository is not opened")
	}

	// 統計情報を初期化
	if a.stats == nil {
		a.stats = &models.RepositoryStats{
			Files:          make(map[string]models.FileChangeInfo),
			RepositoryPath: a.repoPath, // Set repository path
			RepositoryName: filepath.Base(a.repoPath),
		}
	} else {
		// 既に初期化されている場合も、リポジトリパスを確実に設定
		a.stats.RepositoryPath = a.repoPath
	}

	// リポジトリ名を設定
	repoName := filepath.Base(a.repoPath)
	a.stats.RepositoryName = repoName

	// コミット履歴を取得
	commitIter, err := a.repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit history: %w", err)
	}

	// 全コミットを配列に格納
	var commitList []*object.Commit
	var firstCommit, lastCommit *object.Commit
	lastCommitIndex := 0

	err = commitIter.ForEach(func(c *object.Commit) error {
		// 指定日以降のコミットのみを対象にする
		if a.sinceDate != nil && c.Committer.When.Before(*a.sinceDate) {
			return nil
		}

		commitList = append(commitList, c)
		// 最初のコミット（最新）と最後のコミット（最古）を記録
		if lastCommitIndex == 0 {
			lastCommit = c
		}
		firstCommit = c
		lastCommitIndex++

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse commit history: %w", err)
	}

	// 著者の統計情報を保持するマップ
	authors := make(map[string]models.AuthorStat)

	// 並列処理を行う場合（デフォルトはシングルスレッド）
	if a.workerCount > 1 {
		fmt.Printf("Starting parallel commit processing (workers: %d)\n", a.workerCount)

		// WaitGroupを使ってゴルーチンの終了を待機
		var wg sync.WaitGroup

		// 各ワーカーのチャンクサイズを計算
		chunkSize := (len(commitList) + a.workerCount - 1) / a.workerCount

		// ワーカーを起動
		for i := 0; i < a.workerCount; i++ {
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

				// 各ワーカー用の著者情報を一時的に保持
				localAuthors := make(map[string]models.AuthorStat)
				localFiles := make(map[string]models.FileChangeInfo)

				// ローカル処理 - 著者情報を更新
				for _, c := range commits {
					authorName := c.Author.Name
					authorStat, exists := localAuthors[authorName]
					if !exists {
						authorStat = models.AuthorStat{
							Name: authorName,
						}
					}
					authorStat.CommitCount++
					localAuthors[authorName] = authorStat

					// ファイルの変更状況を解析
					// go-gitの操作を同期化
					a.mutex.Lock()
					stats, err := c.Stats()
					a.mutex.Unlock()
					if err != nil {
						fmt.Printf("Error analyzing commit %s: %v\n", c.Hash.String(), err)
						continue
					}

					// 各ファイルの統計情報を更新（ローカル）
					for _, stat := range stats {
						filePath := stat.Name
						if !a.isFileMatched(filePath) {
							continue
						}

						fileInfo, exists := localFiles[filePath]
						if !exists {
							fileInfo = models.FileChangeInfo{
								FilePath:       filePath,
								ChangeCount:    0,
								LastModified:   c.Author.When,
								Authors:        make(map[string]int),
								LineChanges:    make(map[int]int),
								LineContents:   make(map[int]string),
								CommitMessages: []string{},
								FileType:       filepath.Ext(filePath),
							}
						}

						// ファイル情報を更新
						fileInfo.ChangeCount++
						fileInfo.Authors[authorName]++
						if c.Author.When.After(fileInfo.LastModified) {
							fileInfo.LastModified = c.Author.When
						}

						// コミットメッセージを追加（最大5件まで）
						msg := c.Message
						if len(fileInfo.CommitMessages) < 5 && !containsString(fileInfo.CommitMessages, msg) {
							fileInfo.CommitMessages = append(fileInfo.CommitMessages, msg)
						}

						// ファイル内容読み込み処理を追加
						a.mutex.Lock() // ファイル読み込みもロック
						fileObj, err := c.File(filePath)
						if err == nil {
							// バイナリファイルは処理しない
							isBinary, err := fileObj.IsBinary()
							if err == nil && !isBinary {
								// ファイル内容を読み込む
								content, err := fileObj.Contents()
								if err == nil {
									lines := strings.Split(content, "\n")
									for i, line := range lines {
										lineNumber := i + 1
										fileInfo.LineContents[lineNumber] = line
									}
								}
							} else if isBinary {
								// バイナリファイルはスキップ
								fmt.Printf("Skipping binary file '%s'\n", filePath)
							}
						}
						a.mutex.Unlock()

						localFiles[filePath] = fileInfo
					}
				}

				// グローバルマップへの更新をミューテックスで保護
				a.mutex.Lock()

				// 著者情報をマージ
				for name, stat := range localAuthors {
					globalStat, exists := authors[name]
					if exists {
						globalStat.CommitCount += stat.CommitCount
					} else {
						globalStat = stat
					}
					authors[name] = globalStat
				}

				// ファイル情報をマージ
				for path, fileInfo := range localFiles {
					globalFileInfo, exists := a.stats.Files[path]
					if exists {
						// 既存のファイル情報を更新
						globalFileInfo.ChangeCount += fileInfo.ChangeCount
						if fileInfo.LastModified.After(globalFileInfo.LastModified) {
							globalFileInfo.LastModified = fileInfo.LastModified
						}
						// 著者情報をマージ
						for author, count := range fileInfo.Authors {
							globalFileInfo.Authors[author] += count
						}
						// コミットメッセージをマージ（最大5件まで）
						for _, msg := range fileInfo.CommitMessages {
							if len(globalFileInfo.CommitMessages) < 5 && !containsString(globalFileInfo.CommitMessages, msg) {
								globalFileInfo.CommitMessages = append(globalFileInfo.CommitMessages, msg)
							}
						}
						a.stats.Files[path] = globalFileInfo
					} else {
						// 新しいファイル
						a.stats.Files[path] = fileInfo
						a.stats.FileCount++
					}
				}

				a.mutex.Unlock()
			}(commitList[startIdx:endIdx])
		}

		// すべてのワーカーの完了を待機
		wg.Wait()
	} else {
		// シングルスレッドで処理
		for _, c := range commitList {
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

			// ファイルの変更状況を解析
			stats, err := c.Stats()
			if err != nil {
				// エラー処理
				fmt.Printf("Error analyzing commit %s: %v\n", c.Hash.String(), err)
				continue
			}

			// 各ファイルの統計情報を更新
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
						LineContents:   make(map[int]string),
						CommitMessages: []string{},
						FileType:       filepath.Ext(filePath),
					}
					a.stats.FileCount++

					// ファイル内容読み込み - 最新コミットの条件を削除
					// ファイルオブジェクトを取得
					fileObj, err := c.File(filePath)
					if err == nil {
						// バイナリファイルは処理しない
						isBinary, err := fileObj.IsBinary()
						if err == nil && !isBinary {
							// ファイル内容を読み込む
							content, err := fileObj.Contents()
							if err == nil {
								lines := strings.Split(content, "\n")
								for i, line := range lines {
									lineNumber := i + 1
									fileInfo.LineContents[lineNumber] = line
								}
							}
						} else if isBinary {
							// バイナリファイルはスキップ
							fmt.Printf("Skipping binary file '%s'\n", filePath)
						}
					}
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
		}
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
