package ai

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/sashabaranov/go-openai"
	"strings"

	. "backend/config"
)

var client *openai.Client

func Init(token string) {
	client = openai.NewClient(token)
}

type Node struct {
	OriginalMessage string   `json:"original_message" yaml:"-"`
	Message         string   `json:"story" yaml:"story"`
	Options         []string `json:"options" yaml:"options,omitempty"`
	End             bool     `json:"end" yaml:"end,omitempty"`
	Summary         string   `json:"summary" yaml:"summary,omitempty"`
}

const maxTries = 3

func Generate(messages []openai.ChatCompletionMessage) (*Node, error) {
	var resp openai.ChatCompletionResponse
	var err error

	for tries := 0; tries < maxTries; tries++ {
		resp, err = client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		if err != nil {
			continue
		}

		respMsg := resp.Choices[0].Message.Content
		jsonStart := strings.Index(respMsg, "{")
		jsonEnd := strings.Index(respMsg, "}") + 1

		if jsonStart == -1 || jsonEnd == -1 {
			err = errors.New("unknown response")
			continue
		}

		respMsg = respMsg[jsonStart:jsonEnd]
		var node Node
		if err = json.Unmarshal([]byte(respMsg), &node); err != nil {
			continue
		}

		node.OriginalMessage = respMsg

		return &node, nil
	}

	return nil, err
}

func Compress(messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	prompt := make([]openai.ChatCompletionMessage, len(messages)+len(Config.CompressMessages))
	copy(prompt, messages)
	copy(prompt[len(messages):], Config.CompressMessages)

	var resp openai.ChatCompletionResponse
	var err error

	for tries := 0; tries < maxTries; tries++ {
		resp, err = client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: prompt,
			},
		)

		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	ret := []openai.ChatCompletionMessage{{
		Role:    "system",
		Content: resp.Choices[0].Message.Content,
	}}

	return ret, nil
}

func Finish(messages []openai.ChatCompletionMessage) (string, error) {
	prompt := make([]openai.ChatCompletionMessage, len(messages)+len(Config.FinishMessages))
	copy(prompt, messages)
	copy(prompt[len(messages):], Config.FinishMessages)

	var resp openai.ChatCompletionResponse
	var err error

	for tries := 0; tries < maxTries; tries++ {
		resp, err = client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: prompt,
			},
		)

		if err == nil {
			break
		}
	}
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func Image(prompt string) (string, error) {
	resp, err := client.CreateImage(
		context.Background(),
		openai.ImageRequest{
			Prompt:         prompt + " " + Config.ImagePrompt,
			N:              1,
			Size:           openai.CreateImageSize256x256,
			ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Data[0].B64JSON, nil
}
