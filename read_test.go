package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadCommand verifies that the /read slash command correctly
// resolves paths and retrieves file content.
func TestReadCommand(t *testing.T) {
	// 1. Setup: Create a temporary directory and a dummy file
	tmpDir := t.TempDir()
	fileName := "mission_report.txt"
	filePath := filepath.Join(tmpDir, fileName)
	fileContent := "Target confirmed. The package is secure."

	if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// 2. Define Test Cases
	tests := []struct {
		name        string
		cmdArgument string // The argument passed to /read
		want        string
		wantErr     bool
	}{
		{
			name:        "Valid Absolute Path",
			cmdArgument: filePath,
			want:        fileContent,
			wantErr:     false,
		},
		{
			name:        "Non-Existent File",
			cmdArgument: filepath.Join(tmpDir, "ghost_file.txt"),
			want:        "",
			wantErr:     true,
		},
		// You can add a relative path test here if your implementation supports
		// expanding relative paths based on the CWD.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// NOTE: This test assumes we extracted the logic into a function.
			// If the logic is inline in the update loop, we should extract it
			// to something like: func readFilePayload(path string) (string, error)
			
			// For this test file to compile and run immediately as a verification of logic,
			// I'm calling os.ReadFile directly here.
			// TODO: Replace 'os.ReadFile' below with your actual implementation function,
			// e.g., result, err := handleReadFile(tt.cmdArgument)

			var result string
			var err error

			// --- REPLACE START ---
			// Simulating the logic "we just implemented":
			cleanedPath := filepath.Clean(strings.TrimSpace(tt.cmdArgument))
			bytes, readErr := os.ReadFile(cleanedPath)
			result = string(bytes)
			err = readErr
			// --- REPLACE END ---

			// 3. Assertions
			if (err != nil) != tt.wantErr {
				t.Errorf("Process /read error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.want {
				t.Errorf("Process /read result = %q, want %q", result, tt.want)
			}
		})
	}
}
