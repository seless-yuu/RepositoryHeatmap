package models

import "time"

// FileChangeInfo is a struct representing file change information
type FileChangeInfo struct {
	FilePath       string         `json:"file_path"`       // File path
	ChangeCount    int            `json:"change_count"`    // Change count
	LastModified   time.Time      `json:"last_modified"`   // Last modified date
	Authors        map[string]int `json:"authors"`         // Change count by author
	LineChanges    map[int]int    `json:"line_changes"`    // Change count by line number
	LineContents   map[int]string `json:"line_contents"`   // Text content by line number
	CommitMessages []string       `json:"commit_messages"` // Related commit messages
	FileType       string         `json:"file_type"`       // File type (extension, etc.)
	HeatLevel      float64        `json:"heat_level"`      // Relative change frequency level (0.0-1.0)
}

// CommitInfo is a struct representing commit information
type CommitInfo struct {
	Hash      string    `json:"hash"`      // Commit hash
	Author    string    `json:"author"`    // Author name
	Email     string    `json:"email"`     // Author's email address
	Message   string    `json:"message"`   // Commit message
	Timestamp time.Time `json:"timestamp"` // Commit datetime
	Files     []string  `json:"files"`     // List of changed files
}

// RepositoryStats is a struct representing repository-wide statistics
type RepositoryStats struct {
	RepositoryName string                    `json:"repository_name"` // Repository name
	RepositoryPath string                    `json:"repository_path"` // Repository absolute path
	AnalyzedAt     time.Time                 `json:"analyzed_at"`     // Analysis execution datetime
	CommitCount    int                       `json:"commit_count"`    // Total commit count
	AuthorCount    int                       `json:"author_count"`    // Author count
	FileCount      int                       `json:"file_count"`      // File count
	Files          map[string]FileChangeInfo `json:"files"`           // Change information by file
	TopAuthors     []AuthorStat              `json:"top_authors"`     // Top contributors
	TopFiles       []string                  `json:"top_files"`       // Most frequently changed files
	FirstCommitAt  time.Time                 `json:"first_commit_at"` // First commit datetime
	LastCommitAt   time.Time                 `json:"last_commit_at"`  // Last commit datetime
}

// AuthorStat is a struct representing author contribution statistics
type AuthorStat struct {
	Name         string `json:"name"`          // Author name
	CommitCount  int    `json:"commit_count"`  // Commit count
	FilesChanged int    `json:"files_changed"` // Number of files changed
}
