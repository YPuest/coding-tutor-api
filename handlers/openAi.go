package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func GetAIResponse(prompt string) (string, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4Turbo,
			Messages: []openai.ChatCompletionMessage{
				{Role: "developer", Content: "You are a helpful coding tutor."},
				{Role: "user", Content: prompt},
			},
			MaxTokens:   1000,
			Temperature: 0.2,
		},
	)

	if err != nil {
		log.Printf("OpenAI error: %v\n", err)
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func CleanAndExtractJSON(aiResponse string) (string, error) {
	aiResponse = strings.TrimSpace(aiResponse)

	aiResponse = strings.ReplaceAll(aiResponse, "```json", "")
	aiResponse = strings.ReplaceAll(aiResponse, "```", "")

	var reBadEscape = regexp.MustCompile(`\\([^"\\/bfnrtu])`)
	aiResponse = reBadEscape.ReplaceAllString(aiResponse, "$1")

	re := regexp.MustCompile(`(?s)\{.*\}`)
	jsonPart := re.FindString(aiResponse)
	if jsonPart == "" {
		return "", fmt.Errorf("kein JSON-Block gefunden in Antwort: %q", aiResponse)
	}

	return jsonPart, nil
}
