package ai

import (
	"context"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
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

	message := strings.TrimSpace(messageRgx.FindStringSubmatch(respMsg)[1])

	optionsRgx, err := regexp.Compile("\\d\\. (.*)\\n")
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