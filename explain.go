package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
)

func buildExplainPrompt(command string) string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	osName := runtime.GOOS

	return fmt.Sprintf(`You are a shell command explainer. Given a command, output a short plain-text explanation.

Operating system: %s
Shell: %s

Output format rules:
- Plain text only. No markdown, no bold, no backticks, no bullet points, no numbered lists, no headings.
- First line: one sentence summarizing what the command does.
- Following lines: one line per flag/argument, indented with two spaces.
- Nothing else.

Example input: grep -rn "TODO" src/
Example output:
Searches for the text "TODO" in all files under src/ recursively, showing line numbers.
  -r: search recursively through directories
  -n: show line numbers in output
  "TODO": the pattern to search for
  src/: the directory to search in

Command to explain: %s`, osName, shell, command)
}

func explain(model, command string) (string, error) {
	prompt := buildExplainPrompt(command)
	result, err := generate(model, prompt)
	if err != nil {
		return "", err
	}
	return stripMarkdown(result), nil
}

var (
	boldRe      = regexp.MustCompile(`\*\*(.+?)\*\*`)
	backtickRe  = regexp.MustCompile("`([^`]+)`")
	headingRe   = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	numberedRe  = regexp.MustCompile(`(?m)^[ \t]*\d+\.[ \t]+`)
	bulletRe    = regexp.MustCompile(`(?m)^[ \t]*[-*][ \t]+`)
	emptyLineRe = regexp.MustCompile(`\n{3,}`)
)

func stripMarkdown(s string) string {
	s = boldRe.ReplaceAllString(s, "$1")
	s = backtickRe.ReplaceAllString(s, "$1")
	s = headingRe.ReplaceAllString(s, "")
	s = numberedRe.ReplaceAllString(s, "  ")
	s = bulletRe.ReplaceAllString(s, "  ")
	s = emptyLineRe.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}
