package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const endpoint = "https://4sapi.com/v1/chat/completions"

func main() {
	// Read the API key from the environment instead of storing it in source code.
	key := os.Getenv("FOUR_S_API_KEY")
	if key == "" {
		panic("FOUR_S_API_KEY is required")
	}

	// Allow callers to select a model while retaining a sensible default for testing.
	model := os.Getenv("FOUR_S_API_MODEL")
	if model == "" {
		model = "gpt-5.6"
	}

	// 4SAPI accepts the OpenAI-compatible Chat Completions request format.
	body, err := json.Marshal(map[string]any{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "Introduce 4SAPI in one sentence.",
			},
		},
	})
	if err != nil {
		panic(fmt.Errorf("encode request body: %w", err))
	}

	// Send the request to the 4SAPI OpenAI-compatible endpoint.
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		panic(fmt.Errorf("create request: %w", err))
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Errorf("send request: %w", err))
	}
	defer res.Body.Close()

	// Read the response before checking the status so API error details are available.
	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(fmt.Errorf("read response body: %w", err))
	}
	if res.StatusCode >= 300 {
		panic(fmt.Sprintf("HTTP %s: %s", res.Status, data))
	}

	// Decode only the response fields needed to print the assistant's reply.
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		panic(fmt.Errorf("decode response body: %w", err))
	}

	if len(parsed.Choices) > 0 {
		fmt.Println(parsed.Choices[0].Message.Content)
	}
}
