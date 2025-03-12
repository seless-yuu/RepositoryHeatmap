package git

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/repositoryheatmap/pkg/models"
)

// LineTracker is responsible for tracking line-level changes
type LineTracker struct {
	repo        *RepositoryAnalyzer
	workerCount int
	mutex       sync.Mutex
}

// NewLineTracker creates a new LineTracker instance
func NewLineTracker(repo *RepositoryAnalyzer) *LineTracker {
	return &LineTracker{
		repo:        repo,
		workerCount: 1, // Default is single-threaded
	}
}

// SetWorkerCount sets the number of workers for parallel processing
func (lt *LineTracker) SetWorkerCount(count int) {
	if count < 1 {
		count = 1
	}
	lt.workerCount = count
}

// TrackLineChanges tracks line-level changes in repository files
func (lt *LineTracker) TrackLineChanges(stats *models.RepositoryStats) error {
	if lt.repo == nil || lt.repo.repo == nil {
		return fmt.Errorf("repository is not opened")
	}

	// Get HEAD reference
	ref, err := lt.repo.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	// Get commit history
	commits, err := lt.repo.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}

	// 1. Collect all commits
	var commitList []*object.Commit
	err = commits.ForEach(func(c *object.Commit) error {
		// Check date filter if set
		if lt.repo.sinceDate != nil && c.Committer.When.Before(*lt.repo.sinceDate) {
			return nil // Skip commits before the specified date
		}
		commitList = append(commitList, c)
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to collect commit history: %w", err)
	}

	// Skip if less than 2 commits (can't calculate diff)
	if len(commitList) < 2 {
		return nil
	}

	// 2. Process commit pairs with multi-threading
	if lt.workerCount > 1 && len(commitList) > 1 {
		// Adjust worker count (don't use more workers than commits)
		workerCount := lt.workerCount
		if workerCount > len(commitList)-1 {
			workerCount = len(commitList) - 1
		}

		// Split commit pairs among workers
		pairCount := len(commitList) - 1
		pairsPerWorker := pairCount / workerCount
		if pairsPerWorker < 1 {
			pairsPerWorker = 1
		}

		// Launch workers
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

		// Wait for all workers to complete
		wg.Wait()
	} else {
		// Single-threaded processing
		for i := 0; i < len(commitList)-1; i++ {
			lt.processCommitPair(commitList[i], commitList[i+1], stats)
		}
	}

	return nil
}

// processCommitPair processes the diff between two commits
func (lt *LineTracker) processCommitPair(current, next *object.Commit, stats *models.RepositoryStats) {
	// Get diff between two trees
	patch, err := current.Patch(next)
	if err != nil {
		// Continue even if there's an error
		return
	}

	// Process each file patch
	for _, filePatch := range patch.FilePatches() {
		from, to := filePatch.Files()
		var filePath string

		// Get file path (from 'from' for deleted files, from 'to' for added/modified files)
		if from == nil {
			if to == nil {
				continue // Skip if file doesn't exist
			}
			filePath = to.Path()
		} else {
			filePath = from.Path()
		}

		// Lock access to shared resources
		lt.mutex.Lock()

		// Get file info
		fileInfo, exists := stats.Files[filePath]
		if !exists {
			// Skip files not found during analysis
			lt.mutex.Unlock()
			continue
		}

		// Process each chunk's changes
		currentLine := 1
		for _, chunk := range filePatch.Chunks() {
			// Split chunk content into lines
			lines := strings.Split(chunk.Content(), "\n")
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1] // Remove trailing empty line
			}

			// Process based on chunk type
			switch chunk.Type() {
			case diff.Add:
				// Process added lines
				for _, line := range lines {
					if line == "" {
						continue
					}
					fileInfo.LineChanges[currentLine]++
					currentLine++
				}
			case diff.Delete:
				// Process deleted lines
				for _, line := range lines {
					if line == "" {
						continue
					}
					fileInfo.LineChanges[currentLine]++
					currentLine++
				}
			default:
				// Process unchanged lines
				for range lines {
					currentLine++
				}
			}
		}

		// Update map with modified info
		stats.Files[filePath] = fileInfo

		// Release lock
		lt.mutex.Unlock()
	}
}

// CalculateLineHeatLevels calculates heat levels for each line in files
func (lt *LineTracker) CalculateLineHeatLevels(stats *models.RepositoryStats) {
	// Channel for parallel processing
	type fileJob struct {
		path string
		info models.FileChangeInfo
	}

	// 1. Use multi-threaded processing
	if lt.workerCount > 1 && len(stats.Files) > 1 {
		// Adjust worker count (don't use more workers than files)
		workerCount := lt.workerCount
		if workerCount > len(stats.Files) {
			workerCount = len(stats.Files)
		}

		// Create job channels
		jobs := make(chan fileJob, len(stats.Files))
		results := make(chan fileJob, len(stats.Files))

		// Launch workers
		var wg sync.WaitGroup
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					// Find maximum change count in file
					maxLineChanges := 0
					for _, changes := range job.info.LineChanges {
						if changes > maxLineChanges {
							maxLineChanges = changes
						}
					}

					// Calculate heat level for each line based on change count
					if maxLineChanges > 0 {
						for lineNum, changes := range job.info.LineChanges {
							// Normalize heat level to 0.0-1.0 range
							job.info.LineChanges[lineNum] = int(float64(changes) / float64(maxLineChanges) * 10)
						}
					}

					// Send result
					results <- job
				}
			}()
		}

		// Send jobs
		for path, info := range stats.Files {
			jobs <- fileJob{path: path, info: info}
		}
		close(jobs)

		// Wait for all workers to complete
		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect results
		for job := range results {
			stats.Files[job.path] = job.info
		}
	} else {
		// 2. Single-threaded processing
		// Process each file
		for filePath, fileInfo := range stats.Files {
			// Find maximum change count in file
			maxLineChanges := 0
			for _, changes := range fileInfo.LineChanges {
				if changes > maxLineChanges {
					maxLineChanges = changes
				}
			}

			// Calculate heat level for each line based on change count
			if maxLineChanges > 0 {
				for lineNum, changes := range fileInfo.LineChanges {
					// Normalize heat level to 0.0-1.0 range
					fileInfo.LineChanges[lineNum] = int(float64(changes) / float64(maxLineChanges) * 10)
				}
			}

			// Update map with modified info
			stats.Files[filePath] = fileInfo
		}
	}
}
