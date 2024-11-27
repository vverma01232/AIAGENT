package models

type ModelConfig struct {
	Model       string    `json:"model"`
	Stream      bool      `json:"stream"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	TopK        float64   `json:"top_k,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
