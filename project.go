package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectType represents a detected project type with its common tools
type ProjectType struct {
	Name  string
	Tools []string
}

// signatureFile maps a file pattern to a project type
type signatureFile struct {
	pattern     string
	projectType ProjectType
}

// signatureFiles is the ordered list of signature files to check.
// Exact matches are checked before glob patterns.
var signatureFiles = []signatureFile{
	// Go
	{"go.mod", ProjectType{"Go", []string{"go build", "go test", "go run", "go mod"}}},

	// Rust
	{"Cargo.toml", ProjectType{"Rust", []string{"cargo build", "cargo test", "cargo run"}}},

	// Node.js
	{"package.json", ProjectType{"Node.js", []string{"npm", "npx", "node", "yarn"}}},

	// Python (multiple signatures, same type)
	{"requirements.txt", ProjectType{"Python", []string{"pip", "python", "pytest"}}},
	{"pyproject.toml", ProjectType{"Python", []string{"pip", "python", "pytest", "poetry"}}},
	{"setup.py", ProjectType{"Python", []string{"pip", "python", "pytest"}}},

	// Ruby
	{"Gemfile", ProjectType{"Ruby", []string{"bundle", "rake", "ruby"}}},

	// Java
	{"pom.xml", ProjectType{"Java (Maven)", []string{"mvn", "java"}}},
	{"build.gradle", ProjectType{"Java (Gradle)", []string{"gradle", "./gradlew", "java"}}},
	{"build.gradle.kts", ProjectType{"Kotlin (Gradle)", []string{"gradle", "./gradlew", "kotlin"}}},

	// Make
	{"Makefile", ProjectType{"Make-based", []string{"make"}}},

	// Docker
	{"Dockerfile", ProjectType{"Docker", []string{"docker", "docker build"}}},
	{"docker-compose.yml", ProjectType{"Docker Compose", []string{"docker-compose", "docker compose"}}},
	{"docker-compose.yaml", ProjectType{"Docker Compose", []string{"docker-compose", "docker compose"}}},

	// .NET (glob patterns)
	{"*.csproj", ProjectType{".NET", []string{"dotnet build", "dotnet run", "dotnet test"}}},
	{"*.sln", ProjectType{".NET", []string{"dotnet build", "dotnet run", "dotnet test"}}},
}

// detectProjects checks for signature files in dir and returns detected project types.
// Multiple types can be detected if multiple signature files are present.
func detectProjects(dir string) []ProjectType {
	var detected []ProjectType
	seen := make(map[string]bool)

	for _, sig := range signatureFiles {
		if seen[sig.projectType.Name] {
			continue
		}

		if strings.ContainsAny(sig.pattern, "*?[") {
			// Glob pattern
			matches, err := filepath.Glob(filepath.Join(dir, sig.pattern))
			if err == nil && len(matches) > 0 {
				detected = append(detected, sig.projectType)
				seen[sig.projectType.Name] = true
			}
		} else {
			// Exact file name
			path := filepath.Join(dir, sig.pattern)
			if _, err := os.Stat(path); err == nil {
				detected = append(detected, sig.projectType)
				seen[sig.projectType.Name] = true
			}
		}
	}

	return detected
}

// formatProjectInfo returns a formatted string describing detected projects,
// or an empty string if no projects are detected.
func formatProjectInfo(dir string) string {
	projects := detectProjects(dir)
	if len(projects) == 0 {
		return ""
	}

	var parts []string
	for _, p := range projects {
		parts = append(parts, fmt.Sprintf("%s (tools: %s)", p.Name, strings.Join(p.Tools, ", ")))
	}

	return fmt.Sprintf("Detected project type(s): %s", strings.Join(parts, "; "))
}
