package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func ollamaHost() string {
	host := getEnvDefault("OLLAMA_HOST", "http://localhost:11434")
	host = strings.TrimRight(host, "/")
	return host
}

func checkOllama() error {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(ollamaHost() + "/")
	if err != nil {
		return fmt.Errorf("cannot connect to Ollama at %s — is it running?", ollamaHost())
	}
	resp.Body.Close()
	return nil
}

func generate(model, prompt string) (string, error) {
	reqBody := ollamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Post(ollamaHost()+"/api/generate", "application/json", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("request to Ollama failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama error (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("Ollama: %s", ollamaResp.Error)
	}

	result := strings.TrimSpace(ollamaResp.Response)
	// Strip markdown code fences if the model wraps the output
	result = stripCodeFences(result)
	return result, nil
}

func stripCodeFences(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) >= 2 && strings.HasPrefix(lines[0], "```") && strings.HasSuffix(lines[len(lines)-1], "```") {
		inner := lines[1 : len(lines)-1]
		return strings.TrimSpace(strings.Join(inner, "\n"))
	}
	// Single-line backtick wrap
	s = strings.TrimPrefix(s, "`")
	s = strings.TrimSuffix(s, "`")
	return s
}
