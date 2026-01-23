package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const version = "0.1.0"

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
	flag.Parse()

	if showVersion {
		fmt.Printf("ask version %s\n", version)
		fmt.Printf("model: %s\n", *model)
		fmt.Printf("ollama: %s\n", ollamaHost())
		return
	}

	if err := checkOllama(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		runInteractive(*model)
		return
	}

	// One-shot mode: join all args as the query
	query := strings.Join(args, " ")
	command, err := translate(*model, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	confirmAndRun(command)
}
