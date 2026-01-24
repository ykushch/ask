package main

import "testing"

func TestCheckDangerousCommand(t *testing.T) {
	dangerous := []struct {
		cmd         string
		expectMin   int
		desc        string
	}{
		{"rm -rf /", 2, "recursive delete root"},
		{"rm -rf ~", 2, "recursive delete home"},
		{"rm -rf *", 2, "recursive delete wildcard"},
		{"rm -r /tmp/foo", 1, "recursive delete"},
		{"rm -f *.log", 1, "force delete"},
		{"rm *", 1, "delete wildcard"},
		{"sudo rm -rf /var", 2, "sudo recursive delete"},
		{"sudo rm file.txt", 1, "sudo rm"},
		{"dd if=/dev/zero of=/dev/sda", 1, "disk overwrite"},
		{"mkfs.ext4 /dev/sda1", 1, "format filesystem"},
		{"chmod 777 /etc/passwd", 1, "world-writable"},
		{"chmod -R 755 /", 1, "recursive chmod"},
		{"git reset --hard", 1, "git reset hard"},
		{"git push --force origin main", 1, "git force push"},
		{"git push origin main -f", 1, "git force push short"},
		{"git clean -fd", 1, "git clean force"},
		{"kill -9 1234", 1, "kill -9"},
		{"> important.txt", 1, "file truncation"},
		{"  > file.txt", 1, "file truncation with leading space"},
		{"DROP TABLE users", 1, "SQL drop table"},
		{"drop database production", 1, "SQL drop database case insensitive"},
		{"TRUNCATE orders", 1, "SQL truncate"},
	}

	for _, tt := range dangerous {
		t.Run(tt.desc, func(t *testing.T) {
			warnings := checkDangerousCommand(tt.cmd)
			if len(warnings) < tt.expectMin {
				t.Errorf("checkDangerousCommand(%q) returned %d warnings, want >= %d",
					tt.cmd, len(warnings), tt.expectMin)
			}
		})
	}
}

func TestCheckDangerousCommandSafe(t *testing.T) {
	safe := []string{
		"ls -la",
		"git status",
		"git push origin main",
		"cat /etc/hosts",
		"echo hello > output.txt",
		"rm file.txt",
		"chmod 644 file.txt",
		"git reset --soft HEAD~1",
		"kill 1234",
		"find . -name '*.go'",
		"docker rm container_name",
		"grep -r TODO .",
		"tar -czf archive.tar.gz src/",
	}

	for _, cmd := range safe {
		t.Run(cmd, func(t *testing.T) {
			warnings := checkDangerousCommand(cmd)
			if len(warnings) > 0 {
				t.Errorf("checkDangerousCommand(%q) returned %d warnings, want 0; first: %s",
					cmd, len(warnings), warnings[0].message)
			}
		})
	}
}

func TestCheckDangerousCommandSeverity(t *testing.T) {
	// High severity
	warnings := checkDangerousCommand("rm -rf /")
	hasHigh := false
	for _, w := range warnings {
		if w.severity == "high" {
			hasHigh = true
		}
	}
	if !hasHigh {
		t.Error("rm -rf / should have at least one 'high' severity warning")
	}

	// Medium severity
	warnings = checkDangerousCommand("kill -9 123")
	if len(warnings) == 0 {
		t.Fatal("kill -9 should trigger a warning")
	}
	if warnings[0].severity != "medium" {
		t.Errorf("kill -9 should be 'medium' severity, got %q", warnings[0].severity)
	}
}
