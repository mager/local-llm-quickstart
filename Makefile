.PHONY: build run server gguf server-gguf install-llama clean

MODEL_REPO ?= ggml-org/gemma-4-E4B-it-GGUF
MODEL_FILE ?= gemma-4-E4B-it-Q4_K_M.gguf
MODEL_PATH ?= models/$(MODEL_FILE)
PORT ?= 8000
TRANSFORMERS_MODEL ?= ~/LLM/models/gemma-4-E4B-it

build:
	go build ./cmd/local-llm

run:
	go run ./cmd/local-llm --endpoint http://127.0.0.1:$(PORT) --model $(TRANSFORMERS_MODEL)

server:
	./scripts/start-transformers-server.sh

gguf:
	./scripts/download-gemma-e4b-gguf.sh

server-gguf:
	llama-server -m $(MODEL_PATH) --host 127.0.0.1 --port $(PORT) --ctx-size 8192

install-llama:
	brew install llama.cpp

clean:
	rm -f local-llm
