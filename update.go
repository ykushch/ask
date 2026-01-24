package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	repoOwner = "ykushch"
	repoName  = "ask"
	releaseURL = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// parseVersion splits "0.1.0" into [0, 1, 0].
func parseVersion(v string) (major, minor, patch int, ok bool) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	patch, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	return major, minor, patch, true
}

// isNewer returns true if latest is a higher version than current.
func isNewer(latest, current string) bool {
	lMaj, lMin, lPat, lok := parseVersion(latest)
	cMaj, cMin, cPat, cok := parseVersion(current)
	if !lok || !cok {
		return false
	}
	if lMaj != cMaj {
		return lMaj > cMaj
	}
	if lMin != cMin {
		return lMin > cMin
	}
	return lPat > cPat
}

// fetchLatestRelease queries GitHub for the latest release info.
func fetchLatestRelease() (*githubRelease, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}
	return &release, nil
}

// binaryAssetName returns the expected asset name for the current platform.
func binaryAssetName() string {
	return fmt.Sprintf("ask-%s-%s", runtime.GOOS, runtime.GOARCH)
}

// downloadURL finds the download URL for the current platform from the release assets.
func downloadURL(release *githubRelease) (string, error) {
	want := binaryAssetName()
	for _, asset := range release.Assets {
		if asset.Name == want {
			return asset.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, release.TagName)
}

// selfUpdate downloads the latest release binary and replaces the current executable.
func selfUpdate() error {
	release, err := fetchLatestRelease()
	if err != nil {
		return err
	}

	if !isNewer(release.TagName, version) {
		fmt.Printf("already up to date (v%s)\n", version)
		return nil
	}

	url, err := downloadURL(release)
	if err != nil {
		return err
	}

	fmt.Printf("updating ask: v%s → %s\n", version, release.TagName)

	// Download the new binary
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}

	// Write to a temp file next to the executable, then atomic rename
	tmpPath := execPath + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("cannot create temp file: %w", err)
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write update: %w", err)
	}
	tmpFile.Close()

	// Replace the old binary
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("updated to %s\n", release.TagName)
	return nil
}

const updateCheckInterval = 24 * time.Hour

func updateCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ask", "last-update-check")
}

// readCachedVersion returns the cached latest version if the cache is fresh.
func readCachedVersion() (string, bool) {
	path := updateCachePath()
	info, err := os.Stat(path)
	if err != nil {
		return "", false
	}
	if time.Since(info.ModTime()) > updateCheckInterval {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(data)), true
}

func writeCachedVersion(tag string) {
	path := updateCachePath()
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(tag), 0644)
}

// backgroundVersionCheck runs a non-blocking version check and prints a notice
// if a newer version is available. Call this as a goroutine.
func backgroundVersionCheck(done chan<- string) {
	if cached, ok := readCachedVersion(); ok {
		if isNewer(cached, version) {
			done <- cached
		} else {
			done <- ""
		}
		return
	}

	release, err := fetchLatestRelease()
	if err != nil {
		done <- ""
		return
	}
	writeCachedVersion(release.TagName)
	if isNewer(release.TagName, version) {
		done <- release.TagName
	} else {
		done <- ""
	}
}
