package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectProjectsEmpty(t *testing.T) {
	dir := t.TempDir()
	projects := detectProjects(dir)
	if len(projects) != 0 {
		t.Errorf("detectProjects(empty) = %d projects, want 0", len(projects))
	}
}

func TestDetectProjectsGo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)

	projects := detectProjects(dir)
	if len(projects) != 1 {
		t.Fatalf("detectProjects(go.mod) = %d projects, want 1", len(projects))
	}
	if projects[0].Name != "Go" {
		t.Errorf("project name = %q, want %q", projects[0].Name, "Go")
	}
}

func TestDetectProjectsNode(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)

	projects := detectProjects(dir)
	if len(projects) != 1 {
		t.Fatalf("detectProjects(package.json) = %d projects, want 1", len(projects))
	}
	if projects[0].Name != "Node.js" {
		t.Errorf("project name = %q, want %q", projects[0].Name, "Node.js")
	}
}

func TestDetectProjectsMultiple(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine"), 0644)
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("all:"), 0644)

	projects := detectProjects(dir)
	if len(projects) != 3 {
		t.Errorf("detectProjects(multiple) = %d projects, want 3", len(projects))
	}

	names := make(map[string]bool)
	for _, p := range projects {
		names[p.Name] = true
	}
	for _, expected := range []string{"Go", "Docker", "Make-based"} {
		if !names[expected] {
			t.Errorf("expected %q to be detected", expected)
		}
	}
}

func TestDetectProjectsPythonDedup(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask"), 0644)
	os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]"), 0644)
	os.WriteFile(filepath.Join(dir, "setup.py"), []byte(""), 0644)

	projects := detectProjects(dir)
	pythonCount := 0
	for _, p := range projects {
		if p.Name == "Python" {
			pythonCount++
		}
	}
	if pythonCount != 1 {
		t.Errorf("Python detected %d times, want 1", pythonCount)
	}
}

func TestDetectProjectsDotNetGlob(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "MyApp.csproj"), []byte("<Project>"), 0644)

	projects := detectProjects(dir)
	if len(projects) != 1 || projects[0].Name != ".NET" {
		t.Errorf("detectProjects(*.csproj) failed, got %v", projects)
	}
}

func TestFormatProjectInfoEmpty(t *testing.T) {
	dir := t.TempDir()
	info := formatProjectInfo(dir)
	if info != "" {
		t.Errorf("formatProjectInfo(empty) = %q, want empty", info)
	}
}

func TestFormatProjectInfoSingle(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)

	info := formatProjectInfo(dir)
	if info == "" {
		t.Error("formatProjectInfo(package.json) returned empty")
	}
	if !strings.Contains(info, "Node.js") {
		t.Errorf("formatProjectInfo = %q, missing Node.js", info)
	}
	if !strings.Contains(info, "npm") {
		t.Errorf("formatProjectInfo = %q, missing npm tool", info)
	}
}

func TestFormatProjectInfoMultiple(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]"), 0644)
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM rust"), 0644)

	info := formatProjectInfo(dir)
	if !strings.Contains(info, "Rust") {
		t.Errorf("formatProjectInfo = %q, missing Rust", info)
	}
	if !strings.Contains(info, "Docker") {
		t.Errorf("formatProjectInfo = %q, missing Docker", info)
	}
}
