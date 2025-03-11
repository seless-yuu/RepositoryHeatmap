package heatmap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/repositoryheatmap/pkg/models"
)

// Generator is a struct responsible for generating heatmap data
type Generator struct {
	outputDir string
	stats     *models.RepositoryStats
}

// NewGenerator creates a new Generator instance
func NewGenerator(outputDir string, stats *models.RepositoryStats) *Generator {
	return &Generator{
		outputDir: outputDir,
		stats:     stats,
	}
}

// Generate generates heatmap data from repository statistics
func (g *Generator) Generate() error {
	// Generate JSON data
	err := g.generateJSONData()
	if err != nil {
		return fmt.Errorf("failed to generate JSON data: %w", err)
	}

	return nil
}

// generateJSONData generates JSON data from repository statistics
func (g *Generator) generateJSONData() error {
	// Prepare file path
	jsonPath := filepath.Join(g.outputDir, fmt.Sprintf("%s-heatmap.json", g.stats.RepositoryName))

	// Create file
	file, err := os.Create(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}
	defer file.Close()

	// Sort file list
	sortedFiles := getSortedFilesByHeat(g.stats)

	// Set top files
	topN := 10
	if len(sortedFiles) < topN {
		topN = len(sortedFiles)
	}
	for i := 0; i < topN; i++ {
		g.stats.TopFiles = append(g.stats.TopFiles, sortedFiles[i].FilePath)
	}

	// Sort authors
	sort.Slice(g.stats.TopAuthors, func(i, j int) bool {
		return g.stats.TopAuthors[i].CommitCount > g.stats.TopAuthors[j].CommitCount
	})

	// Limit top authors
	if len(g.stats.TopAuthors) > 10 {
		g.stats.TopAuthors = g.stats.TopAuthors[:10]
	}

	// JSON encode
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(g.stats); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	fmt.Printf("Heatmap data saved to JSON: %s\n", jsonPath)
	return nil
}

// getSortedFilesByHeat returns files sorted by heat level
func getSortedFilesByHeat(stats *models.RepositoryStats) []models.FileChangeInfo {
	files := make([]models.FileChangeInfo, 0, len(stats.Files))

	// Convert map to array
	for _, fileInfo := range stats.Files {
		files = append(files, fileInfo)
	}

	// Sort by change count in descending order
	sort.Slice(files, func(i, j int) bool {
		return files[i].ChangeCount > files[j].ChangeCount
	})

	return files
}
