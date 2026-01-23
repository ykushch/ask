# ask

A CLI tool that translates natural language into shell commands using local AI models via [Ollama](https://ollama.com).

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/ykushch/ask/main/install.sh | bash
```

This will:
- Install Ollama if not already present
- Pull the default model (`qwen2.5-coder:7b`)
- Install the `ask` binary to `~/.local/bin`

## Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/ykushch/ask/main/uninstall.sh | bash
```

## Usage

### One-shot mode

```bash
ask find all markdown files in this directory
# → find . -name "*.md" [Enter to run]

ask show disk usage sorted by size
# → du -sh * | sort -h [Enter to run]

ask kill the process on port 3000
# → lsof -t -i:3000 | xargs kill [Enter to run]
```

### Interactive mode

```bash
ask
```

Drops into a REPL where you can type queries continuously:

```
projects > list go files
→ find . -name "*.go" [Enter to run]
projects > compress the src folder
→ tar -czf src.tar.gz src [Enter to run]
```

Interactive commands:
- `!help` — show available commands
- `!model NAME` — switch model
- `!model` — show current model
- `!cmd` — run `cmd` directly (bypass AI)
- `Ctrl+D` — exit

### Specify a different model

```bash
# Via flag
ask --model llama3 show my public ip

# Via environment variable
export ASK_MODEL=deepseek-r1
ask list running docker containers
```

### Check version

```bash
ask -v
# ask version 0.1.0
# model: qwen2.5-coder:7b
# ollama: http://localhost:11434
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `ASK_MODEL` | Ollama model to use | `qwen2.5-coder:7b` |
| `OLLAMA_HOST` | Ollama server URL | `http://localhost:11434` |

## Requirements

- macOS or Linux
- [Ollama](https://ollama.com) (installed automatically by the install script)

## Build from source

```bash
git clone https://github.com/ykushch/ask.git
cd ask
go build -o ask .
```
