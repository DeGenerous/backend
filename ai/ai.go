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

type imageGeneration struct {
	Prompt string `json:"prompt"`
}

type Compression struct {
	Message string `json:"message"`
	Step    int    `json:"step"`
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

func Compress(messages []openai.ChatCompletionMessage, step int) (*Compression, error) {
	var resp openai.ChatCompletionResponse
	var err error

	promptMessages := make([]openai.ChatCompletionMessage, len(messages))
	copy(promptMessages, messages)
	promptMessages = append(promptMessages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: Config.CompressMessage,
	})

	for tries := 0; tries < maxTries; tries++ {
		resp, err = client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: promptMessages,
			},
		)

		if err != nil {
			continue
		}

		respMsg := resp.Choices[0].Message.Content

		ret := &Compression{
			Message: respMsg,
			Step:    step,
		}

		return ret, nil
	}

	return nil, err
}

func Image(messages []openai.ChatCompletionMessage) (string, error) {
	var resp openai.ChatCompletionResponse
	var err error

	promptMessages := make([]openai.ChatCompletionMessage, len(messages))
	copy(promptMessages, messages)
	promptMessages = append(promptMessages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: Config.ImagePromptPrompt,
	})

	for tries := 0; tries < maxTries; tries++ {
		resp, err = client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: promptMessages,
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
		var image imageGeneration
		if err = json.Unmarshal([]byte(respMsg), &image); err != nil {
			continue
		}

		imgResp, err := client.CreateImage(
			context.Background(),
			openai.ImageRequest{
				Prompt:         image.Prompt + ", " + Config.ImagePrompt,
				N:              1,
				Size:           openai.CreateImageSize256x256,
				ResponseFormat: openai.CreateImageResponseFormatB64JSON,
			},
		)

		if err != nil {
			continue
		}

		return imgResp.Data[0].B64JSON, nil
	}

	return "", err
}
