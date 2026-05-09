package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"investment-analysis/persistence/model"
)

// newTestSettings returns settings pointed at the given server URL.
func newTestSettings(serverURL string) model.LLMSettings {
	return model.LLMSettings{
		Provider:     "lm-studio",
		Endpoint:     serverURL + "/v1/chat/completions",
		Model:        "test-model",
		Temperature:  0.7,
		SystemPrompt: "You are a test assistant.",
	}
}

// serveChatResponse starts a test server that returns a fixed reply string.
func serveChatResponse(t *testing.T, reply string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{
				{Message: Message{Role: "assistant", Content: reply}},
			},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestInfer_Success verifies that Infer returns the model reply.
func TestInfer_Success(t *testing.T) {
	srv := serveChatResponse(t, "The answer is 42.")
	client := NewClient(newTestSettings(srv.URL))

	got, err := client.Infer(context.Background(), "What is the answer?")
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	print(got)
	if got != "The answer is 42." {
		t.Errorf("got %q; want %q", got, "The answer is 42.")
	}
}

// TestInfer_SystemPromptIncluded verifies that the system prompt is sent as
// the first message in the request body.
func TestInfer_SystemPromptIncluded(t *testing.T) {
	var captured chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{{Message: Message{Role: "assistant", Content: "ok"}}},
		})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(newTestSettings(srv.URL))
	if _, err := client.Infer(context.Background(), "hello"); err != nil {
		t.Fatalf("Infer: %v", err)
	}

	if len(captured.Messages) < 2 {
		t.Fatalf("expected at least 2 messages, got %d", len(captured.Messages))
	}
	if captured.Messages[0].Role != "system" {
		t.Errorf("first message role = %q; want %q", captured.Messages[0].Role, "system")
	}
	if captured.Messages[1].Role != "user" {
		t.Errorf("second message role = %q; want %q", captured.Messages[1].Role, "user")
	}
	if captured.Messages[1].Content != "hello" {
		t.Errorf("user message = %q; want %q", captured.Messages[1].Content, "hello")
	}
}

// TestInfer_NoSystemPrompt verifies that a blank system prompt is omitted.
func TestInfer_NoSystemPrompt(t *testing.T) {
	var captured chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{{Message: Message{Role: "assistant", Content: "ok"}}},
		})
	}))
	t.Cleanup(srv.Close)

	settings := newTestSettings(srv.URL)
	settings.SystemPrompt = ""
	client := NewClient(settings)

	if _, err := client.Infer(context.Background(), "hi"); err != nil {
		t.Fatalf("Infer: %v", err)
	}
	if len(captured.Messages) != 1 {
		t.Errorf("expected 1 message (no system prompt), got %d", len(captured.Messages))
	}
	if captured.Messages[0].Role != "user" {
		t.Errorf("message role = %q; want %q", captured.Messages[0].Role, "user")
	}
}

// TestChat_MultiTurn verifies that Chat passes through all provided messages.
func TestChat_MultiTurn(t *testing.T) {
	var captured chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{{Message: Message{Role: "assistant", Content: "got it"}}},
		})
	}))
	t.Cleanup(srv.Close)

	messages := []Message{
		{Role: "system", Content: "be brief"},
		{Role: "user", Content: "first"},
		{Role: "assistant", Content: "sure"},
		{Role: "user", Content: "second"},
	}
	client := NewClient(newTestSettings(srv.URL))
	if _, err := client.Chat(context.Background(), messages); err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if len(captured.Messages) != 4 {
		t.Errorf("sent %d messages; want 4", len(captured.Messages))
	}
}

// TestInfer_APIKeyForwarded verifies that a non-empty APIKey is sent as a
// Bearer token.
func TestInfer_APIKeyForwarded(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message Message `json:"message"`
			}{{Message: Message{Role: "assistant", Content: "ok"}}},
		})
	}))
	t.Cleanup(srv.Close)

	settings := newTestSettings(srv.URL)
	settings.APIKey = "sk-test-key"
	client := NewClient(settings)

	if _, err := client.Infer(context.Background(), "ping"); err != nil {
		t.Fatalf("Infer: %v", err)
	}
	if gotAuth != "Bearer sk-test-key" {
		t.Errorf("Authorization = %q; want %q", gotAuth, "Bearer sk-test-key")
	}
}

// TestInfer_ServerError verifies that a non-2xx HTTP response is returned as
// an error containing the status code.
func TestInfer_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model not loaded", http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	client := NewClient(newTestSettings(srv.URL))
	_, err := client.Infer(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for 503 response, got nil")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("error %q should mention 503", err.Error())
	}
}

// TestInfer_EmptyChoices verifies that a response with no choices is an error.
func TestInfer_EmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResponse{Choices: nil})
	}))
	t.Cleanup(srv.Close)

	client := NewClient(newTestSettings(srv.URL))
	_, err := client.Infer(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for empty choices, got nil")
	}
}
