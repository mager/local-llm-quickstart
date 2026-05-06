package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Content      string
	FinishReason string
}

type Client struct {
	endpoint   string
	model      string
	httpClient *http.Client
}

func NewClient(endpoint string, model string) Client {
	return Client{
		endpoint: strings.TrimRight(endpoint, "/"),
		model:    expandHome(model),
		httpClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func expandHome(path string) string {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home + strings.TrimPrefix(path, "~")
	}
	return path
}

func (c Client) Chat(ctx context.Context, messages []Message, maxTokens int, temperature float64) (Response, error) {
	requestBody := chatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      false,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint+"/v1/chat/completions",
		bytes.NewReader(payload),
	)
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("request local model: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("local model returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var decoded chatCompletionResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return Response{}, fmt.Errorf("decode local model response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return Response{}, fmt.Errorf("local model returned no choices")
	}

	return Response{
		Content:      strings.TrimSpace(decoded.Choices[0].Message.Content),
		FinishReason: decoded.Choices[0].FinishReason,
	}, nil
}

type chatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}
