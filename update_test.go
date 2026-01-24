package main

import (
	"runtime"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input               string
		major, minor, patch int
		ok                  bool
	}{
		{"0.1.0", 0, 1, 0, true},
		{"v1.2.3", 1, 2, 3, true},
		{"10.20.30", 10, 20, 30, true},
		{"0.0.0", 0, 0, 0, true},
		{"1.2", 0, 0, 0, false},
		{"1.2.3.4", 0, 0, 0, false},
		{"abc", 0, 0, 0, false},
		{"1.2.x", 0, 0, 0, false},
		{"", 0, 0, 0, false},
		{"v", 0, 0, 0, false},
	}

	for _, tt := range tests {
		major, minor, patch, ok := parseVersion(tt.input)
		if major != tt.major || minor != tt.minor || patch != tt.patch || ok != tt.ok {
			t.Errorf("parseVersion(%q) = (%d, %d, %d, %v), want (%d, %d, %d, %v)",
				tt.input, major, minor, patch, ok, tt.major, tt.minor, tt.patch, tt.ok)
		}
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"0.2.0", "0.1.0", true},
		{"1.0.0", "0.9.9", true},
		{"0.1.1", "0.1.0", true},
		{"0.1.0", "0.1.0", false},
		{"0.1.0", "0.2.0", false},
		{"0.0.9", "0.1.0", false},
		{"v1.0.0", "v0.9.0", true},
		{"invalid", "0.1.0", false},
		{"0.1.0", "invalid", false},
		{"2.0.0", "1.99.99", true},
	}

	for _, tt := range tests {
		got := isNewer(tt.latest, tt.current)
		if got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

func TestBinaryAssetName(t *testing.T) {
	name := binaryAssetName()
	expected := "ask-" + runtime.GOOS + "-" + runtime.GOARCH
	if name != expected {
		t.Errorf("binaryAssetName() = %q, want %q", name, expected)
	}
}

func TestDownloadURL(t *testing.T) {
	release := &githubRelease{
		TagName: "v0.2.0",
		Assets: []githubAsset{
			{Name: "ask-linux-amd64", BrowserDownloadURL: "https://example.com/ask-linux-amd64"},
			{Name: "ask-darwin-arm64", BrowserDownloadURL: "https://example.com/ask-darwin-arm64"},
			{Name: "ask-darwin-amd64", BrowserDownloadURL: "https://example.com/ask-darwin-amd64"},
		},
	}

	url, err := downloadURL(release)
	if err != nil {
		t.Fatalf("downloadURL() unexpected error: %v", err)
	}

	expectedName := binaryAssetName()
	expectedURL := "https://example.com/" + expectedName
	if url != expectedURL {
		t.Errorf("downloadURL() = %q, want %q", url, expectedURL)
	}

	// Test missing platform
	emptyRelease := &githubRelease{
		TagName: "v0.2.0",
		Assets:  []githubAsset{{Name: "ask-windows-amd64", BrowserDownloadURL: "https://example.com/ask-windows-amd64"}},
	}
	_, err = downloadURL(emptyRelease)
	if err == nil {
		t.Error("downloadURL() expected error for missing platform, got nil")
	}
}
