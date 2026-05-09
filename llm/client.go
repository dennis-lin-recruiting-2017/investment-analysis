// Package llm provides a client for making inference calls to LM Studio and
// other OpenAI-compatible local LLM servers.
//
// Basic usage:
//
//	settings, _ := store.GetSettings("lm-studio")
//	client := llm.NewClient(settings)
//	reply, err := client.Infer(ctx, "Summarise this document: ...")
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"investment-analysis/persistence/model"
)

// --- wire types (OpenAI chat-completions schema) ----------------------------

// Message is a single turn in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// --- Client -----------------------------------------------------------------

// Client sends inference requests to an LM Studio (or any OpenAI-compatible)
// server using the settings stored in the persistence layer.
type Client struct {
	settings   model.LLMSettings
	httpClient *http.Client
}

// NewClient returns a Client configured from settings.
// The HTTP timeout is fixed at 5 minutes; LLM inference on large documents
// can be slow.
func NewClient(settings model.LLMSettings) *Client {
	return &Client{
		settings:   settings,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

// Infer sends prompt as a user message (preceded by the configured system
// prompt) and returns the model's reply text.  It is a convenience wrapper
// around Chat.
func (c *Client) Infer(ctx context.Context, prompt string) (string, error) {
	messages := []Message{}
	if sp := strings.TrimSpace(c.settings.SystemPrompt); sp != "" {
		messages = append(messages, Message{Role: "system", Content: sp})
	}
	messages = append(messages, Message{Role: "user", Content: prompt})
	return c.Chat(ctx, messages)
}

// Chat sends an arbitrary sequence of messages and returns the model's reply
// text.  Use this for multi-turn conversations where you manage the message
// history yourself.
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	reqBody := chatRequest{
		Model:       c.settings.Model,
		Messages:    messages,
		Temperature: c.settings.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("llm: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.settings.Endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("llm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.settings.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.settings.APIKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm: POST %s: %w", c.settings.Endpoint, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("llm: HTTP %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}

	var result chatResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("llm: decode response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("llm: server error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("llm: empty choices in response")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}
