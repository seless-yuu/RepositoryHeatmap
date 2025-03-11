package heatmap

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"github.com/repositoryheatmap/internal/utils"
	"github.com/repositoryheatmap/pkg/models"
)

// Visualizer is a struct responsible for visualizing heatmaps
type Visualizer struct {
	outputDir      string
	outputType     string
	stats          *models.RepositoryStats
	maxFilesToShow int
	inputJSONPath  string
}

// TreeNode represents a node in the treemap
type TreeNode struct {
	Name      string
	Path      string
	Size      float64 // File change frequency or heat level
	HeatLevel float64 // Heat level between 0.0-1.0
	Children  []*TreeNode
	IsDir     bool
	Rect      Rectangle // Rectangle area for drawing
}

// Rectangle represents a rectangular area in the treemap
type Rectangle struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// NewVisualizer creates a new Visualizer instance
func NewVisualizer(outputDir, outputType string, stats *models.RepositoryStats, maxFilesToShow int, inputJSONPath string) *Visualizer {
	return &Visualizer{
		outputDir:      outputDir,
		outputType:     outputType,
		stats:          stats,
		maxFilesToShow: maxFilesToShow,
		inputJSONPath:  inputJSONPath,
	}
}

// SetRepoPath sets the repository path
func (v *Visualizer) SetRepoPath(repoPath string) {
	// This function is kept for backward compatibility, but does nothing
}

// SetMaxFilesToShow sets the maximum number of files to display
func (v *Visualizer) SetMaxFilesToShow(maxFiles int) {
	if maxFiles < 1 {
		maxFiles = 1 // At least one file should be displayed
	}
	v.maxFilesToShow = maxFiles
}

// Visualize generates the heatmap
func (v *Visualizer) Visualize() error {
	// Check if the output directory exists
	if err := utils.EnsureDirectoryExists(v.outputDir); err != nil {
		return fmt.Errorf("Failed to create output directory: %w", err)
	}

	// Generate heatmap based on the output type
	if v.outputType == "svg" || v.outputType == "webp" {
		// Generate repository-wide heatmap
		if err := v.generateRepositoryHeatmap(); err != nil {
			return fmt.Errorf("Failed to generate repository heatmap: %w", err)
		}

		// Generate file-wide heatmaps
		if err := v.generateFileHeatmaps(); err != nil {
			return fmt.Errorf("Failed to generate file heatmaps: %w", err)
		}
	} else {
		return fmt.Errorf("Unsupported output format: %s", v.outputType)
	}

	fmt.Printf("Visualization completed\n")
	fmt.Printf("Results saved in %s heatmap directory\n", v.outputDir)

	return nil
}

// GetHeatColor returns the color based on the heat level
func GetHeatColor(heatLevel float64) string {
	// Gentle warm color scheme from low to high temperature
	if heatLevel < 0.2 {
		return "#FFF5E6" // Very low!
	} else if heatLevel < 0.4 {
		return "#FFE0B2" // Low!
	} else if heatLevel < 0.6 {
		return "#FFAB66" // Medium orange
	} else if heatLevel < 0.8 {
		return "#FF8533" // Medium orange
	} else {
		return "#E65C00" // High!
	}
}

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// generateRepositoryHeatmap generates the repository-wide heatmap
func (v *Visualizer) generateRepositoryHeatmap() error {
	// Verify repository name
	if v.stats.RepositoryName == "" {
		return fmt.Errorf("Repository name not set")
	}

	// Check and create output directory
	if err := utils.EnsureDirectoryExists(v.outputDir); err != nil {
		return fmt.Errorf("Failed to create output directory: %w", err)
	}

	// Determine output file path
	outputFileName := fmt.Sprintf("%s-repository-heatmap.%s", v.stats.RepositoryName, v.outputType)
	outputPath := filepath.Join(v.outputDir, outputFileName)

	// Create directory for file heatmaps
	fileHeatmapDir := filepath.Join(v.outputDir, "file-heatmaps")
	if err := utils.EnsureDirectoryExists(fileHeatmapDir); err != nil {
		return fmt.Errorf("Failed to create file heatmap directory: %w", err)
	}

	// Generate heatmap based on output type
	var err error
	if v.outputType == "svg" {
		err = v.generateSVGRepositoryHeatmap(outputPath)
	} else {
		err = v.generateWebPRepositoryHeatmap(outputPath)
	}

	if err != nil {
		return err
	}

	// Generate file-wide heatmaps
	if err := v.generateFileHeatmaps(); err != nil {
		return err
	}

	return nil
}

// generateSVGRepositoryHeatmap generates SVG repository heatmap
func (v *Visualizer) generateSVGRepositoryHeatmap(outputPath string) error {
	// Create file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Failed to create SVG file: %w", err)
	}
	defer file.Close()

	// Drawing parameters
	canvasWidth := 1200
	canvasHeight := 900
	if len(v.stats.Files) > 20 {
		// Adjust height if there are many files
		canvasHeight = 900 + (len(v.stats.Files)/20)*100
	}
	marginTop := 100
	marginRight := 50
	marginBottom := 50
	marginLeft := 50
	treeMapWidth := canvasWidth - marginLeft - marginRight
	treeMapHeight := canvasHeight - marginTop - marginBottom

	// Sort files by change frequency and limit display count
	files := getSortedFilesByHeat(v.stats)
	if v.maxFilesToShow > 0 {
		if v.maxFilesToShow < len(files) {
			files = files[:v.maxFilesToShow]
		}
	} else {
		v.maxFilesToShow = len(files) // If maxFilesToShow is 0, display all files
	}

	// Build tree structure
	root := v.buildFileTreeFromFiles(files)

	// Calculate treemap layout
	rootRect := Rectangle{
		X:      float64(marginLeft),
		Y:      float64(marginTop),
		Width:  float64(treeMapWidth),
		Height: float64(treeMapHeight),
	}
	v.layoutTreeMap(root, rootRect)

	// Write SVG header
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <title>%s Repository Heatmap (%s)</title>
  <text x="%d" y="30" text-anchor="middle" font-size="20" font-family="Segoe UI, Helvetica, Arial" fill="#333333">%s</text>
  <text x="%d" y="50" text-anchor="middle" font-size="12" font-family="Segoe UI, Helvetica, Arial" fill="#666666">Commit count: %d, File count: %d</text>
  <text x="%d" y="65" text-anchor="middle" font-size="12" font-family="Segoe UI, Helvetica, Arial" fill="#666666">Period: %s ~ %s</text>
  <text x="%d" y="80" text-anchor="middle" font-size="12" font-family="Segoe UI, Helvetica, Arial" fill="#666666">Display: %d / %d files</text>
  <g transform="translate(10, 100)">
    <text x="10" y="0" font-size="10" font-family="Segoe UI, Helvetica, Arial" fill="#333333">Low</text>
    <rect x="30" y="-10" width="20" height="12" fill="#FFFFFF" stroke="#CCCCCC" />
    <rect x="80" y="-10" width="20" height="12" fill="#FFE5CC" />
    <rect x="130" y="-10" width="20" height="12" fill="#FFD199" />
    <rect x="180" y="-10" width="20" height="12" fill="#FFAB66" />
    <rect x="230" y="-10" width="20" height="12" fill="#FF8533" />
    <rect x="280" y="-10" width="20" height="12" fill="#E65C00" />
    <text x="305" y="0" font-size="10" font-family="Segoe UI, Helvetica, Arial" fill="#333333">High</text>
  </g>
`,
		canvasWidth, canvasHeight,
		v.stats.RepositoryName, v.outputType,
		canvasWidth/2, v.stats.RepositoryName,
		canvasWidth/2, v.stats.CommitCount, len(v.stats.Files),
		canvasWidth/2, v.stats.FirstCommitAt.Format("2006/01/02"), v.stats.LastCommitAt.Format("2006/01/02"),
		canvasWidth/2, len(files), len(v.stats.Files))

	// Draw treemap
	v.renderTreeMap(file, root)

	// Write SVG footer
	fmt.Fprintf(file, `</svg>`)

	fmt.Printf("Repository heatmap saved to SVG: %s\n", outputPath)
	return nil
}

// buildFileTreeFromFiles builds tree structure from file information
func (v *Visualizer) buildFileTreeFromFiles(files []models.FileChangeInfo) *TreeNode {
	root := &TreeNode{
		Name:     "root",
		Path:     "",
		IsDir:    true,
		Children: []*TreeNode{},
	}

	// Add files to tree
	for _, fileInfo := range files {
		path := fileInfo.FilePath
		pathParts := strings.Split(path, "/")

		// If path is not separated by '/' or '\', it's Windows
		if len(pathParts) == 1 {
			pathParts = strings.Split(path, "\\")
		}

		currentNode := root

		// Build treemap hierarchy
		for i := 0; i < len(pathParts)-1; i++ {
			dirName := pathParts[i]
			if dirName == "" {
				continue
			}

			// Search for existing treemap node
			found := false
			for _, child := range currentNode.Children {
				if child.Name == dirName && child.IsDir {
					currentNode = child
					found = true
					break
				}
			}

			// If not found, create new treemap node
			if !found {
				dirNode := &TreeNode{
					Name:     dirName,
					Path:     strings.Join(pathParts[:i+1], "/"),
					IsDir:    true,
					Children: []*TreeNode{},
				}
				currentNode.Children = append(currentNode.Children, dirNode)
				currentNode = dirNode
			}
		}

		// Add file node
		fileName := pathParts[len(pathParts)-1]
		fileNode := &TreeNode{
			Name:      fileName,
			Path:      fileInfo.FilePath,
			Size:      float64(fileInfo.ChangeCount),
			HeatLevel: fileInfo.HeatLevel,
			IsDir:     false,
			Children:  []*TreeNode{},
		}
		currentNode.Children = append(currentNode.Children, fileNode)
	}

	// Calculate treemap size
	calculateDirSize(root)

	return root
}

// calculateDirSize calculates the size of treemap node (recursive)
func calculateDirSize(node *TreeNode) float64 {
	if !node.IsDir {
		return node.Size
	}

	var total float64
	for _, child := range node.Children {
		total += calculateDirSize(child)
	}

	node.Size = total
	node.HeatLevel = 0 // Treemap heat level can also be calculated using average of child nodes
	return total
}

// layoutTreeMap calculates treemap layout
func (v *Visualizer) layoutTreeMap(node *TreeNode, rect Rectangle) {
	node.Rect = rect

	if !node.IsDir || len(node.Children) == 0 {
		return
	}

	// Calculate total size of child nodes
	var totalSize float64
	for _, child := range node.Children {
		totalSize += child.Size
	}

	// If total size is 0, do nothing
	if totalSize <= 0 {
		return
	}

	// Square-based layout algorithm
	// If rectangle is horizontal, split vertically; if vertical, split horizontally
	isHorizontal := rect.Width > rect.Height

	// Current position
	x := rect.X
	y := rect.Y

	// Remaining height
	remainingWidth := rect.Width
	remainingHeight := rect.Height

	for _, child := range node.Children {
		// Ratio of child node
		ratio := child.Size / totalSize

		// Calculate child node rectangle
		var childRect Rectangle
		if isHorizontal {
			// Split horizontally
			childRect = Rectangle{
				X:      x,
				Y:      y,
				Width:  remainingWidth * ratio,
				Height: rect.Height,
			}
			x += childRect.Width
		} else {
			// Split vertically
			childRect = Rectangle{
				X:      x,
				Y:      y,
				Width:  rect.Width,
				Height: remainingHeight * ratio,
			}
			y += childRect.Height
		}

		// Recursive layout
		v.layoutTreeMap(child, childRect)
	}
}

// renderTreeMap draws treemap
func (v *Visualizer) renderTreeMap(file *os.File, node *TreeNode) {
	// Set border and background color
	border := "#CCCCCC"
	bgColor := "#FFFFFF"

	// Assign color based on heat level to treemap nodes
	if node.IsDir {
		bgColor = "#F8F8F8" // Treemap directory is slightly gray
	} else {
		// Get color based on heat level
		bgColor = GetHeatColor(node.HeatLevel)
	}

	// Draw treemap node rectangle
	fmt.Fprintf(file, `  <rect x="%f" y="%f" width="%f" height="%f" fill="%s" stroke="%s">
    <title>%s (Changed: %d)</title>
  </rect>`,
		node.Rect.X, node.Rect.Y, node.Rect.Width, node.Rect.Height,
		bgColor, border, node.Path, int(node.Size))

	// If it's a file node, add link
	if !node.IsDir {
		// Sanitize file name
		safeFileName := utils.SanitizePath(node.Path)

		// Debug information
		fmt.Printf("Link generation: %s -> %s\n", node.Path, safeFileName)

		// Link target is file in file-heatmaps directory
		// Always specify relative path (relative path from repository heatmap SVG)
		linkPath := fmt.Sprintf("./file-heatmaps/%s.%s", safeFileName, v.outputType)

		// Start link tag (add escape processing)
		fmt.Fprintf(file, `  <a href="%s" target="_blank">`, html.EscapeString(linkPath))
	}

	// Display node name if there's enough space for it (width > 40px and height > 14px)
	if node.Rect.Width > 40 && node.Rect.Height > 14 {
		// Display file/directory name
		name := node.Name
		// If name is long, it's omitted (fit to display width)
		if float64(len(name)*7) > node.Rect.Width {
			maxChars := int(node.Rect.Width / 7)
			if maxChars > 3 {
				name = name[:maxChars-3] + "..."
			} else if maxChars > 0 {
				name = name[:maxChars]
			} else {
				name = ""
			}
		}

		if name != "" {
			// Text color is white or black (based on background color)
			textColor := "#000000" // Default is black
			if node.HeatLevel > 0.6 {
				textColor = "#FFFFFF" // White text for dark background color
			}

			fmt.Fprintf(file, `  <text x="%f" y="%f" font-size="12" font-family="Segoe UI, Helvetica, Arial" fill="%s">%s</text>`,
				node.Rect.X+5, node.Rect.Y+15, textColor, name)
		}
	}

	// If it's a file node, close link tag
	if !node.IsDir {
		fmt.Fprintf(file, `  </a>`)
	}

	// Recursively draw child nodes
	for _, child := range node.Children {
		v.renderTreeMap(file, child)
	}
}

// generateWebPRepositoryHeatmap generates WebP repository heatmap
func (v *Visualizer) generateWebPRepositoryHeatmap(outputPath string) error {
	// TODO: Actual WebP generation processing
	// Go language may require a different library to generate WebP format

	// Create dummy file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Failed to create WebP file: %w", err)
	}
	file.Close()

	fmt.Printf("Repository heatmap saved to WebP: %s\n", outputPath)
	return nil
}

// generateFileHeatmaps generates file-wide heatmaps
func (v *Visualizer) generateFileHeatmaps() error {
	// Check output directory
	fileHeatmapDir := filepath.Join(v.outputDir, "file-heatmaps")
	if err := utils.EnsureDirectoryExists(fileHeatmapDir); err != nil {
		return fmt.Errorf("Failed to create file heatmap directory: %w", err)
	}

	// Sort files by change frequency
	files := getSortedFilesByHeat(v.stats)

	// Check maxFilesToShow value
	maxFiles := v.maxFilesToShow
	if maxFiles <= 0 || maxFiles > len(files) {
		maxFiles = len(files)
	}

	// Generate heatmaps only for selected files
	selectedFiles := files[:maxFiles]

	// Generate heatmaps for selected files
	for _, fileInfo := range selectedFiles {
		filePath := fileInfo.FilePath
		// Convert file path to safe format
		safeFileName := utils.SanitizePath(filePath)

		// Create output file path
		outputPath := filepath.Join(fileHeatmapDir, safeFileName+"."+v.outputType)

		// Check if parent directory exists
		parentDir := filepath.Dir(outputPath)
		if err := utils.EnsureDirectoryExists(parentDir); err != nil {
			return fmt.Errorf("Failed to create directory: %s: %w", parentDir, err)
		}

		// Get color based on file change frequency
		color := GetHeatColor(fileInfo.HeatLevel)

		// SVG format
		var err error
		if v.outputType == "svg" {
			err = v.generateSVGFileHeatmap(outputPath, filePath, fileInfo, color)
		} else {
			// WebP format
			err = v.generateWebPFileHeatmap(outputPath, filePath, fileInfo, color)
		}

		// If error occurs, abort processing
		if err != nil {
			return fmt.Errorf("Failed to generate heatmap for file '%s': %w", filePath, err)
		}
	}

	return nil
}

// readFileContent reads file content
func (v *Visualizer) readFileContent(filePath string) ([]string, error) {
	// 1. First, check if LineContents are saved in JSON data
	fileInfo, exists := v.stats.Files[filePath]
	if exists && len(fileInfo.LineContents) > 0 {
		// If already read content exists, use it
		lines := make([]string, len(fileInfo.LineContents))
		for lineNum, content := range fileInfo.LineContents {
			lines[lineNum-1] = content // LineContents are indexed from 1
		}
		return lines, nil
	}

	// 2. If repository path is set, check if it's actually existing
	if v.stats.RepositoryPath != "" {
		fullPath := filepath.Join(v.stats.RepositoryPath, filePath)
		// Check if path actually exists
		if _, err := os.Stat(fullPath); err == nil {
			return v.readFile(fullPath)
		}
		// If file is not found, return error message
		return []string{fmt.Sprintf("// File %s not found (%s)", filePath, fullPath)}, nil
	}

	// 3. If repository path is not set, return error message
	return []string{"// Repository path is not set, so cannot read file content"}, nil
}

func (v *Visualizer) readFile(fullPath string) ([]string, error) {
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// escapeSpecialChars escapes special characters in SVG
func escapeSpecialChars(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&apos;")
	return text
}

// generateSVGFileHeatmap generates SVG file heatmap
func (v *Visualizer) generateSVGFileHeatmap(outputPath, filePath string, fileInfo models.FileChangeInfo, color string) error {
	// Debug information
	fmt.Printf("Generating SVG file heatmap: %s -> %s\n", filePath, outputPath)

	// Check and create output directory
	outputDir := filepath.Dir(outputPath)
	if err := utils.EnsureDirectoryExists(outputDir); err != nil {
		return fmt.Errorf("Failed to create output directory: %s: %w", outputDir, err)
	}

	// Create file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Failed to create SVG file: %w", err)
	}
	defer file.Close()

	// Get file content
	var fileContent []string

	// If line contents are saved in JSON, use them
	if len(fileInfo.LineContents) > 0 {
		// Get maximum line number
		maxLine := 0
		for lineNum := range fileInfo.LineContents {
			if lineNum > maxLine {
				maxLine = lineNum
			}
		}

		// Get line contents from LineContents
		fileContent = make([]string, maxLine)
		for i := range fileContent {
			lineNum := i + 1
			if content, exists := fileInfo.LineContents[lineNum]; exists {
				fileContent[i] = content
			}
		}

		fmt.Printf("Read '%s' content from JSON (%d lines)\n", filePath, maxLine)
	} else {
		// If line contents are not saved in JSON, read file content from repository (backward compatibility)
		var readErr error
		fileContent, readErr = v.readFileContent(filePath)
		if readErr != nil {
			// If error occurs, use empty content and continue
			fileContent = []string{"// Unable to read file content"}
			fmt.Printf("Warning: Unable to read '%s' content: %v\n", filePath, readErr)
		}
	}

	// Drawing parameters
	const (
		lineHeight   = 20
		marginTop    = 100
		marginLeft   = 50
		marginRight  = 50
		marginBottom = 50
		maxWidth     = 1200
	)

	// Calculate canvas height based on line count
	contentHeight := len(fileContent) * lineHeight
	canvasHeight := contentHeight + marginTop + marginBottom
	if canvasHeight < 400 {
		canvasHeight = 400 // Minimum height
	}

	// Heatmap title to display (file path)
	title := filePath
	if len(title) > 80 {
		title = "..." + title[len(title)-77:]
	}

	// HTML escape processing
	escapedTitle := html.EscapeString(title)

	// Write SVG header
	fmt.Fprintf(file, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">
  <rect width="100%%" height="100%%" fill="#f0f0f0"/>
  <text x="600" y="40" font-size="24" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle" fill="#333333">File Heatmap</text>
  <text x="600" y="70" font-size="14" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle" fill="#333333">%s</text>
  <text x="600" y="90" font-family="Segoe UI, Helvetica, Arial" text-anchor="middle" fill="#666666">Changed: %d, Last updated: %s</text>

  <!-- Scrollable code view -->
  <g transform="translate(%d, %d)">
    <!-- Line number and code-->
`,
		maxWidth, canvasHeight,
		escapedTitle, fileInfo.ChangeCount, fileInfo.LastModified.Format("2006/01/02"),
		marginLeft, marginTop)

	// Draw each line of code
	for i, line := range fileContent {
		// Calculate line number
		lineNum := i + 1

		// This line's heat level (default is 0)
		heatLevel := 0
		if level, exists := fileInfo.LineChanges[lineNum]; exists {
			heatLevel = level
		}

		// Get color
		lineColor := "#ffffff" // Default is white
		if heatLevel > 0 {
			// Convert heat level to 0.0-1.0 range
			normalizedHeat := float64(heatLevel) / 10.0
			if normalizedHeat > 1.0 {
				normalizedHeat = 1.0
			}
			lineColor = GetHeatColor(normalizedHeat)
		}

		// Escape text
		escapedText := escapeSpecialChars(line)

		// Draw line
		fmt.Fprintf(file, `    <g>
      <rect x="0" y="%d" width="%d" height="%d" fill="%s" />
      <text x="-10" y="%d" text-anchor="end" font-size="12" font-family="Consolas, Menlo, monospace" fill="#666666">%d</text>
      <text x="5" y="%d" font-size="12" font-family="Consolas, Menlo, monospace" fill="#333333">%s</text>
    </g>
`,
			i*lineHeight, maxWidth-marginLeft-marginRight, lineHeight, lineColor,
			i*lineHeight+14, lineNum,
			i*lineHeight+14, escapedText)
	}

	// Write SVG footer
	fmt.Fprintf(file, `  </g>
</svg>`)

	return nil
}

// generateWebPFileHeatmap generates WebP file heatmap
func (v *Visualizer) generateWebPFileHeatmap(outputPath, filePath string, fileInfo models.FileChangeInfo, color string) error {
	// Check and create treemap directory
	if err := utils.EnsureDirectoryExists(filepath.Dir(outputPath)); err != nil {
		return fmt.Errorf("Failed to create treemap directory: %s: %w", filepath.Dir(outputPath), err)
	}

	// TODO: Actual WebP generation processing
	// Go language may require a different library to generate WebP format

	// Create dummy file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("Failed to create WebP file: %w", err)
	}
	file.Close()

	return nil
}
