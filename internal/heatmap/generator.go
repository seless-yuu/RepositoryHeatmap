package heatmap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/repositoryheatmap/pkg/models"
)

// Generator はヒートマップデータ生成を担当する構造体
type Generator struct {
	outputDir string
	stats     *models.RepositoryStats
}

// NewGenerator は新しいGeneratorインスタンスを作成する
func NewGenerator(outputDir string, stats *models.RepositoryStats) *Generator {
	return &Generator{
		outputDir: outputDir,
		stats:     stats,
	}
}

// Generate はリポジトリ統計からヒートマップデータを生成する
func (g *Generator) Generate() error {
	// JSONデータを生成
	err := g.generateJSONData()
	if err != nil {
		return fmt.Errorf("JSONデータの生成に失敗しました: %w", err)
	}

	return nil
}

// generateJSONData はリポジトリ統計からJSONデータを生成する
func (g *Generator) generateJSONData() error {
	// ファイルパスを準備
	jsonPath := filepath.Join(g.outputDir, fmt.Sprintf("%s-heatmap.json", g.stats.RepositoryName))

	// ファイルの作成
	file, err := os.Create(jsonPath)
	if err != nil {
		return fmt.Errorf("JSONファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// ファイルリストをソート
	sortedFiles := getSortedFilesByHeat(g.stats)

	// トップファイルを設定
	topN := 10
	if len(sortedFiles) < topN {
		topN = len(sortedFiles)
	}
	for i := 0; i < topN; i++ {
		g.stats.TopFiles = append(g.stats.TopFiles, sortedFiles[i].FilePath)
	}

	// 著者をソート
	sort.Slice(g.stats.TopAuthors, func(i, j int) bool {
		return g.stats.TopAuthors[i].CommitCount > g.stats.TopAuthors[j].CommitCount
	})

	// トップ著者を制限
	if len(g.stats.TopAuthors) > 10 {
		g.stats.TopAuthors = g.stats.TopAuthors[:10]
	}

	// JSONエンコード
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(g.stats); err != nil {
		return fmt.Errorf("JSONエンコードに失敗しました: %w", err)
	}

	fmt.Printf("ヒートマップデータをJSONに保存しました: %s\n", jsonPath)
	return nil
}

// getSortedFilesByHeat は変更頻度でソートされたファイルリストを返す
func getSortedFilesByHeat(stats *models.RepositoryStats) []models.FileChangeInfo {
	files := make([]models.FileChangeInfo, 0, len(stats.Files))

	// マップから配列に変換
	for _, fileInfo := range stats.Files {
		files = append(files, fileInfo)
	}

	// 変更回数の多い順にソート
	sort.Slice(files, func(i, j int) bool {
		return files[i].ChangeCount > files[j].ChangeCount
	})

	return files
}

// GetHeatColor は変更頻度に基づいて色を返す (0.0-1.0)
func GetHeatColor(heatLevel float64) string {
	// ヒートレベルに応じた色を返す
	// 0.0 (低) = 青系, 0.5 = 緑系, 1.0 (高) = 赤系
	if heatLevel < 0.2 {
		return "#0000FF" // 青
	} else if heatLevel < 0.4 {
		return "#0066FF" // 青緑
	} else if heatLevel < 0.6 {
		return "#00FF00" // 緑
	} else if heatLevel < 0.8 {
		return "#FFCC00" // 黄色
	} else {
		return "#FF0000" // 赤
	}
}
