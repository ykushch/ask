package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var knownCommands = map[string]bool{
	"ls": true, "pwd": true, "clear": true, "exit": true, "quit": true,
	"whoami": true, "date": true, "cal": true, "top": true, "htop": true,
	"history": true, "which": true, "man": true, "touch": true, "head": true,
	"tail": true, "grep": true, "find": true, "sort": true, "wc": true,
	"diff": true, "tar": true, "zip": true, "unzip": true,
}

var shellPrefixes = []string{
	"cd ", "ls ", "echo ", "cat ", "mkdir ", "rm ", "cp ", "mv ",
	"git ", "npm ", "node ", "npx ", "python", "pip ", "brew ", "curl ",
	"wget ", "chmod ", "chown ", "sudo ", "vi ", "vim ", "nano ", "code ",
	"open ", "export ", "source ", "docker ", "kubectl ", "aws ", "gcloud ",
	"./", "/", "~", "$", ">", ">>", "|", "&&",
}

func isNaturalLanguage(input string) bool {
	if strings.HasPrefix(input, "!") {
		return false
	}
	if knownCommands[input] {
		return false
	}
	for _, prefix := range shellPrefixes {
		if strings.HasPrefix(input, prefix) {
			return false
		}
	}
	return true
}

func executeCommand(cmd string) (string, string, error) {
	c := exec.Command("sh", "-c", cmd)
	c.Dir, _ = os.Getwd()

	var stdout, stderr strings.Builder
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	return stdout.String(), stderr.String(), err
}

func confirmAndRun(cmd string) {
	fmt.Printf("\033[33m→ %s\033[0m [Enter to run] ", cmd)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	input := scanner.Text()

	if input != "" {
		return
	}

	if strings.HasPrefix(cmd, "cd ") {
		path := strings.TrimSpace(cmd[3:])
		path = expandHome(path)
		if err := os.Chdir(path); err != nil {
			fmt.Fprintf(os.Stderr, "cd: %v\n", err)
		}
		return
	}

	stdout, stderr, _ := executeCommand(cmd)
	if stdout != "" {
		fmt.Print(stdout)
	}
	if stderr != "" {
		fmt.Fprint(os.Stderr, stderr)
	}
	addToHistory(cmd, stdout+stderr)
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		return home + path[1:]
	}
	return path
}
