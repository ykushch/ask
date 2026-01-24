package main

import (
	"os"
	"strings"
	"testing"
)

func TestIsNaturalLanguage(t *testing.T) {
	naturalLanguage := []string{
		"find all go files",
		"show disk usage",
		"what time is it",
		"list running processes",
		"compress the folder",
		"how much memory is used",
	}

	shellCommands := []string{
		"ls",
		"pwd",
		"git status",
		"cd /tmp",
		"echo hello",
		"cat file.txt",
		"npm install",
		"python script.py",
		"./run.sh",
		"/usr/bin/env",
		"~/script.sh",
		"$HOME",
		"> output.txt",
		"docker ps",
	}

	for _, input := range naturalLanguage {
		if !isNaturalLanguage(input) {
			t.Errorf("isNaturalLanguage(%q) = false, want true", input)
		}
	}

	for _, input := range shellCommands {
		if isNaturalLanguage(input) {
			t.Errorf("isNaturalLanguage(%q) = true, want false", input)
		}
	}

	// Test ! prefix forces shell
	if isNaturalLanguage("!echo hello") {
		t.Error("isNaturalLanguage(\"!echo hello\") = true, want false (! prefix)")
	}

	// Test ? prefix forces explain (not natural language)
	if isNaturalLanguage("?ls -la") {
		t.Error("isNaturalLanguage(\"?ls -la\") = true, want false (? prefix)")
	}
	if isNaturalLanguage("?") {
		t.Error("isNaturalLanguage(\"?\") = true, want false (? prefix)")
	}

	// Test known commands without arguments
	if isNaturalLanguage("whoami") {
		t.Error("isNaturalLanguage(\"whoami\") = true, want false (known command)")
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/Documents", home + "/Documents"},
		{"~/.config", home + "/.config"},
		{"/tmp/file", "/tmp/file"},
		{"relative/path", "relative/path"},
		{"~", home},
	}

	for _, tt := range tests {
		got := expandHome(tt.input)
		if !strings.HasPrefix(got, tt.want) && got != tt.want {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
