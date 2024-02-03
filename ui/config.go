package ui

import (
	"sort"

	"github.com/MohammadBnei/go-openai-cli/ui/form"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

func NewConfigModel() tea.Model {

	group := []huh.Field{}

	stringConfig := make(map[string]string)
	boolConfig := make(map[string]bool)

	for k, v := range viper.AllSettings() {
		switch v := v.(type) {
		case string:
			group = append(group, huh.NewInput().Title(k).Key(k).Value(&v))
			stringConfig[k] = v
		case bool:
			group = append(group, huh.NewSelect[bool]().Title(k).Key(k).Value(&v).Options(huh.NewOptions[bool](true, false)...))
			boolConfig[k] = v
		}
	}

	sort.Slice(group, func(i, j int) bool {
		return group[i].GetKey() < group[j].GetKey()
	})

	group = append(group, huh.NewConfirm().Title("Save").Key("save"))

	onSubmit := func(form *huh.Form) tea.Cmd {
		for k, v := range stringConfig {
			if value := form.GetString(k); value != v {

				viper.Set(k, value)
			}
		}
		for k, v := range boolConfig {
			if value := form.GetBool(k); value != v {
				viper.Set(k, value)
			}
		}
		err := viper.WriteConfig()
		if err != nil {
			return func() tea.Msg {
				return err
			}
		}
		return nil
	}

	return form.NewEditModel(huh.NewForm(huh.NewGroup(group...)), onSubmit)
}
