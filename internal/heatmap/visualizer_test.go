package heatmap

import (
	"testing"

	"github.com/repositoryheatmap/pkg/models"
)

func TestSortFilesByHeat(t *testing.T) {
	// Create test repository statistics
	stats := &models.RepositoryStats{
		RepositoryName: "test-repo",
		Files: map[string]models.FileChangeInfo{
			"file1.go": {
				FilePath:    "file1.go",
				ChangeCount: 10,
				HeatLevel:   0.5,
			},
			"file2.go": {
				FilePath:    "file2.go",
				ChangeCount: 5,
				HeatLevel:   0.25,
			},
			"file3.go": {
				FilePath:    "file3.go",
				ChangeCount: 20,
				HeatLevel:   1.0,
			},
			"file4.go": {
				FilePath:    "file4.go",
				ChangeCount: 1,
				HeatLevel:   0.05,
			},
		},
	}

	// Execute the function
	sortedFiles := sortFilesByHeat(stats)

	// Verify expected order
	expectedOrder := []string{"file3.go", "file1.go", "file2.go", "file4.go"}

	if len(sortedFiles) != len(expectedOrder) {
		t.Errorf("Expected %d files, got %d", len(expectedOrder), len(sortedFiles))
		return
	}

	for i, expected := range expectedOrder {
		if sortedFiles[i].FilePath != expected {
			t.Errorf("Expected file at position %d to be %s, got %s", i, expected, sortedFiles[i].FilePath)
		}
	}
}

func TestGetHeatColor(t *testing.T) {
	testCases := []struct {
		name          string
		heatLevel     float64
		expectedColor string
	}{
		{
			name:          "Very low heat (< 0.2)",
			heatLevel:     0.1,
			expectedColor: "#FFF5E6",
		},
		{
			name:          "Low heat (0.2-0.4)",
			heatLevel:     0.3,
			expectedColor: "#FFE0B2",
		},
		{
			name:          "Medium heat (0.4-0.6)",
			heatLevel:     0.5,
			expectedColor: "#FFAB66",
		},
		{
			name:          "Medium-high heat (0.6-0.8)",
			heatLevel:     0.7,
			expectedColor: "#FF8533",
		},
		{
			name:          "High heat (>= 0.8)",
			heatLevel:     0.9,
			expectedColor: "#E65C00",
		},
		{
			name:          "Invalid negative heat",
			heatLevel:     -0.1,
			expectedColor: "#FFF5E6",
		},
		{
			name:          "Invalid heat over 1.0",
			heatLevel:     1.1,
			expectedColor: "#E65C00",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetHeatColor(tc.heatLevel)
			if result != tc.expectedColor {
				t.Errorf("GetHeatColor(%f) = %s, expected %s", tc.heatLevel, result, tc.expectedColor)
			}
		})
	}
}

func TestBuildFileTreeFromFiles(t *testing.T) {
	// Create test file list
	files := []models.FileChangeInfo{
		{
			FilePath:    "src/main.go",
			ChangeCount: 10,
			HeatLevel:   0.5,
		},
		{
			FilePath:    "src/utils/helper.go",
			ChangeCount: 5,
			HeatLevel:   0.25,
		},
		{
			FilePath:    "README.md",
			ChangeCount: 3,
			HeatLevel:   0.15,
		},
	}

	// Create test visualizer
	v := NewVisualizer("output", "svg", &models.RepositoryStats{}, 10, "")

	// Execute tree construction
	tree := v.buildFileTreeFromFiles(files)

	// Verify root node
	if tree == nil {
		t.Fatal("Tree node should not be nil")
	}

	if tree.Name != "root" {
		t.Errorf("Expected root node name to be 'root', got %s", tree.Name)
	}

	// Verify number of child nodes (src directory and README.md file)
	if len(tree.Children) != 2 {
		t.Errorf("Expected root to have 2 children, got %d", len(tree.Children))
	}

	// Verify directory and child node structure
	var srcDir *TreeNode
	var readmeFile *TreeNode

	for _, child := range tree.Children {
		if child.Name == "src" && child.IsDir {
			srcDir = child
		} else if child.Name == "README.md" && !child.IsDir {
			readmeFile = child
		}
	}

	if srcDir == nil {
		t.Error("Expected to find a 'src' directory node")
	} else if len(srcDir.Children) != 2 {
		t.Errorf("Expected 'src' to have 2 children (main.go and utils dir), got %d", len(srcDir.Children))
	}

	if readmeFile == nil {
		t.Error("Expected to find a 'README.md' file node")
	}
}

func TestParseJSONPath(t *testing.T) {
	testCases := []struct {
		name          string
		jsonPath      string
		expectedDir   string
		expectedBase  string
		expectedError bool
	}{
		{
			name:          "Valid JSON path with directory",
			jsonPath:      "output/data/heatmap.json",
			expectedDir:   "output/data",
			expectedBase:  "heatmap.json",
			expectedError: false,
		},
		{
			name:          "JSON path without directory",
			jsonPath:      "heatmap.json",
			expectedDir:   ".",
			expectedBase:  "heatmap.json",
			expectedError: false,
		},
		{
			name:          "Empty path",
			jsonPath:      "",
			expectedDir:   "",
			expectedBase:  "",
			expectedError: true,
		},
		{
			name:          "Directory only",
			jsonPath:      "output/data/",
			expectedDir:   "",
			expectedBase:  "",
			expectedError: true,
		},
		{
			name:          "Invalid file extension",
			jsonPath:      "output/data/heatmap.txt",
			expectedDir:   "",
			expectedBase:  "",
			expectedError: true,
		},
		{
			name:          "Multiple JSON extensions",
			jsonPath:      "output/data/heatmap.json.json",
			expectedDir:   "",
			expectedBase:  "",
			expectedError: true,
		},
		{
			name:          "Path with spaces",
			jsonPath:      "output/my data/heat map.json",
			expectedDir:   "output/my data",
			expectedBase:  "heat map.json",
			expectedError: false,
		},
		{
			name:          "Absolute path",
			jsonPath:      "/usr/local/data/heatmap.json",
			expectedDir:   "/usr/local/data",
			expectedBase:  "heatmap.json",
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, base, err := parseJSONPath(tc.jsonPath)

			if tc.expectedError {
				if err == nil {
					t.Errorf("Expected error for path %q, but got nil", tc.jsonPath)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for path %q: %v", tc.jsonPath, err)
				}

				if dir != tc.expectedDir {
					t.Errorf("Expected directory %q, got %q", tc.expectedDir, dir)
				}

				if base != tc.expectedBase {
					t.Errorf("Expected base name %q, got %q", tc.expectedBase, base)
				}
			}
		})
	}
}
