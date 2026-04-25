package model

type LLMSettings struct {
	Provider     string  `json:"provider"`
	Endpoint     string  `json:"endpoint"`
	Model        string  `json:"model"`
	APIKey       string  `json:"apiKey,omitempty"`
	Temperature  float64 `json:"temperature"`
	SystemPrompt string  `json:"systemPrompt"`
}

func DefaultSettings(provider string) LLMSettings {
	switch provider {
	case "ollama":
		return LLMSettings{
			Provider:     "ollama",
			Endpoint:     "http://localhost:11434/api/generate",
			Model:        "llama3.2",
			Temperature:  0.7,
			SystemPrompt: "You are a helpful assistant.",
		}
	case "external-endpoint":
		return LLMSettings{
			Provider:     "external-endpoint",
			Endpoint:     "https://api.example.com/v1/chat/completions",
			Model:        "gpt-4.1-mini",
			Temperature:  0.7,
			SystemPrompt: "You are a helpful assistant.",
		}
	default:
		return LLMSettings{
			Provider:     "lm-studio",
			Endpoint:     "http://localhost:1234/v1/chat/completions",
			Model:        "local-model",
			Temperature:  0.7,
			SystemPrompt: "You are a helpful assistant.",
		}
	}
}

func Providers() []string {
	return []string{"lm-studio", "ollama", "external-endpoint"}
}
