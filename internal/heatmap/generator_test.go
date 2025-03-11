package heatmap

import (
	"strings"
	"testing"

	"github.com/repositoryheatmap/pkg/models"
)

// Test function for English error messages
func TestGeneratorErrorMessages(t *testing.T) {
	// Create test Generator
	stats := &models.RepositoryStats{
		RepositoryName: "test-repo",
	}

	// Test generateJSONData method
	// Note: Private methods cannot be tested directly, need reflection or function replacement
	// Here we demonstrate testing English messages indirectly

	// Create an invalid path to trigger an error since we cannot mock OS file creation functions
	invalidDir := "/invalid/directory/that/does/not/exist"
	invalidGenerator := NewGenerator(invalidDir, stats)

	// Call Generate function which will trigger internal errors and return English error messages
	err := invalidGenerator.Generate()

	// Verify error occurred
	if err == nil {
		t.Error("Expected error for invalid directory, but got nil")
	}

	// Verify error messages are in English
	expectedMessages := []string{
		"failed to generate JSON data",
		"failed to create JSON file",
		"failed to encode JSON",
	}

	// Check if any of the expected messages are present
	found := false
	if err != nil {
		errMsg := err.Error()
		for _, msg := range expectedMessages {
			if strings.Contains(errMsg, msg) {
				found = true
				break
			}
		}

		// Verify message is in English
		if !found {
			t.Errorf("Error message does not contain any expected messages: %s", errMsg)
		}

		// Verify message starts with lowercase
		if len(errMsg) > 0 && errMsg[0] >= 'A' && errMsg[0] <= 'Z' {
			t.Errorf("Error message should start with lowercase: %s", errMsg)
		}
	}
}
