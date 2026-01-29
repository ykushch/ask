package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	statsFileName    = "stats.json"
	maxStatsFileSize = 100 * 1024 // 100KB
	statsVersion     = 1
)

type Counters struct {
	TotalInvocations    int `json:"total_invocations"`
	CommandsGenerated   int `json:"commands_generated"`
	CommandsExecuted    int `json:"commands_executed"`
	ExplainCalls        int `json:"explain_calls"`
	InteractiveSessions int `json:"interactive_sessions"`
	OneshotCommands     int `json:"oneshot_commands"`
}

type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Mode      string    `json:"mode"` // "oneshot", "interactive", "explain"
	Model     string    `json:"model"`
	Query     string    `json:"query"`
	Command   string    `json:"command,omitempty"`
	Executed  bool      `json:"executed"`
}

type Stats struct {
	Version   int            `json:"version"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Counters  Counters       `json:"counters"`
	Models    map[string]int `json:"models"`
	History   []HistoryEntry `json:"history"`
}

func statsFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ask", statsFileName)
}

func LoadStats() (*Stats, error) {
	path := statsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return new empty stats
			return &Stats{
				Version:   statsVersion,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Models:    make(map[string]int),
				History:   []HistoryEntry{},
			}, nil
		}
		return nil, fmt.Errorf("reading stats file: %w", err)
	}

	var stats Stats
	if err := json.Unmarshal(data, &stats); err != nil {
		// Corrupt file, start fresh
		return &Stats{
			Version:   statsVersion,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Models:    make(map[string]int),
			History:   []HistoryEntry{},
		}, nil
	}

	if stats.Models == nil {
		stats.Models = make(map[string]int)
	}
	if stats.History == nil {
		stats.History = []HistoryEntry{}
	}

	return &stats, nil
}

func (s *Stats) Save() error {
	s.UpdatedAt = time.Now()

	// Ensure directory exists
	dir := filepath.Dir(statsFilePath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating stats directory: %w", err)
	}

	// Marshal and check size
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling stats: %w", err)
	}

	// If too large, truncate history
	for len(data) > maxStatsFileSize && len(s.History) > 10 {
		// Keep newest 50% of history
		half := len(s.History) / 2
		s.History = s.History[half:]
		data, _ = json.MarshalIndent(s, "", "  ")
	}

	if err := os.WriteFile(statsFilePath(), data, 0644); err != nil {
		return fmt.Errorf("writing stats file: %w", err)
	}

	return nil
}

func (s *Stats) RecordInvocation() {
	s.Counters.TotalInvocations++
}

func (s *Stats) RecordOneshotCommand(model, query, command string) {
	s.Counters.OneshotCommands++
	s.Counters.CommandsGenerated++
	s.Models[model]++

	s.History = append(s.History, HistoryEntry{
		Timestamp: time.Now(),
		Mode:      "oneshot",
		Model:     model,
		Query:     truncateString(query, 100),
		Command:   truncateString(command, 200),
		Executed:  false,
	})
}

func (s *Stats) RecordInteractiveSession() {
	s.Counters.InteractiveSessions++
}

func (s *Stats) RecordInteractiveCommand(model, query, command string) {
	s.Counters.CommandsGenerated++
	s.Models[model]++

	s.History = append(s.History, HistoryEntry{
		Timestamp: time.Now(),
		Mode:      "interactive",
		Model:     model,
		Query:     truncateString(query, 100),
		Command:   truncateString(command, 200),
		Executed:  false,
	})
}

func (s *Stats) RecordExecution() {
	s.Counters.CommandsExecuted++
	// Mark last history entry as executed
	if len(s.History) > 0 {
		s.History[len(s.History)-1].Executed = true
	}
}

func (s *Stats) RecordExplain(model string) {
	s.Counters.ExplainCalls++
	s.Models[model]++
}

func ShowStats() error {
	stats, err := LoadStats()
	if err != nil {
		return err
	}

	fmt.Println("ask usage statistics")
	fmt.Println("────────────────────")

	c := stats.Counters
	fmt.Printf("Total invocations:     %d\n", c.TotalInvocations)
	fmt.Printf("Commands generated:    %d\n", c.CommandsGenerated)
	if c.CommandsGenerated > 0 {
		pct := float64(c.CommandsExecuted) / float64(c.CommandsGenerated) * 100
		fmt.Printf("Commands executed:     %d  (%.0f%%)\n", c.CommandsExecuted, pct)
	} else {
		fmt.Printf("Commands executed:     %d\n", c.CommandsExecuted)
	}
	fmt.Printf("Explain calls:         %d\n", c.ExplainCalls)
	fmt.Printf("Interactive sessions:  %d\n", c.InteractiveSessions)
	fmt.Printf("One-shot commands:     %d\n", c.OneshotCommands)

	if len(stats.Models) > 0 {
		fmt.Println()
		fmt.Println("Model usage:")

		// Sort models by usage count (descending)
		type modelCount struct {
			name  string
			count int
		}
		var models []modelCount
		total := 0
		for name, count := range stats.Models {
			models = append(models, modelCount{name, count})
			total += count
		}
		sort.Slice(models, func(i, j int) bool {
			return models[i].count > models[j].count
		})

		for _, m := range models {
			pct := float64(m.count) / float64(total) * 100
			fmt.Printf("  %-20s %d  (%.0f%%)\n", m.name, m.count, pct)
		}
	}

	// Show file info
	fmt.Println()
	path := statsFilePath()
	if info, err := os.Stat(path); err == nil {
		sizeKB := float64(info.Size()) / 1024
		fmt.Printf("Stats file: %s (%.1fKB)\n", path, sizeKB)
	} else {
		fmt.Printf("Stats file: %s (not created yet)\n", path)
	}
	fmt.Printf("Tracking since: %s\n", stats.CreatedAt.Format("2006-01-02"))

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
