package utils

import (
	"testing"
)

func TestSanitizePath(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty path",
			input:    "",
			expected: "unknown_file",
		},
		{
			name:     "Simple file name",
			input:    "test.txt",
			expected: "test.txt",
		},
		{
			name:     "File with spaces",
			input:    "test file.txt",
			expected: "test_file.txt",
		},
		{
			name:     "File with special characters",
			input:    "test@#$%.txt",
			expected: "test____.txt",
		},
		{
			name:     "File with unicode characters",
			input:    "テスト.txt",
			expected: "___.txt",
		},
		{
			name:     "README file with underscore",
			input:    "README_zh.md",
			expected: "README_zh.md",
		},
		{
			name:     "Python file with underscore",
			input:    "test_utils.py",
			expected: "test_utils.py",
		},
		{
			name:     "Directory path with special chars",
			input:    "dir with spaces/sub-dir/@file.txt",
			expected: "dir_with_spaces/sub-dir/_file.txt",
		},
		{
			name:     "Windows path style",
			input:    "C:\\Users\\test\\file.txt",
			expected: "C__Users_test_file.txt",
		},
		{
			name:     "Complex path with special files",
			input:    "src/modules/README_zh.md",
			expected: "src/modules/README_zh.md",
		},
		{
			name:     "Path with backslash and unicode",
			input:    "src\\テスト\\file.txt",
			expected: "src_____file.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := SanitizePath(tc.input)
			if result != tc.expected {
				t.Errorf("SanitizePath(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestReplaceUnsafeChars(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "With safe characters only",
			input:    "abcABC123_-.",
			expected: "abcABC123_-.",
		},
		{
			name:     "With unsafe characters",
			input:    "hello world!@#$%^&*()",
			expected: "hello_world__________",
		},
		{
			name:     "With unicode characters",
			input:    "こんにちは世界",
			expected: "_______",
		},
		{
			name:     "With mixed safe and unsafe characters",
			input:    "test-file_name.txt!",
			expected: "test-file_name.txt_",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := replaceUnsafeChars(tc.input)
			if result != tc.expected {
				t.Errorf("replaceUnsafeChars(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestEnsureDirectoryExists(t *testing.T) {
	// Create test directory
	testDir := t.TempDir()

	// Test creating a non-existent directory
	newDir := testDir + "/new_directory"
	err := EnsureDirectoryExists(newDir)
	if err != nil {
		t.Errorf("Failed to create directory: %v", err)
	}

	// Test with an existing directory
	err = EnsureDirectoryExists(newDir)
	if err != nil {
		t.Errorf("Failed with existing directory: %v", err)
	}
}
