package models

import "time"

// FileChangeInfo はファイルの変更情報を表す構造体
type FileChangeInfo struct {
	FilePath       string         `json:"file_path"`       // ファイルパス
	ChangeCount    int            `json:"change_count"`    // 変更回数
	LastModified   time.Time      `json:"last_modified"`   // 最終変更日時
	Authors        map[string]int `json:"authors"`         // 著者別の変更回数
	LineChanges    map[int]int    `json:"line_changes"`    // 行番号ごとの変更回数
	LineContents   map[int]string `json:"line_contents"`   // 行番号ごとのテキスト内容
	CommitMessages []string       `json:"commit_messages"` // 関連するコミットメッセージ
	FileType       string         `json:"file_type"`       // ファイルタイプ（拡張子など）
	HeatLevel      float64        `json:"heat_level"`      // 相対的な変更頻度レベル (0.0-1.0)
}

// CommitInfo はコミット情報を表す構造体
type CommitInfo struct {
	Hash      string    `json:"hash"`      // コミットハッシュ
	Author    string    `json:"author"`    // 著者名
	Email     string    `json:"email"`     // 著者のメールアドレス
	Message   string    `json:"message"`   // コミットメッセージ
	Timestamp time.Time `json:"timestamp"` // コミット日時
	Files     []string  `json:"files"`     // 変更されたファイル一覧
}

// RepositoryStats はリポジトリ全体の統計情報を表す構造体
type RepositoryStats struct {
	RepositoryName string                    `json:"repository_name"` // リポジトリ名
	AnalyzedAt     time.Time                 `json:"analyzed_at"`     // 解析実行日時
	CommitCount    int                       `json:"commit_count"`    // 総コミット数
	AuthorCount    int                       `json:"author_count"`    // 著者数
	FileCount      int                       `json:"file_count"`      // ファイル数
	Files          map[string]FileChangeInfo `json:"files"`           // ファイルごとの変更情報
	TopAuthors     []AuthorStat              `json:"top_authors"`     // 上位貢献者
	TopFiles       []string                  `json:"top_files"`       // 最も変更頻度の高いファイル
	FirstCommitAt  time.Time                 `json:"first_commit_at"` // 最初のコミット日時
	LastCommitAt   time.Time                 `json:"last_commit_at"`  // 最後のコミット日時
}

// AuthorStat は著者の貢献統計を表す構造体
type AuthorStat struct {
	Name         string `json:"name"`          // 著者名
	CommitCount  int    `json:"commit_count"`  // コミット数
	FilesChanged int    `json:"files_changed"` // 変更したファイル数
}
