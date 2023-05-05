package ai

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"

	. "backend/config"
)

var client *openai.Client

func Init(token string) {
	client = openai.NewClient(token)
}

type Node struct {
	OriginalMessage string   `json:"original_message"`
	Message         string   `json:"message"`
	Options         []string `json:"options"`
}

func Compress(messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	var prompt []openai.ChatCompletionMessage
	copy(prompt, messages)
	prompt = append(prompt, openai.ChatCompletionMessage{
		Role:    "user",
		Content: "Can you make a short story based on our conversation that should contain about a sentence from every step",
	})

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: prompt,
		},
	)

	if err != nil {
		return nil, err
	}
	respMsg := resp.Choices[0].Message.Content

	ret := []openai.ChatCompletionMessage{{
		Role:    "system",
		Content: respMsg,
	}}

	return ret, err
}

func Generate(messages []openai.ChatCompletionMessage) (*Node, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)

	if err != nil {
		return nil, err
	}
	respMsg := resp.Choices[0].Message.Content

	messageRgx, err := regexp.Compile("^((.|\\n)*)Options:")
	if err != nil {
		return nil, err
	}

	bla := messageRgx.FindStringSubmatch(respMsg)
	if len(bla) < 2 {
		return nil, errors.New(respMsg)
	}
	message := strings.TrimSpace(bla[1])

	optionsRgx, err := regexp.Compile("Choice \\d\\. (.*)\\n")
	if err != nil {
		return nil, err
	}

	found := optionsRgx.FindAllStringSubmatch(respMsg+"\n", -1)

	var options []string
	for _, f := range found {
		options = append(options, f[1])
	}

	return &Node{OriginalMessage: respMsg, Message: message, Options: options}, nil
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
