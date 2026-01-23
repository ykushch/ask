package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

const maxHistory = 10
const maxContextChars = 4000

type historyEntry struct {
	command string
	output  string
}

var commandHistory []historyEntry

func addToHistory(command, output string) {
	if len(output) > 500 {
		output = output[:500]
	}
	commandHistory = append(commandHistory, historyEntry{command: command, output: output})
	for len(commandHistory) > maxHistory {
		commandHistory = commandHistory[1:]
	}
	for contextSize() > maxContextChars && len(commandHistory) > 1 {
		commandHistory = commandHistory[1:]
	}
}

func contextSize() int {
	total := 0
	for _, e := range commandHistory {
		total += len(e.command) + len(e.output)
	}
	return total
}

func formatHistory() string {
	if len(commandHistory) == 0 {
		return "No previous commands."
	}
	start := 0
	if len(commandHistory) > 5 {
		start = len(commandHistory) - 5
	}
	var lines []string
	for i, entry := range commandHistory[start:] {
		lines = append(lines, fmt.Sprintf("%d. $ %s", i+1, entry.command))
		if entry.output != "" {
			outputLines := strings.Split(strings.TrimSpace(entry.output), "\n")
			limit := 2
			if len(outputLines) < limit {
				limit = len(outputLines)
			}
			for _, ol := range outputLines[:limit] {
				lines = append(lines, "   "+ol)
			}
		}
	}
	return strings.Join(lines, "\n")
}

func buildPrompt(userInput string) string {
	cwd, _ := os.Getwd()
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	osName := runtime.GOOS

	history := formatHistory()

	return fmt.Sprintf(`You are a shell command translator. Convert the user's request into a shell command.
Current directory: %s
Operating system: %s
Shell: %s

Recent command history:
%s

Rules:
- Output ONLY the command, nothing else
- No explanations, no markdown, no backticks
- If unclear, make a reasonable assumption
- Prefer simple, common commands
- Use the command history for context (e.g., "do that again", "delete the file I just created")

User request: %s`, cwd, osName, shell, history, userInput)
}
