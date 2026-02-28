# go-laude

A Go program that interacts with a language model (Claude via OpenRouter) and can:

- Read files
- Write files
- Execute Bash commands

This project is a learning tool and experimental AI assistant shell built in Go.

---

## Features

- Send prompts to the AI from the command line
- AI can call three tools:
  - **Read**: Read and return the contents of a file
  - **Write**: Write content to a file
  - **Bash**: Execute a shell command
- Loop interaction up to 10 rounds per session
- Fully logs AI responses and tool outputs

---

## Prerequisites

- Go installed (version 1.20+ recommended)
- An OpenRouter API key with access to Claude or compatible models

---

## Setup

1. Clone the repository:

```bash
git clone https://github.com/<your-username>/go-llm-shell.git
cd go-llm-shell
