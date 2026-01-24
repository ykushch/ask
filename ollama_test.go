package main

import (
	"os"
	"testing"
)

func TestStripCodeFences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "multiline fences",
			input: "```bash\nls -la\n```",
			want:  "ls -la",
		},
		{
			name:  "multiline fences with multiple lines",
			input: "```sh\necho hello\necho world\n```",
			want:  "echo hello\necho world",
		},
		{
			name:  "single backticks",
			input: "`ls -la`",
			want:  "ls -la",
		},
		{
			name:  "no fences",
			input: "ls -la",
			want:  "ls -la",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "fences with language tag",
			input: "```python\nprint('hello')\n```",
			want:  "print('hello')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripCodeFences(tt.input)
			if got != tt.want {
				t.Errorf("stripCodeFences() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOllamaHost(t *testing.T) {
	// Test default value
	os.Unsetenv("OLLAMA_HOST")
	got := ollamaHost()
	if got != "http://localhost:11434" {
		t.Errorf("ollamaHost() default = %q, want %q", got, "http://localhost:11434")
	}

	// Test env override
	os.Setenv("OLLAMA_HOST", "http://custom:5000")
	defer os.Unsetenv("OLLAMA_HOST")
	got = ollamaHost()
	if got != "http://custom:5000" {
		t.Errorf("ollamaHost() with env = %q, want %q", got, "http://custom:5000")
	}

	// Test trailing slash removal
	os.Setenv("OLLAMA_HOST", "http://custom:5000/")
	got = ollamaHost()
	if got != "http://custom:5000" {
		t.Errorf("ollamaHost() trailing slash = %q, want %q", got, "http://custom:5000")
	}
}
