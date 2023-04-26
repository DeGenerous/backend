package config

import (
	"log"
	"os"

	"crypto/rsa"

	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"

	"backend/keys"
)

const configPath = "./config.yaml"

type config struct {
	CORSUrls        []string                       `yaml:"cors_urls"`
	Port            int                            `yaml:"port"`
	Token           string                         `yaml:"token"`
	ImagePrompt     string                         `yaml:"image_prompt"`
	InitialMessages []openai.ChatCompletionMessage `yaml:"initial_messages"`
	Key             *rsa.PrivateKey
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
