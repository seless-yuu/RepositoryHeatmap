package heatmap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/repositoryheatmap/pkg/models"
)

// Visualizer はヒートマップの可視化を担当する構造体
type Visualizer struct {
	outputDir  string
	outputType string
	stats      *models.RepositoryStats
}

// NewVisualizer は新しいVisualizerインスタンスを作成する
func NewVisualizer(outputDir, outputType string, stats *models.RepositoryStats) *Visualizer {
	return &Visualizer{
		outputDir:  outputDir,
		outputType: strings.ToLower(outputType),
		stats:      stats,
	}
}

// Visualize はヒートマップデータを可視化する
func (v *Visualizer) Visualize() error {
	// 出力タイプのチェック
	if v.outputType != "svg" && v.outputType != "webp" {
		return fmt.Errorf("サポートされていない出力形式です: %s (svg または webp を使用してください)", v.outputType)
	}

	// リポジトリ全体のヒートマップを生成
	if err := v.generateRepositoryHeatmap(); err != nil {
		return fmt.Errorf("リポジトリヒートマップの生成に失敗しました: %w", err)
	}

	// ファイル単位のヒートマップを生成
	if err := v.generateFileHeatmaps(); err != nil {
		return fmt.Errorf("ファイルヒートマップの生成に失敗しました: %w", err)
	}

	return nil
}

// generateRepositoryHeatmap はリポジトリ全体のヒートマップを生成する
func (v *Visualizer) generateRepositoryHeatmap() error {
	// 出力ファイルパスを作成
	outputPath := filepath.Join(v.outputDir, fmt.Sprintf("%s-repository-heatmap.%s", v.stats.RepositoryName, v.outputType))

	// SVG形式の場合
	if v.outputType == "svg" {
		return v.generateSVGRepositoryHeatmap(outputPath)
	}

	// WebP形式の場合
	return v.generateWebPRepositoryHeatmap(outputPath)
}

// generateSVGRepositoryHeatmap はSVG形式のリポジトリヒートマップを生成する
func (v *Visualizer) generateSVGRepositoryHeatmap(outputPath string) error {
	// TODO: 実際のSVG生成処理の実装
	// この実装は簡略化されており、実際のプロジェクトでは
	// SVGライブラリを使用してヒートマップを生成する

	// ダミーのSVGファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("SVGファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// SVGの基本構造を書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="800" height="600" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <text x="50%%" y="50" text-anchor="middle" font-size="24">%s リポジトリヒートマップ</text>
  <!-- ここにヒートマップの詳細なSVG要素が生成されます -->
</svg>`, v.stats.RepositoryName)

	fmt.Printf("リポジトリヒートマップをSVGに保存しました: %s\n", outputPath)
	return nil
}

// generateWebPRepositoryHeatmap はWebP形式のリポジトリヒートマップを生成する
func (v *Visualizer) generateWebPRepositoryHeatmap(outputPath string) error {
	// TODO: 実際のWebP生成処理の実装
	// Go言語でWebP形式を生成するには、別のライブラリが必要になる可能性があります

	// ダミーのファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("WebPファイルの作成に失敗しました: %w", err)
	}
	file.Close()

	fmt.Printf("リポジトリヒートマップをWebPに保存しました: %s\n", outputPath)
	return nil
}

// generateFileHeatmaps は各ファイルのヒートマップを生成する
func (v *Visualizer) generateFileHeatmaps() error {
	// ファイルヒートマップの出力ディレクトリを作成
	fileHeatmapDir := filepath.Join(v.outputDir, "file-heatmaps")
	if err := os.MkdirAll(fileHeatmapDir, 0755); err != nil {
		return fmt.Errorf("ファイルヒートマップディレクトリの作成に失敗しました: %w", err)
	}

	// 各ファイルのヒートマップを生成
	for filePath, fileInfo := range v.stats.Files {
		// ファイル名を安全な形式に変換
		safeFileName := strings.ReplaceAll(filePath, string(os.PathSeparator), "_")
		safeFileName = strings.ReplaceAll(safeFileName, ".", "_")
		outputPath := filepath.Join(fileHeatmapDir, fmt.Sprintf("%s.%s", safeFileName, v.outputType))

		// ファイルの変更頻度に基づく色を取得
		color := GetHeatColor(fileInfo.HeatLevel)

		// SVG形式の場合
		if v.outputType == "svg" {
			if err := v.generateSVGFileHeatmap(outputPath, filePath, fileInfo, color); err != nil {
				return fmt.Errorf("ファイル '%s' のSVGヒートマップ生成に失敗しました: %w", filePath, err)
			}
		} else {
			// WebP形式の場合
			if err := v.generateWebPFileHeatmap(outputPath, filePath, fileInfo, color); err != nil {
				return fmt.Errorf("ファイル '%s' のWebPヒートマップ生成に失敗しました: %w", filePath, err)
			}
		}
	}

	fmt.Printf("ファイルヒートマップを保存しました: %s\n", fileHeatmapDir)
	return nil
}

// generateSVGFileHeatmap はSVG形式のファイルヒートマップを生成する
func (v *Visualizer) generateSVGFileHeatmap(outputPath, filePath string, fileInfo models.FileChangeInfo, color string) error {
	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("SVGファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// SVGの基本構造を書き込む
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="800" height="400" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <text x="10" y="30" font-size="16" font-family="monospace">ファイル: %s</text>
  <text x="10" y="50" font-size="14" font-family="monospace">変更回数: %d</text>
  <rect x="10" y="70" width="780" height="30" fill="%s"/>
  <!-- ここに行ごとのヒートマップが生成されます -->
</svg>`, filePath, fileInfo.ChangeCount, color)

	return nil
}

// generateWebPFileHeatmap はWebP形式のファイルヒートマップを生成する
func (v *Visualizer) generateWebPFileHeatmap(outputPath, filePath string, fileInfo models.FileChangeInfo, color string) error {
	// TODO: 実際のWebP生成処理の実装
	// Go言語でWebP形式を生成するには、別のライブラリが必要になる可能性があります

	// ダミーのファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("WebPファイルの作成に失敗しました: %w", err)
	}
	file.Close()

	return nil
}
