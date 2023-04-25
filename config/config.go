package config

import (
	"backend/keys"
	"log"
	"os"

	"crypto/rsa"

	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v3"
)

const configPath = "./config.yaml"

type config struct {
	CORSUrls        []string                       `yaml:"cors_urls"`
	Token           string                         `yaml:"token"`
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
