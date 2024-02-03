package ui

import (
	"github.com/MohammadBnei/go-openai-cli/ui/form"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

func NewConfigModel() tea.Model {

	group := []huh.Field{}

	for k, v := range viper.AllSettings() {
		switch v := v.(type) {
		case string:
			group = append(group, huh.NewInput().Title(k).Key(k).Value(&v))
		case bool:
			group = append(group, huh.NewSelect[bool]().Title(k).Key(k).Value(&v).Options(huh.NewOptions[bool](true, false)...))
		}
	}

	group = append(group, huh.NewConfirm().Title("Save").Key("save"))

	return form.NewModel(huh.NewForm(huh.NewGroup(group...)), "Config", nil)
}
