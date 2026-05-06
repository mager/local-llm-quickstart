# Local LLM Quickstart

A small Go TUI for talking to a local OpenAI-compatible model server.

The app does not load model weights itself. It speaks to a local OpenAI-compatible
endpoint, which keeps the interface simple and lets the serving runtime do the hard
inference work.

## Shape

```text
Go Bubble Tea TUI
        |
        v
http://127.0.0.1:8080/v1/chat/completions
        |
        v
llama.cpp, Transformers Serve, vLLM, or another compatible runtime
```

## Fast Path: GGUF + llama.cpp

This is the recommended local interactive path.

```bash
brew install llama.cpp
make gguf
make server-gguf PORT=8080
```

In another terminal:

```bash
LOCAL_LLM_ENDPOINT=http://127.0.0.1:8080 LOCAL_LLM_MODEL=local make run PORT=8080
```

That downloads `ggml-org/gemma-4-E4B-it-GGUF` and uses
`gemma-4-E4B-it-Q4_K_M.gguf`, about 5.34 GB.

## Full-Weight Transformers Path

Use this if you specifically want to try the full Hugging Face weights.

### 1. Finish The Gemma Download

This app is set up to use the exact model you already started downloading:

```text
~/LLM/models/gemma-4-E4B-it
```

Your download command:

```bash
cd ~/LLM
hf download google/gemma-4-E4B-it \
  --local-dir models/gemma-4-E4B-it
```

### 2. Install The Serving Runtime

In the same Python environment you used for the download:

```bash
cd ~/LLM
source .venv/bin/activate
pip install -U "transformers[serving]" torch accelerate torchvision pillow
```

Hugging Face's `transformers serve` command starts an OpenAI-compatible local server.

### 3. Start The Local Server

From this repo:

```bash
cd ~/Code/local-llm-quickstart
make server
```

This serves:

```text
http://127.0.0.1:8000/v1/chat/completions
```

### 4. Run The TUI

In another terminal:

```bash
make run
```

Or directly:

```bash
go run ./cmd/local-llm \
  --endpoint http://127.0.0.1:8000 \
  --model ~/LLM/models/gemma-4-E4B-it
```

## TUI Commands

```text
/help                show commands
/continue            continue after a token-limit cutoff
/model               show the endpoint and model name
/reset               clear the conversation
/save transcript.md  save the current chat
/tokens auto         estimate max tokens from the prompt
/tokens 4096         manually set max tokens
/temp 0.4            set temperature
/quit                quit
```

Scroll long responses with `PageUp`/`PageDown`, `Ctrl+U`/`Ctrl+D`,
`Ctrl+G` for top, `Ctrl+B` for bottom, or your mouse/trackpad wheel.
Assistant responses are rendered as terminal Markdown with Charmbracelet Glamour,
so headings, lists, emphasis, and code blocks are easier to scan.

The default token mode is `auto`. Tiny prompts get small budgets, and bigger
requests like itineraries, drafts, or code tasks get more room. If a server stops
because it hit the token cap, the TUI tells you to use `/continue` or raise the
budget.

## Offline Check

Before a flight, run everything once while online:

```bash
make server-gguf PORT=8080
LOCAL_LLM_ENDPOINT=http://127.0.0.1:8080 LOCAL_LLM_MODEL=local make run PORT=8080
```

Then turn Wi-Fi off and run the same commands. If the server starts and the TUI answers,
token costs are zero. The costs are battery, heat, RAM, and patience.

## Why This Way

This project keeps a clean boundary between the hard part and the fun part.

The serving runtime handles model loading and inference. The Go app owns the interface:
prompts, chat history, transcripts, keyboard flow, and whatever workflow features we
want to add next.
