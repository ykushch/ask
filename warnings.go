package main

import (
	"fmt"
	"os"
	"regexp"
)

type dangerousPattern struct {
	pattern  string
	message  string
	severity string
}

type compiledPattern struct {
	re       *regexp.Regexp
	message  string
	severity string
}

var dangerousPatterns = []dangerousPattern{
	// File deletion - broad targets
	{`\brm\s+.*-[^\s]*r[^\s]*\s+[/~*]`, "Recursive deletion targeting a broad path", "high"},
	{`\brm\s+.*-[^\s]*r`, "Recursive file deletion", "medium"},
	{`\brm\s+(-[^\s]*\s+)*\*`, "Deleting files with wildcard", "high"},
	{`\brm\s+-[^\s]*f`, "Force file deletion (no confirmation)", "medium"},
	{`\bsudo\s+rm\b`, "Deleting files as root", "high"},

	// Disk/filesystem
	{`\bdd\s+if=`, "Direct disk write — may overwrite partitions", "high"},
	{`\bmkfs\b`, "Formatting a filesystem", "high"},

	// Permissions
	{`\bchmod\s+777\b`, "Setting world-writable permissions", "medium"},
	{`\bchmod\s+.*-R\b`, "Recursive permission change", "medium"},

	// Git destructive
	{`\bgit\s+reset\s+--hard\b`, "Discards all uncommitted changes", "medium"},
	{`\bgit\s+push\s+.*--force\b`, "Force push may overwrite remote history", "high"},
	{`\bgit\s+push\s+.*\s-f\b`, "Force push may overwrite remote history", "high"},
	{`\bgit\s+clean\s+.*-[^\s]*f`, "Removes untracked files permanently", "medium"},

	// Process
	{`\bkill\s+-9\b`, "Forceful process termination (no cleanup)", "medium"},

	// File truncation (redirect at start of command)
	{`^\s*>`, "File truncation — will erase file contents", "high"},

	// SQL destructive
	{`(?i)\bDROP\s+TABLE\b`, "Drops a database table permanently", "high"},
	{`(?i)\bDROP\s+DATABASE\b`, "Drops an entire database permanently", "high"},
	{`(?i)\bTRUNCATE\b`, "Truncates table data permanently", "high"},

	// Fork bomb
	{`:\(\)\s*\{.*\|.*\}`, "Potential fork bomb — may crash the system", "high"},
}

var compiledPatterns []compiledPattern

func init() {
	for _, dp := range dangerousPatterns {
		compiledPatterns = append(compiledPatterns, compiledPattern{
			re:       regexp.MustCompile(dp.pattern),
			message:  dp.message,
			severity: dp.severity,
		})
	}
}

func checkDangerousCommand(cmd string) []compiledPattern {
	var matched []compiledPattern
	for _, cp := range compiledPatterns {
		if cp.re.MatchString(cmd) {
			matched = append(matched, cp)
		}
	}
	return matched
}

func printWarnings(warnings []compiledPattern) {
	for _, w := range warnings {
		var color string
		if w.severity == "high" {
			color = "\033[31m" // red
		} else {
			color = "\033[33m" // yellow
		}
		fmt.Fprintf(os.Stderr, "%s  ⚠ Warning: %s\033[0m\n", color, w.message)
	}
}

func warnIfDangerous(cmd string) {
	warnings := checkDangerousCommand(cmd)
	printWarnings(warnings)
}
