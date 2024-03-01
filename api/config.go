package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MohammadBnei/go-ai-cli/config"
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

func GetLlmModel() (llm llms.Model, err error) {
	model := viper.GetString(config.AI_MODEL_NAME)
	switch viper.GetString(config.AI_API_TYPE) {
	case API_OPENAI:
		llm, err = openai.New(openai.WithToken(viper.GetString(config.AI_OPENAI_KEY)), openai.WithModel(model))
	case API_HUGGINGFACE:
		llm, err = huggingface.New(huggingface.WithToken(viper.GetString(config.AI_HUGGINGFACE_KEY)), huggingface.WithModel(model))
	case API_OLLAMA:
		llm, err = ollama.New(ollama.WithModel(model), ollama.WithServerURL(viper.GetString(config.AI_OLLAMA_HOST)))
	default:
		return nil, errors.New("invalid api type")
	}

	return
}

func GetGenerateFunction() (func(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error), error) {
	llm, err := GetLlmModel()
	if err != nil {
		return nil, err
	}

	return llm.GenerateContent, nil
}

func GetApiTypeList() []string {
	return []string{API_OPENAI, API_OLLAMA}
}

func GetApiModelList() ([]string, error) {
	switch viper.GetString(config.AI_API_TYPE) {
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
	req, err := http.NewRequest("GET", viper.GetString(config.AI_OLLAMA_HOST)+"/api/tags", nil)
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
	c := openaiHelper.NewClient(viper.GetString(config.AI_OPENAI_KEY))
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

func GetOpenAiImageModelList() ([]string, error) {
	return []string{"dall-e-3", "dall-e-2"}, nil
}
