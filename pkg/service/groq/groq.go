package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	groqEndpoint = "https://api.groq.com/openai/v1/chat/completions"
	defaultModel = "llama-3.3-70b-versatile"
	httpTimeout  = 15 * time.Second
)

// Message represents a single chat message in the OpenAI-compatible format.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request is the payload sent to the Groq chat completions API.
type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Response is the parsed response from the Groq API.
type Response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Client wraps the Groq API communication.
type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

// NewClient creates a new Groq client. Reads GROQ_API_KEY from environment.
// Returns nil if the key is not configured.
func NewClient() *Client {
	key := os.Getenv("GROQ_API_KEY")
	if key == "" {
		return nil
	}
	return &Client{
		apiKey:     key,
		httpClient: &http.Client{Timeout: httpTimeout},
		model:      defaultModel,
	}
}

// Complete sends a chat completion request and returns the text content.
func (c *Client) Complete(ctx context.Context, systemPrompt, userPrompt string, temperature float64) (string, error) {
	req := Request{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: temperature,
		MaxTokens:   300,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, groqEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("groq api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq api returned %d: %s", resp.StatusCode, string(respBody))
	}

	var groqResp Response
	if err := json.Unmarshal(respBody, &groqResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("groq api returned no choices")
	}

	return groqResp.Choices[0].Message.Content, nil
}
