// Package llm provides a client for making inference calls to LM Studio and
// other OpenAI-compatible local LLM servers.
//
// Basic usage:
//
//	settings, _ := store.GetSettings("lm-studio")
//	client := llm.NewClient(settings)
//	resp, err := client.Infer(ctx, "Summarise this document: ...")
//	fmt.Println(resp.Content, resp.TotalTokens, resp.Elapsed)
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"investment-analysis/util"
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

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage *usage `json:"usage,omitempty"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// --- public response type ---------------------------------------------------

// Response is the result of a Chat or Infer call: the model's reply plus
// metadata about the call.  Token counts are zero if the server did not
// return a usage block.
type Response struct {
	Content          string        `json:"content"`
	PromptTokens     int           `json:"prompt_tokens"`
	CompletionTokens int           `json:"completion_tokens"`
	TotalTokens      int           `json:"total_tokens"`
	RequestedAt      time.Time     `json:"requested_at"`
	RespondedAt      time.Time     `json:"responded_at"`
	Elapsed          time.Duration `json:"elapsed"`
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
// prompt) and returns the model's reply plus call metadata.  It is a
// convenience wrapper around Chat.
func (c *Client) Infer(ctx context.Context, prompt string) (_ *Response, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "llm.Client.Infer",
			"provider", c.settings.Provider, "model", c.settings.Model, "promptLen", len(prompt))
	}()
	messages := []Message{}
	if sp := strings.TrimSpace(c.settings.SystemPrompt); sp != "" {
		messages = append(messages, Message{Role: "system", Content: sp})
	}
	messages = append(messages, Message{Role: "user", Content: prompt})
	return c.Chat(ctx, messages)
}

// Chat sends an arbitrary sequence of messages and returns the model's reply
// plus call metadata.  Use this for multi-turn conversations where you
// manage the message history yourself.
func (c *Client) Chat(ctx context.Context, messages []Message) (_ *Response, err error) {
	defer func() {
		util.LogIfErr(ctx, &err, "llm.Client.Chat",
			"provider", c.settings.Provider, "model", c.settings.Model, "messageCount", len(messages))
	}()
	reqBody := chatRequest{
		Model:       c.settings.Model,
		Messages:    messages,
		Temperature: c.settings.Temperature,
		Stream:      false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("llm: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.settings.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("llm: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.settings.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.settings.APIKey)
	}

	requestedAt := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm: POST %s: %w", c.settings.Endpoint, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	respondedAt := time.Now()
	if err != nil {
		return nil, fmt.Errorf("llm: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("llm: HTTP %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}

	var result chatResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("llm: decode response: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("llm: server error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("llm: empty choices in response")
	}

	out := &Response{
		Content:     strings.TrimSpace(result.Choices[0].Message.Content),
		RequestedAt: requestedAt,
		RespondedAt: respondedAt,
		Elapsed:     respondedAt.Sub(requestedAt),
	}
	if result.Usage != nil {
		out.PromptTokens = result.Usage.PromptTokens
		out.CompletionTokens = result.Usage.CompletionTokens
		out.TotalTokens = result.Usage.TotalTokens
	}
	return out, nil
}
