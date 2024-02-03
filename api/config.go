package api

import (
	"context"
	"errors"

	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/huggingface"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

type API_TYPE string

const (
	API_OPENAI      = "OPENAI"
	API_HUGGINGFACE = "HUGGINGFACE"
	API_OLLAMA      = "OLLAMA"
)

func GetGenerateFunction() (func(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error), error) {
	model := viper.GetString("model")
	switch viper.GetString("API_TYPE") {
	case API_OPENAI:
		llm, err := openai.New(openai.WithToken(viper.GetString("OPENAI_KEY")), openai.WithModel(model))
		return llm.GenerateContent, err
	case API_HUGGINGFACE:
		llm, err := huggingface.New(huggingface.WithToken(viper.GetString("HUGGINGFACE_KEY")), huggingface.WithModel(model))
		return llm.GenerateContent, err
	case API_OLLAMA:
		llama, err := ollama.New(ollama.WithModel(model))
		return llama.GenerateContent, err
	default:
		return nil, errors.New("invalid api type")
	}
}
