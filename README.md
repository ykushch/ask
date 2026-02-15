```
           _
  __ _ ___| | __
 / _` / __| |/ /
| (_| \__ \   <
 \__,_|___/_|\_\
```

[![CI](https://github.com/ykushch/ask/actions/workflows/ci.yml/badge.svg)](https://github.com/ykushch/ask/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ykushch/ask)](https://github.com/ykushch/ask/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/ykushch/ask/total)](https://github.com/ykushch/ask/releases)
[![Go](https://img.shields.io/badge/Go-1.21-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey)](https://github.com/ykushch/ask)
[![Ollama](https://img.shields.io/badge/Ollama-local%20AI-1a1a2e)](https://ollama.com)
[![Runs Locally](https://img.shields.io/badge/runs%20100%25-locally-green)](#)

<img src="assets/ask-demo.gif" alt="ask demo" width="720" />

**Natural language to shell commands, powered by local AI.**

A CLI tool that translates what you mean into what to type, using [Ollama](https://ollama.com) models running entirely on your machine.

## Features

- **Runs 100% locally** — No API keys, no cloud, no data leaves your machine
- **Project-aware** — Detects Go, Node, Python, Rust, and tailors commands accordingly
- **Explain mode** — Don't know what a command does? Ask `?tar -czf`
- **Safety warnings** — Flags dangerous commands like `rm -rf` before execution
- **Interactive REPL** — Conversational shell with command history context
- **Usage statistics** — Track your usage with `--stats`

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
- `!explain CMD` — explain a shell command
- `?CMD` — explain a shell command (shorthand)
- `?` — explain the last executed command
- `!cmd` — run `cmd` directly (bypass AI)
- `Ctrl+D` — exit

### Explain mode

Don't know what a command does? Ask for an explanation:

```bash
ask --explain "find . -name '*.go' -exec grep 'func main' {} +"
# Finds all .go files in the current directory tree and searches each one for lines containing "func main".
#   -name '*.go': match files ending in .go
#   -exec grep 'func main' {} +: run grep on the matched files
```

Also works in interactive mode with the `?` prefix:

```
projects > ?tar -czf src.tar.gz src
# Compresses the src directory into a gzipped tar archive named src.tar.gz.
#   -c: create a new archive
#   -z: compress with gzip
#   -f src.tar.gz: name the output file
```

### Project-aware suggestions

Commands are tailored to your project type. `ask` detects signature files in the current directory:

| File | Detected As | Example |
|------|-------------|---------|
| `go.mod` | Go | `ask run tests` → `go test ./...` |
| `package.json` | Node.js | `ask run tests` → `npm test` |
| `Cargo.toml` | Rust | `ask build` → `cargo build` |
| `requirements.txt` | Python | `ask run app` → `python app.py` |
| `Makefile` | Make-based | `ask build` → `make` |
| `Dockerfile` | Docker | Docker-aware suggestions |

### Safety warnings

Dangerous commands are flagged with a warning before execution:

```
  ⚠ Warning: Recursive deletion targeting a broad path
→ rm -rf ~/Documents [Enter to run]
```

Detected patterns include `rm -rf`, `dd`, `mkfs`, `chmod 777`, `git push --force`, `DROP TABLE`, and more. Warnings are informational — you can still press Enter to proceed.

### Specify a different model

```bash
# Via flag
ask --model llama3 show my public ip

# Via environment variable
export ASK_MODEL=deepseek-r1
ask list running docker containers
```

### Update

```bash
ask --update
```

Self-updates the binary to the latest GitHub release. A background version check also runs on every invocation — if a newer version is available, you'll see a notice after the command completes.

### Check version

```bash
ask -v
# ask version 0.1.0
# model: qwen2.5-coder:7b
# ollama: http://localhost:11434
```

### Usage statistics

Track how you use `ask` over time:

```bash
ask --stats
# ask usage statistics
# ────────────────────
# Total invocations:     150
# Commands generated:    120
# Commands executed:      95  (79%)
# Explain calls:          15
# Interactive sessions:   20
# One-shot commands:     100
#
# Model usage:
#   qwen2.5-coder:7b    140  (93%)
#   llama3               10  (7%)
#
# Stats file: ~/.ask/stats.json (12KB)
# Tracking since: 2026-01-29
```

Statistics are stored locally in `~/.ask/stats.json`.

## Recommended models
 
The default model is `qwen2.5-coder:7b` — a good balance of speed and accuracy for shell command generation. Depending on your hardware and needs, you may want to try other models:
 
| Model | Size | Best For | Pull Command |
|-------|------|----------|--------------|
| `qwen2.5-coder:7b` | 4.7 GB | General use (default) | `ollama pull qwen2.5-coder:7b` |
| `deepseek-coder:6.7b` | 3.8 GB | Code-focused, lighter | `ollama pull deepseek-coder:6.7b` |
| `nemotron-mini` | 2.7 GB | Lightweight, low-resource machines | `ollama pull nemotron-mini` |
| `nemotron-3-nano` | 24 GB | Reasoning-heavy queries, 1M context | `ollama pull nemotron-3-nano` |
| `deepseek-r1` | 4.7 GB | Reasoning tasks | `ollama pull deepseek-r1` |
 
**Tips:**
- For most users, the default `qwen2.5-coder:7b` works well
- On machines with limited RAM (<8 GB), try `nemotron-mini` (2.7 GB) or `deepseek-coder:6.7b` (3.8 GB)
- For complex commands requiring step-by-step reasoning, `nemotron-3-nano` or `deepseek-r1` may produce better results (requires 32 GB+ RAM)
 
Switch models with `--model` or set `ASK_MODEL`:
 
```bash
ask --model nemotron-mini show disk usage
# or
export ASK_MODEL=nemotron-mini
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

## Development

```bash
# Set up git hooks (runs tests on commit)
make setup

# Run tests
make test
```
