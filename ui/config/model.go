package config

import (
	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/charmbracelet/huh"
)

func newModelSelectForm(value string) (*huh.Form, error) {
	models, err := api.GetApiModelList()
	if err != nil {
		return nil, err
	}

	return huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Key(config.AI_MODEL_NAME).Value(&value).Title("Model").Options(huh.NewOptions[string](models...)...))), nil
}

func newImageModelSelectForm(value string) (*huh.Form, error) {
	models, err := api.GetOpenAiImageModelList()
	if err != nil {
		return nil, err
	}

	return huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Key(config.AI_OPENAI_IMAGE_MODEL).Value(&value).Title("OpenAI Image Model").Options(huh.NewOptions[string](models...)...))), nil
}

func newApiTypeSelectForm(value string) *huh.Form {
	return huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Key(config.AI_API_TYPE).Value(&value).Title("API Type").Options(huh.NewOptions[string](api.GetApiTypeList()...)...)))
}
