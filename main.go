package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// version is set at build time via -ldflags "-X main.version=..."
// Falls back to "dev" for local builds.
var version = "dev"

func getEnvDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func translate(model, input string) (string, error) {
	prompt := buildPrompt(input)
	return generate(model, prompt)
}

func main() {
	defaultModel := getEnvDefault("ASK_MODEL", "qwen2.5-coder:7b")
	model := flag.String("model", defaultModel, "Ollama model to use")
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	var doUpdate bool
	flag.BoolVar(&doUpdate, "update", false, "Update ask to the latest version")
	var doExplain bool
	flag.BoolVar(&doExplain, "explain", false, "Explain a shell command instead of generating one")
	flag.Parse()

	if doUpdate {
		if err := selfUpdate(); err != nil {
			fmt.Fprintf(os.Stderr, "update failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if showVersion {
		fmt.Printf("ask version %s\n", version)
		fmt.Printf("model: %s\n", *model)
		fmt.Printf("ollama: %s\n", ollamaHost())
		return
	}

	// Start background version check
	updateCh := make(chan string, 1)
	go backgroundVersionCheck(updateCh)

	if err := checkOllama(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		runInteractive(*model)
		printUpdateNotice(updateCh)
		return
	}

	// One-shot mode: join all args as the query
	query := strings.Join(args, " ")

	if doExplain {
		spinner := NewSpinner("Explaining...")
		spinner.Start()
		explanation, err := explain(*model, query)
		spinner.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(explanation)
		printUpdateNotice(updateCh)
		return
	}

	spinner := NewSpinner("Thinking...")
	spinner.Start()
	command, err := translate(*model, query)
	spinner.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	confirmAndRun(command)
	printUpdateNotice(updateCh)
}

func printUpdateNotice(ch <-chan string) {
	select {
	case tag := <-ch:
		if tag != "" {
			fmt.Fprintf(os.Stderr, "\nA new version of ask is available (%s). Run \"ask --update\" to upgrade.\n", tag)
		}
	default:
		// Check didn't finish in time, skip silently
	}
}
