package ui

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/golang-module/carbon/v2"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/tigergraph/promptui"
)

func SendAsSystem(systemPrompts map[string]string) error {
	systemPrompt := promptui.Prompt{
		Label: "specify model behavior",
	}
	command, err := systemPrompt.Run()
	if err != nil {
		return err
	}

	systemPrompts[time.Now().String()] = command

	service.AddMessage(service.ChatMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: command,
		Date:    time.Now(),
	})

	saveSystem := promptui.Select{
		Label: "save prompt ?",
		Items: []string{"yes", "no"},
	}

	_, choice, err := saveSystem.Run()
	if err != nil {
		fmt.Println("could not save : ", err)
	}
	if choice == "yes" {
		AddToSystemList(command, time.Now().Format("2006-01-02 15:04:05"))
	}

	return nil
}

func ListSystemCommand() error {
	savedSystemPrompt := viper.GetStringMapString("systems")
	keyStringFromMap := lo.MapToSlice[string, string, string](savedSystemPrompt, func(key string, value string) string {
		return key
	})
	if len(keyStringFromMap) == 0 {
		return errors.New("no saved systems")
	}
	sort.Slice(keyStringFromMap, func(i, j int) bool {
		return carbon.Parse(keyStringFromMap[i]).Gt(carbon.Parse(keyStringFromMap[j]))
	})
	idx, err := fuzzyfinder.FindMulti(
		keyStringFromMap,
		func(i int) string {
			return savedSystemPrompt[keyStringFromMap[i]]
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Date: %s\n%s", keyStringFromMap[i], AddReturnOnWidth(w/3-1, savedSystemPrompt[keyStringFromMap[i]]))
		}),
	)
	if err != nil {
		return err
	}
	for _, id := range idx {
		service.AddMessage(service.ChatMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: savedSystemPrompt[keyStringFromMap[id]],
			Date:    time.Now(),
		})

	}
	return nil
}

func DeleteSystemCommand() error {
	savedSystemPrompt := viper.GetStringMapString("systems")
	keyStringFromMap := lo.MapToSlice[string, string, string](savedSystemPrompt, func(key string, value string) string {
		return key
	})
	if len(keyStringFromMap) == 0 {
		return errors.New("no saved systems")
	}
	sort.Slice(keyStringFromMap, func(i, j int) bool {
		return carbon.Parse(keyStringFromMap[i]).Gt(carbon.Parse(keyStringFromMap[j]))
	})
	idx, err := fuzzyfinder.FindMulti(
		keyStringFromMap,
		func(i int) string {
			return savedSystemPrompt[keyStringFromMap[i]]
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Date: %s\n%s", keyStringFromMap[i], AddReturnOnWidth(w/3-1, savedSystemPrompt[keyStringFromMap[i]]))
		}),
	)
	if err != nil {
		return err
	}
	for _, id := range idx {
		RemoveFromSystemList(keyStringFromMap[id])
		fmt.Printf("removed %s \n", keyStringFromMap[id])
	}
	return nil
}

func AddToSystemList(command string, key string) {
	if key == "" {
		key = time.Now().Format("2006-01-02 15:04:05")
	}
	systems := viper.GetStringMapString("systems")
	systems[key] = command
	viper.Set("systems", systems)
	viper.GetViper().WriteConfig()
}
func RemoveFromSystemList(time string) {
	systems := viper.GetStringMapString("systems")
	delete(systems, time)
	viper.Set("systems", systems)
	viper.GetViper().WriteConfig()
}
