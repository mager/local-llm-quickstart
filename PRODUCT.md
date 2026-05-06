# PRODUCT.md

## Product

`local-llm-quickstart` is a small Go terminal app for talking to a local OpenAI-compatible model server.

It is not a model runtime. It is not a framework. It is not trying to hide how local inference works.

It is the smallest useful client boundary between a human and whatever local model server happens to be running today.

## Register

Product.

The design serves repeated use in a terminal. It should feel calm, legible, responsive, and slightly delightful, but it should never become a landing page, demo toy, or decorative terminal art project.

## Audience

The primary user is a developer who wants to try local LLMs without getting trapped in model-runtime ceremony.

They are probably comfortable with:

- Terminal apps
- Homebrew
- Go
- Python virtual environments
- Hugging Face downloads
- Copying commands between two terminal panes

They may not yet understand:

- Why model weights and model servers are separate
- Why GGUF feels faster than full safetensors locally
- Why an OpenAI-compatible local API is useful
- Why token limits affect answer quality
- What `llama-server` logs mean

The product should teach those things through use, not through lectures.

## Job To Be Done

When I have a local model server running, I want a tiny, pleasant chat interface that speaks the OpenAI API shape, so I can test prompts, inspect behavior, save transcripts, and swap runtimes without rewriting client code.

## Core Promise

Keep the app boundary boring.

The model server does the hard inference work. The TUI owns the human experience:

- Message entry
- Conversation display
- Markdown rendering
- Token budget ergonomics
- Transcript saving
- Local endpoint switching
- Commands that make recovery easy

## Non-Goals

This project should not become:

- A model manager
- A package manager for weights
- A full desktop chat app
- A benchmark suite
- A prompt engineering framework
- An abstraction layer over every LLM provider
- A replacement for llama.cpp, Ollama, Transformers Serve, or vLLM

It can document those tools. It should not absorb their jobs.

## Product Principles

### Local-first, not local-mystical

The app should make local inference feel approachable without pretending it is magic. Users should see the endpoint, model name, token budget, and server boundary.

### Runtime replaceable

The app should speak to any OpenAI-compatible local server. llama.cpp is the fast path today. Transformers Serve and vLLM remain valid alternatives.

### Fast path first

The recommended experience is GGUF plus `llama-server` because it feels interactive on a laptop. Full safetensors can be documented as an advanced path.

### Small tool, real polish

The app should stay small, but not crude. The difference between "toy" and "tool" is often scroll behavior, contrast, continuation, wrapping, and readable output.

### Honest failure

If the server is not running, say that. If the model hits a token limit, say that. If the answer is cut off, offer `/continue`.

### No fake intelligence in the shell

Auto token budgeting can be helpful, but it should remain understandable. Guess, show the guess, and let the user override it.

## Current Feature Set

- Go Bubble Tea TUI
- OpenAI-compatible `/v1/chat/completions` client
- GGUF fast path through `llama-server`
- Transformers Serve fallback path
- Auto token budgeting
- Manual token override
- `/continue` recovery after cutoffs
- Markdown rendering through Glamour
- Viewport scrolling
- Transcript saving
- Endpoint/model status in the header

## Experience Goals

The first good session should feel like this:

1. Start `llama-server`.
2. Run the TUI.
3. Ask a simple prompt.
4. Get a fast Markdown-rendered answer.
5. Ask a longer prompt.
6. Scroll comfortably.
7. If the answer cuts off, type `/continue`.
8. Save the transcript if useful.

## Copy Voice

The voice should be plain, technical, and slightly warm.

Good:

- "Stopped at token limit. Type `/continue`, or use `/tokens 4096` and ask again."
- "Start typing below. The model stays local; this TUI only talks to your localhost server."
- "endpoint=http://127.0.0.1:8080"

Avoid:

- Marketing claims
- Anthropomorphic model language
- Cute error messages that hide the actual problem
- Long help text in the main UI

## Key Concepts To Teach

### The TUI is not the runtime

The TUI sends HTTP requests. `llama-server`, Transformers Serve, or vLLM loads weights and generates tokens.

### OpenAI-compatible is the boundary

The value is not OpenAI specifically. The value is a common API shape that lets local and cloud runtimes be swapped.

### Quantized does not mean toy

The 5.34 GB GGUF path can feel dramatically better than full local safetensors for interactive chat.

### Token budget is UX

Too few tokens makes the model look broken. Too many can waste time. The app should help pick a reasonable default and explain cutoffs.

## Success Criteria

The project is succeeding if:

- A developer can clone it and get a local chat running in under 15 minutes after model download.
- The README explains the fast path without burying it under alternatives.
- Long Markdown answers are readable in a terminal.
- Cutoff answers are recoverable.
- The code remains small enough to read in one sitting.
- The repo teaches the runtime boundary clearly enough that users can adapt it to their own project.

## Sharp Edges To Keep Visible

- Model downloads are large.
- First load can be slow.
- Runtime setup differs by machine.
- Generated answers are not guaranteed to be correct.
- Local inference still costs battery, heat, RAM, and patience.

## Future Ideas

- Toggle Markdown style themes.
- Add `/copy` if clipboard support behaves consistently.
- Show elapsed time and generated token count when providers expose usage.
- Add streaming responses.
- Add a simple config file for default endpoint/model.
- Add transcript browser commands.
- Add a compact "server health" check before first prompt.

## Anti-Features

Do not add accounts.

Do not add telemetry.

Do not add hosted defaults.

Do not hide localhost behind vague labels.

Do not make the app require a specific model.

Do not make the TUI dependent on `llama-server` specifically.

