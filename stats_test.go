package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadStats_NewFile(t *testing.T) {
	// Use temp directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	stats, err := LoadStats()
	if err != nil {
		t.Fatalf("LoadStats failed: %v", err)
	}

	if stats.Version != statsVersion {
		t.Errorf("expected version %d, got %d", statsVersion, stats.Version)
	}
	if stats.Models == nil {
		t.Error("Models map should be initialized")
	}
	if stats.History == nil {
		t.Error("History slice should be initialized")
	}
}

func TestLoadStats_CorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create corrupt file
	askDir := filepath.Join(tmpDir, ".ask")
	os.MkdirAll(askDir, 0755)
	os.WriteFile(filepath.Join(askDir, "stats.json"), []byte("not json"), 0644)

	stats, err := LoadStats()
	if err != nil {
		t.Fatalf("LoadStats should handle corrupt file: %v", err)
	}

	if stats.Version != statsVersion {
		t.Errorf("expected fresh stats with version %d, got %d", statsVersion, stats.Version)
	}
}

func TestStats_Save(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	stats := &Stats{
		Version:   statsVersion,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Models:    map[string]int{"test-model": 5},
		History:   []HistoryEntry{},
	}

	err := stats.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(filepath.Join(tmpDir, ".ask", "stats.json"))
	if err != nil {
		t.Fatalf("Could not read saved file: %v", err)
	}

	var loaded Stats
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Saved file is not valid JSON: %v", err)
	}

	if loaded.Models["test-model"] != 5 {
		t.Errorf("expected model count 5, got %d", loaded.Models["test-model"])
	}
}

func TestStats_RecordInvocation(t *testing.T) {
	stats := &Stats{
		Version: statsVersion,
		Models:  make(map[string]int),
		History: []HistoryEntry{},
	}

	stats.RecordInvocation()
	stats.RecordInvocation()
	stats.RecordInvocation()

	if stats.Counters.TotalInvocations != 3 {
		t.Errorf("expected 3 invocations, got %d", stats.Counters.TotalInvocations)
	}
}

func TestStats_RecordOneshotCommand(t *testing.T) {
	stats := &Stats{
		Version: statsVersion,
		Models:  make(map[string]int),
		History: []HistoryEntry{},
	}

	stats.RecordOneshotCommand("llama3", "list files", "ls -la")

	if stats.Counters.OneshotCommands != 1 {
		t.Errorf("expected 1 oneshot command, got %d", stats.Counters.OneshotCommands)
	}
	if stats.Counters.CommandsGenerated != 1 {
		t.Errorf("expected 1 command generated, got %d", stats.Counters.CommandsGenerated)
	}
	if stats.Models["llama3"] != 1 {
		t.Errorf("expected model count 1, got %d", stats.Models["llama3"])
	}
	if len(stats.History) != 1 {
		t.Errorf("expected 1 history entry, got %d", len(stats.History))
	}
	if stats.History[0].Mode != "oneshot" {
		t.Errorf("expected mode 'oneshot', got %s", stats.History[0].Mode)
	}
}

func TestStats_RecordExecution(t *testing.T) {
	stats := &Stats{
		Version: statsVersion,
		Models:  make(map[string]int),
		History: []HistoryEntry{
			{Mode: "oneshot", Executed: false},
		},
	}

	stats.RecordExecution()

	if stats.Counters.CommandsExecuted != 1 {
		t.Errorf("expected 1 command executed, got %d", stats.Counters.CommandsExecuted)
	}
	if !stats.History[0].Executed {
		t.Error("expected last history entry to be marked as executed")
	}
}

func TestStats_SizeTruncation(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	stats := &Stats{
		Version:   statsVersion,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Models:    make(map[string]int),
		History:   []HistoryEntry{},
	}

	// Add many history entries to exceed size limit
	for i := 0; i < 2000; i++ {
		stats.History = append(stats.History, HistoryEntry{
			Timestamp: time.Now(),
			Mode:      "oneshot",
			Model:     "test-model",
			Query:     "this is a test query that is somewhat long",
			Command:   "echo 'this is a test command that is also somewhat long'",
			Executed:  true,
		})
	}

	err := stats.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Check file size
	info, err := os.Stat(filepath.Join(tmpDir, ".ask", "stats.json"))
	if err != nil {
		t.Fatalf("Could not stat file: %v", err)
	}

	if info.Size() > maxStatsFileSize {
		t.Errorf("file size %d exceeds max %d", info.Size(), maxStatsFileSize)
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer string", 10, "this is..."},
	}

	for _, tt := range tests {
		result := truncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}
