package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MohammadBnei/go-openai-cli/service"
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
		AddToSystemList(command)
	}

	return nil
}

func ListSystemCommand() error {
	savedSystemPrompt := viper.GetStringMapString("systems")
	slicedList := lo.MapToSlice[string, string, string](savedSystemPrompt, func(key string, value string) string {
		if key == "" || value == "" {
			return ""
		}
		return fmt.Sprintf("%s - %s", key, value)
	})
	if len(slicedList) == 0 {
		return errors.New("no saved systems")
	}
	slicedList = append(slicedList, "cancel")
	systemPrompt := promptui.Select{
		Items: slicedList,
		Label: "Choose a previous system command",
	}

	id, choice, err := systemPrompt.Run()
	if err != nil {
		return err
	}
	if choice == "" {
		return errors.New("there was an error in the command")
	}
	if choice == "cancel" {
		return nil
	}
	service.AddMessage(service.ChatMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: strings.Split(slicedList[id], " - ")[1],
		Date:    time.Now(),
	})
	return nil
}

func DeleteSystemCommand() error {
	savedSystemPrompt := viper.GetStringMapString("systems")
	slicedList := lo.MapToSlice[string, string, string](savedSystemPrompt, func(key string, value string) string {
		return fmt.Sprintf("%s - %s", key, value)
	})
	if len(slicedList) == 0 {
		return errors.New("no saved systems")
	}
	systemPrompt := promptui.Select{
		Items: slicedList,
		Label: "Choose a previous system command",
	}

	idx, _, err := systemPrompt.Run()
	if err != nil {
		return err
	}
	RemoveFromSystemList(strings.Split(slicedList[idx], " - ")[0])
	fmt.Printf("removed %s \n", slicedList[idx])
	return nil
}

func AddToSystemList(command string) {
	systems := viper.GetStringMapString("systems")
	systems[time.Now().Format("2006-01-02 15:04:05")] = command
	viper.Set("systems", systems)
	viper.GetViper().WriteConfig()
}
func RemoveFromSystemList(time string) {
	systems := viper.GetStringMapString("systems")
	delete(systems, time)
	viper.Set("systems", systems)
	viper.GetViper().WriteConfig()
}
