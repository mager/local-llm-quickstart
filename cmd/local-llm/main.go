package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mager/local-llm-quickstart/internal/llm"
	"github.com/mager/local-llm-quickstart/internal/tui"
)

func main() {
	endpoint := flag.String("endpoint", envOrDefault("LOCAL_LLM_ENDPOINT", "http://127.0.0.1:8000"), "OpenAI-compatible local server URL")
	model := flag.String("model", envOrDefault("LOCAL_LLM_MODEL", "~/LLM/models/gemma-4-E4B-it"), "model name or local path sent to the local server")
	maxTokens := flag.Int("tokens", 0, "max tokens to generate; 0 means auto")
	temperature := flag.Float64("temp", 0.7, "sampling temperature")
	flag.Parse()

	client := llm.NewClient(*endpoint, *model)
	app := tui.New(tui.Config{
		Client:      client,
		Endpoint:    *endpoint,
		Model:       *model,
		MaxTokens:   *maxTokens,
		Temperature: *temperature,
	})

	if _, err := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "local-llm: %v\n", err)
		os.Exit(1)
	}
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
