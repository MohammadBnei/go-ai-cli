package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/huggingface"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"

	openaiHelper "github.com/sashabaranov/go-openai"
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
		llama, err := ollama.New(ollama.WithModel(model), ollama.WithServerURL(viper.GetString("OLLAMA_HOST")))
		return llama.GenerateContent, err
	default:
		return nil, errors.New("invalid api type")
	}
}

func GetApiTypeList() []string {
	return []string{API_OPENAI, API_OLLAMA}
}

func GetApiModelList() ([]string, error) {
	switch viper.GetString("API_TYPE") {
	case API_OPENAI:
		return GetOpenAiModelList()
	case API_HUGGINGFACE:
	case API_OLLAMA:
		return GetOllamaModelList()
	default:
		return nil, errors.New("invalid api type")
	}

	return nil, nil
}

func GetOllamaModelList() ([]string, error) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:11434/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	jsonDataFromHttp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jsonData map[string]any
	err = json.Unmarshal(jsonDataFromHttp, &jsonData)
	if err != nil {
		return nil, err
	}

	models := lo.Map(jsonData["models"].([]any), func(i any, _ int) string {
		return i.(map[string]any)["name"].(string)
	})

	return models, nil
}

func GetOpenAiModelList() ([]string, error) {
	c := openaiHelper.NewClient(viper.GetString("OPENAI_KEY"))
	models, err := c.ListModels(context.Background())
	if err != nil {
		return nil, err
	}

	modelsList := []string{}
	for _, model := range models.Models {
		modelsList = append(modelsList, model.ID)
	}

	return modelsList, nil
}
