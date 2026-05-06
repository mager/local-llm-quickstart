#!/usr/bin/env bash
set -euo pipefail

MODEL_DIR="${MODEL_DIR:-$HOME/LLM/models/gemma-4-E4B-it}"
PORT="${PORT:-8000}"

if [[ ! -d "$MODEL_DIR" ]]; then
  echo "Model folder not found: $MODEL_DIR" >&2
  echo "Let your hf download finish first." >&2
  exit 1
fi

if ! find "$MODEL_DIR" -maxdepth 1 \( -name "*.safetensors" -o -name "*.bin" \) | grep -q .; then
  echo "Model folder exists, but weights are not present yet: $MODEL_DIR" >&2
  echo "Let your hf download finish first." >&2
  exit 1
fi

TRANSFORMERS_BIN="${TRANSFORMERS_BIN:-transformers}"
if [[ -x "$HOME/LLM/.venv/bin/transformers" ]]; then
  TRANSFORMERS_BIN="$HOME/LLM/.venv/bin/transformers"
fi

if ! command -v "$TRANSFORMERS_BIN" >/dev/null 2>&1; then
  echo "transformers CLI not found." >&2
  echo "Install it in your active Python env with:" >&2
  echo '  pip install -U "transformers[serving]" torch accelerate torchvision pillow' >&2
  exit 1
fi

echo "Starting Transformers server on http://127.0.0.1:$PORT"
echo "Use model: $MODEL_DIR"
"$TRANSFORMERS_BIN" serve --host 127.0.0.1 --port "$PORT"
