package ai

import (
	"os"

	"gopkg.in/yaml.v3"
)

const (
	configFileName = ".ako/llm.config.yaml"
)

type Config struct {
	Ollama struct {
		Enable bool   `yaml:"enable"`
		Host   string `yaml:"host"`
		Model  string `yaml:"model"`
	} `yaml:"ollama"`
	Gemini struct {
		Enable bool   `yaml:"enable"`
		Host   string `yaml:"host"`
		Model  string `yaml:"model"`
		APIKey string `yaml:"api_key"`
	} `yaml:"gemini"`
	Anthropic struct {
		Enable bool   `yaml:"enable"`
		Host   string `yaml:"host"`
		Model  string `yaml:"model"`
		APIKey string `yaml:"api_key"`
	} `yaml:"anthropic"`
	OpenAI struct {
		Enable bool   `yaml:"enable"`
		Host   string `yaml:"host"`
		Model  string `yaml:"model"`
		APIKey string `yaml:"api_key"`
	} `yaml:"openai"`
}

var GlobalConfig Config

var globalConfigNotExists bool

func init() {
	if err := ReadConfig(); err != nil {
		globalConfigNotExists = true
	}
}

func ReadConfig() error {
	configFile, err := os.Open(configFileName)
	if err != nil {
		return err
	}
	defer configFile.Close()

	decoder := yaml.NewDecoder(configFile)
	if err := decoder.Decode(&GlobalConfig); err != nil {
		return err
	}

	return nil
}

func SaveConfig() error {
	if err := os.MkdirAll(".ako", os.ModePerm); err != nil {
		return err
	}

	configFile, err := os.Create(configFileName)
	if err != nil {
		return err
	}
	defer configFile.Close()

	encoder := yaml.NewEncoder(configFile)
	if err := encoder.Encode(GlobalConfig); err != nil {
		return err
	}

	return nil
}

func InitConfig() error {
	GlobalConfig = Config{
		Ollama: struct {
			Enable bool   `yaml:"enable"`
			Host   string `yaml:"host"`
			Model  string `yaml:"model"`
		}{
			Enable: true,
			Host:   "http://localhost:11434",
			Model:  "gemma3:4b",
		},
	}

	if err := SaveConfig(); err != nil {
		return err
	}

	return nil
}
