package main

import (
	"strings"
	"testing"
)

func resetHistory() {
	commandHistory = nil
}

func TestContextSize(t *testing.T) {
	resetHistory()

	if got := contextSize(); got != 0 {
		t.Errorf("contextSize() empty = %d, want 0", got)
	}

	commandHistory = []historyEntry{{command: "ls", output: "file.go"}}
	if got := contextSize(); got != 9 { // "ls" (2) + "file.go" (7)
		t.Errorf("contextSize() single = %d, want 9", got)
	}

	commandHistory = []historyEntry{
		{command: "ls", output: "a"},
		{command: "pwd", output: "/tmp"},
	}
	// "ls"(2) + "a"(1) + "pwd"(3) + "/tmp"(4) = 10
	if got := contextSize(); got != 10 {
		t.Errorf("contextSize() multiple = %d, want 10", got)
	}
}

func TestFormatHistory(t *testing.T) {
	resetHistory()

	// Empty history
	got := formatHistory()
	if got != "No previous commands." {
		t.Errorf("formatHistory() empty = %q, want %q", got, "No previous commands.")
	}

	// Under 5 entries
	commandHistory = []historyEntry{
		{command: "ls", output: "file1\nfile2"},
		{command: "pwd", output: "/home/user"},
	}
	got = formatHistory()
	if !strings.Contains(got, "$ ls") {
		t.Error("formatHistory() should contain '$ ls'")
	}
	if !strings.Contains(got, "$ pwd") {
		t.Error("formatHistory() should contain '$ pwd'")
	}

	// Over 5 entries - only last 5 shown
	resetHistory()
	for i := 0; i < 7; i++ {
		commandHistory = append(commandHistory, historyEntry{command: "cmd" + string(rune('a'+i)), output: ""})
	}
	got = formatHistory()
	// Should not contain first two commands
	if strings.Contains(got, "cmda") {
		t.Error("formatHistory() should not show entries beyond last 5")
	}
	if strings.Contains(got, "cmdb") {
		t.Error("formatHistory() should not show entries beyond last 5")
	}
	// Should contain the last entries
	if !strings.Contains(got, "cmdg") {
		t.Error("formatHistory() should show last entry")
	}

	// Output truncation (only 2 lines shown)
	resetHistory()
	commandHistory = []historyEntry{
		{command: "cat big", output: "line1\nline2\nline3\nline4"},
	}
	got = formatHistory()
	if strings.Contains(got, "line3") {
		t.Error("formatHistory() should truncate output to 2 lines")
	}
}

func TestAddToHistory(t *testing.T) {
	resetHistory()

	// Test max history cap
	for i := 0; i < 15; i++ {
		addToHistory("cmd", "out")
	}
	if len(commandHistory) > maxHistory {
		t.Errorf("addToHistory() len = %d, want <= %d", len(commandHistory), maxHistory)
	}

	// Test output truncation at 500 chars
	resetHistory()
	longOutput := strings.Repeat("x", 1000)
	addToHistory("cmd", longOutput)
	if len(commandHistory[0].output) > 500 {
		t.Errorf("addToHistory() output len = %d, want <= 500", len(commandHistory[0].output))
	}

	// Test max context chars cap
	resetHistory()
	for i := 0; i < 20; i++ {
		addToHistory("command", strings.Repeat("a", 400))
	}
	if contextSize() > maxContextChars {
		t.Errorf("addToHistory() contextSize = %d, want <= %d", contextSize(), maxContextChars)
	}
}
