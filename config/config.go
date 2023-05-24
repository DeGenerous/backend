package config

import (
	"fmt"
	"log"
	"os"

	"crypto/rsa"

	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"

	"backend/keys"
)

const configPath = "./config.yaml"

type config struct {
	CORSUrls          []string                       `yaml:"cors_urls"`
	Domain            string                         `yaml:"domain"`
	Port              int                            `yaml:"port"`
	DiscordToken      string                         `yaml:"discord_token"`
	ChannelID         string                         `yaml:"channel_id"`
	OpenAIToken       string                         `yaml:"ai_token"`
	ImagePrompt       string                         `yaml:"image_prompt"`
	ImagePromptPrompt string                         `yaml:"image_prompt_prompt"`
	MaxSteps          int                            `yaml:"max_steps"`
	CompressionLimit  int                            `yaml:"compression_limit"`
	PromptMessages    []openai.ChatCompletionMessage `yaml:"prompt_messages"`
	CompressMessage   string                         `yaml:"compress_message"`
	CompressionPrompt string                         `yaml:"compression_prompt"`
	Contracts         struct {
		Explorer string `yaml:"explorer"`
		Nft      string `yaml:"nft"`
	} `yaml:"contracts"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
		Username string `yaml:"user"`
		Password string `yaml:"pass"`
		Redis    string `yaml:"redis"`
	} `yaml:"database"`
	Key *rsa.PrivateKey
}

var Config config

func (cfg *config) Read() error {
	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := yaml.Unmarshal(file, cfg); err != nil {
		log.Fatal(err)
	}

	key, err := keys.ReadKey()
	if err != nil {
		key, err = keys.GenerateKey()
		if err != nil {
			return err
		}
	}
	cfg.Key = key

	return nil
}

func (cfg *config) CompressionPromptMessages(step int, summary string) []openai.ChatCompletionMessage {
	promptMessages := make([]openai.ChatCompletionMessage, len(cfg.PromptMessages))
	copy(promptMessages, cfg.PromptMessages)
	promptMessages[0].Content += "\n" + fmt.Sprintf(cfg.CompressionPrompt, step, summary)

	return promptMessages
}
