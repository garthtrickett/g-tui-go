package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func logToFile(message string) {
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Can't log, so just print to stderr
		fmt.Fprintf(os.Stderr, "failed to open log file: %v", err)
		return
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "%s\n", message); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write to log file: %v", err)
	}
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for i := 1; i < len(strs); i++ {
		for !strings.HasPrefix(strs[i], prefix) {
			prefix = prefix[:len(prefix)-1]
			if len(prefix) == 0 {
				return ""
			}
		}
	}
	return prefix
}

func appendFile(outFile *os.File, path string) {
	base := filepath.Base(path)
	if base == "a.txt" || base == "concat.json" {
		return
	}
	logToFile("Processing path: " + path)
	fileContent, err := os.ReadFile(path)
	if err != nil {
		logToFile(fmt.Sprintf("Error reading file %s: %v", path, err))
		return
	}

	header := fmt.Sprintf("--- %s ---\n", path)
	footer := fmt.Sprintf("\n--- END %s ---\n", path)

	outFile.WriteString(header)
	outFile.Write(fileContent)
	outFile.WriteString(footer)
}
