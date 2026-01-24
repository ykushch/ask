package main

import (
	"strings"
	"testing"
)

func TestBuildExplainPrompt(t *testing.T) {
	prompt := buildExplainPrompt("ls -la")

	if !strings.Contains(prompt, "ls -la") {
		t.Error("buildExplainPrompt() should contain the command")
	}
	if !strings.Contains(prompt, "shell command explainer") {
		t.Error("buildExplainPrompt() should contain explainer role")
	}
	if !strings.Contains(prompt, "Operating system:") {
		t.Error("buildExplainPrompt() should contain OS info")
	}
	if !strings.Contains(prompt, "Shell:") {
		t.Error("buildExplainPrompt() should contain shell info")
	}
	if !strings.Contains(prompt, "flag") {
		t.Error("buildExplainPrompt() should mention flags in rules")
	}
}

func TestBuildExplainPromptComplex(t *testing.T) {
	cmd := "find . -name '*.go' -exec grep 'func main' {} +"
	prompt := buildExplainPrompt(cmd)

	if !strings.Contains(prompt, cmd) {
		t.Error("buildExplainPrompt() should contain the full complex command")
	}
}

func TestStripMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bold markers",
			input: "The **ls** command lists **files**",
			want:  "The ls command lists files",
		},
		{
			name:  "backticks",
			input: "Use `ls -la` to list files",
			want:  "Use ls -la to list files",
		},
		{
			name:  "headings",
			input: "### Overall Intent\nSome text",
			want:  "Overall Intent\nSome text",
		},
		{
			name:  "numbered list",
			input: "1. First item\n2. Second item\n3. Third item",
			want:  "First item\n  Second item\n  Third item",
		},
		{
			name:  "bullet list",
			input: "- First item\n- Second item",
			want:  "First item\n  Second item",
		},
		{
			name:  "multiple empty lines",
			input: "Line one\n\n\n\nLine two",
			want:  "Line one\n\nLine two",
		},
		{
			name:  "combined markdown",
			input: "### Command\n\n**ls** `-la`\n\n- shows files\n- includes hidden",
			want:  "Command\n\nls -la\n\n  shows files\n  includes hidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("stripMarkdown() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestBuildExplainPromptDiffersFromBuildPrompt(t *testing.T) {
	input := "ls -la"
	explainPrompt := buildExplainPrompt(input)
	translatePrompt := buildPrompt(input)

	if explainPrompt == translatePrompt {
		t.Error("explain prompt and translate prompt should be different")
	}
	if strings.Contains(explainPrompt, "Output ONLY the command") {
		t.Error("explain prompt should not contain translation rules")
	}
	if strings.Contains(translatePrompt, "Explain") {
		t.Error("translate prompt should not contain explain rules")
	}
}
