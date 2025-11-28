package main

import (
	"strings"
	"testing"
)

// parseCommand is a helper function we assume exists or should exist 
// to decouple logic from the huge Update() switch statement.
// If you don't have this, extracting it makes testing 10x easier.
func parseCommand(input string) (string, string, bool) {
	trimmed := strings.TrimSpace(input)
	if !strings.HasPrefix(trimmed, "/") {
		return "", "", false
	}

	parts := strings.SplitN(trimmed, " ", 2)
	cmd := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = parts[1]
	}
	return cmd, arg, true
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCmd   string
		wantArg   string
		wantIsCmd bool
	}{
		{
			name:      "Normal Read Command",
			input:     "/read /tmp/test.txt",
			wantCmd:   "/read",
			wantArg:   "/tmp/test.txt",
			wantIsCmd: true,
		},
		{
			name:      "Command Without Arg",
			input:     "/help",
			wantCmd:   "/help",
			wantArg:   "",
			wantIsCmd: true,
		},
		{
			name:      "Regular Chat Message",
			input:     "Can you read this?",
			wantCmd:   "",
			wantArg:   "",
			wantIsCmd: false,
		},
		{
			name:      "Whitespace Trimming",
			input:     "   /read config.json   ",
			wantCmd:   "/read",
			wantArg:   "config.json", // Should be trimmed by parser logic usually
			wantIsCmd: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, arg, isCmd := parseCommand(tt.input)
			
			// If your parser doesn't trim args automatically, 
			// you might need to adjust the expectation or the implementation.
			arg = strings.TrimSpace(arg) 

			if isCmd != tt.wantIsCmd {
				t.Errorf("got isCmd %v, want %v", isCmd, tt.wantIsCmd)
			}
			if cmd != tt.wantCmd {
				t.Errorf("got cmd %q, want %q", cmd, tt.wantCmd)
			}
			if arg != tt.wantArg {
				t.Errorf("got arg %q, want %q", arg, tt.wantArg)
			}
		})
	}
}
