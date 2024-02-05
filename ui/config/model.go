package config

import (
	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/charmbracelet/huh"
)

func newModelSelectForm(value string) (*huh.Form, error) {
	models, err := api.GetApiModelList()
	if err != nil {
		return nil, err
	}

	return huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Key("model").Value(&value).Title("Model").Options(huh.NewOptions[string](models...)...))), nil
}

func newApiTypeSelectForm(value string) *huh.Form {
	return huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Key("api_type").Value(&value).Title("API Type").Options(huh.NewOptions[string](api.GetApiTypeList()...)...)))
}
