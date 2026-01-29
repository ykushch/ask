package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

var lastCommand string

func runInteractive(model string, stats *Stats) {
	fmt.Println("ask — natural language shell (type !help for commands, Ctrl+D to exit)")
	fmt.Println()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          buildInteractivePrompt(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	for {
		rl.SetPrompt(buildInteractivePrompt())
		line, err := rl.Readline()
		if err != nil { // EOF or interrupt
			fmt.Println()
			break
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		// Special commands
		if input == "!help" {
			printHelp()
			continue
		}
		if strings.HasPrefix(input, "!model ") {
			newModel := strings.TrimSpace(input[7:])
			if newModel != "" {
				model = newModel
				fmt.Printf("model set to: %s\n", model)
			}
			continue
		}
		if input == "!model" {
			fmt.Printf("current model: %s\n", model)
			continue
		}
		if strings.HasPrefix(input, "!explain ") {
			cmd := strings.TrimSpace(input[9:])
			if cmd == "" {
				fmt.Println("Usage: !explain <command>")
				continue
			}
			stats.RecordExplain(model)
			spinner := NewSpinner("Explaining...")
			spinner.Start()
			explanation, err := explain(model, cmd)
			spinner.Stop()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\033[31merror: %v\033[0m\n", err)
				continue
			}
			fmt.Println(explanation)
			continue
		}
		if strings.HasPrefix(input, "!") {
			cmd := input[1:]
			if cmd == "" {
				continue
			}
			warnIfDangerous(cmd)
			stdout, stderr, _ := executeCommand(cmd)
			if stdout != "" {
				fmt.Print(stdout)
			}
			if stderr != "" {
				fmt.Fprint(os.Stderr, stderr)
			}
			addToHistory(cmd, stdout+stderr)
			lastCommand = cmd
			continue
		}

		// Explain commands
		if input == "?" {
			if lastCommand == "" {
				fmt.Println("No previous command to explain.")
				continue
			}
			stats.RecordExplain(model)
			spinner := NewSpinner("Explaining...")
			spinner.Start()
			explanation, err := explain(model, lastCommand)
			spinner.Stop()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\033[31merror: %v\033[0m\n", err)
				continue
			}
			fmt.Println(explanation)
			continue
		}
		if strings.HasPrefix(input, "?") {
			cmd := strings.TrimSpace(input[1:])
			if cmd == "" {
				continue
			}
			stats.RecordExplain(model)
			spinner := NewSpinner("Explaining...")
			spinner.Start()
			explanation, err := explain(model, cmd)
			spinner.Stop()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\033[31merror: %v\033[0m\n", err)
				continue
			}
			fmt.Println(explanation)
			continue
		}

		// cd handling
		if input == "cd" {
			home, _ := os.UserHomeDir()
			os.Chdir(home)
			continue
		}
		if strings.HasPrefix(input, "cd ") {
			path := expandHome(strings.TrimSpace(input[3:]))
			if err := os.Chdir(path); err != nil {
				fmt.Fprintf(os.Stderr, "cd: %v\n", err)
			}
			continue
		}

		// Direct shell command (not natural language)
		if !isNaturalLanguage(input) {
			warnIfDangerous(input)
			stdout, stderr, _ := executeCommand(input)
			if stdout != "" {
				fmt.Print(stdout)
			}
			if stderr != "" {
				fmt.Fprint(os.Stderr, stderr)
			}
			addToHistory(input, stdout+stderr)
			lastCommand = input
			continue
		}

		// Natural language → translate via Ollama
		spinner := NewSpinner("Thinking...")
		spinner.Start()
		command, err := translate(model, input)
		spinner.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\033[31merror: %v\033[0m\n", err)
			continue
		}
		stats.RecordInteractiveCommand(model, input, command)
		confirmAndRun(command, stats)
		lastCommand = command
	}
}

func buildInteractivePrompt() string {
	cwd, _ := os.Getwd()
	dir := filepath.Base(cwd)
	return fmt.Sprintf("\033[32m%s\033[0m > ", dir)
}

func printHelp() {
	fmt.Println("  !help        — show this help")
	fmt.Println("  !model NAME  — switch Ollama model")
	fmt.Println("  !model       — show current model")
	fmt.Println("  !explain CMD — explain a shell command")
	fmt.Println("  ?CMD         — explain a shell command (shorthand)")
	fmt.Println("  ?            — explain the last executed command")
	fmt.Println("  !cmd         — run cmd directly (bypass AI)")
	fmt.Println("  Ctrl+D       — exit")
	fmt.Println()
}
