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
	OriginalMessage string   `json:"original_message" yaml:"message"`
	Message         string   `json:"message" yaml:"-"`
	Options         []string `json:"options" yaml:"-"`
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

	optionsRgx, err := regexp.Compile("Choice \\d[.:] (.*)\\n")
	if err != nil {
		return nil, err
	}

	found := optionsRgx.FindAllStringSubmatch(respMsg[len(bla):]+"\n", -1)

	var options []string
	for _, f := range found {
		options = append(options, f[1])
	}

	return &Node{OriginalMessage: respMsg, Message: message, Options: options}, nil
}

func Compress(messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	prompt := make([]openai.ChatCompletionMessage, len(messages)+len(Config.CompressMessages))
	copy(prompt, messages)
	copy(prompt[len(messages):], Config.CompressMessages)

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

	ret := []openai.ChatCompletionMessage{{
		Role:    "system",
		Content: resp.Choices[0].Message.Content,
	}}
	return ret, err
}

func Finish(messages []openai.ChatCompletionMessage) (string, error) {
	prompt := make([]openai.ChatCompletionMessage, len(messages)+len(Config.FinishMessages))
	copy(prompt, messages)
	copy(prompt[len(messages):], Config.FinishMessages)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: prompt,
		},
	)

	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, err
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
