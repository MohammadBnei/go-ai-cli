package prompt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
	"github.com/spf13/viper"
)

func GetLabel(pc *service.PromptConfig) string {
	label := "prompt"
	tokens := pc.ChatMessages.TotalTokens
	if tokens != 0 {
		label = fmt.Sprintf("%d🔤 🧠", tokens)
	}

	if pc.MdMode {
		label = fmt.Sprintf("🖥️  %s ", label)
	}

	return label
}

func CommandSelectionFactory() func(cmd string, pc *service.PromptConfig) error {
	commandMap := make(map[string]func(*service.PromptConfig) error)

	command.AddAllCommand(commandMap)
	keys := lo.Keys[string](commandMap)

	return func(cmd string, pc *service.PromptConfig) error {

		var err error

		switch {
		case cmd == "":
			commandMap["help"](pc)
		case cmd == "\\":
			selection, err2 := fuzzyfinder.Find(keys, func(i int) string {
				return keys[i]
			})
			if err2 != nil {
				return err2
			}

			err = commandMap[keys[selection]](pc)
		case strings.HasPrefix(cmd, "\\"):
			command, ok := commandMap[cmd[1:]]
			if !ok {
				return errors.New("command not found")
			}
			err = command(pc)
		}

		return err
	}
}

func OpenAiPrompt() {
	var label string

	if clipboard.Unsupported {
		fmt.Println("clipboard is not avalaible on this os")
	}

	commandMap := make(map[string]func(*service.PromptConfig) error)

	fmt.Println("for help type 'h'")

	command.AddAllCommand(commandMap)

	promptConfig := &service.PromptConfig{
		MdMode:       viper.GetBool("md"),
		ChatMessages: service.NewChatMessages("default"),
	}

	defaulSystemPrompt := viper.GetStringMapString("default-systems")
	savedSystemPrompt := viper.GetStringMapString("systems")
	for k := range defaulSystemPrompt {
		promptConfig.ChatMessages.AddMessage(savedSystemPrompt[k], service.RoleSystem)
	}

	// history := []string{}

PromptLoop:
	for {
		label = GetLabel(promptConfig)

		userPrompt, err := ui.StringPrompt(label)
		if err != nil {
			fmt.Println(err)
			continue PromptLoop
		}

		if userPrompt == "" {
			continue PromptLoop
		}

		cmd := strings.TrimSpace(userPrompt)
		keys := lo.Keys[string](commandMap)

		switch {
		case cmd == "":
			commandMap["help"](promptConfig)
		case cmd == "\\":
			selection, err2 := fuzzyfinder.Find(keys, func(i int) string {
				return keys[i]
			})
			if err2 != nil {
				fmt.Println(err)
				continue PromptLoop
			}

			err = commandMap[keys[selection]](promptConfig)
		case strings.HasPrefix(cmd, "\\"):
			color.Green(cmd)
			command, ok := commandMap[cmd[1:]]
			if !ok {
				fmt.Println("command not found")
				commandMap["help"](promptConfig)
				continue PromptLoop
			}
			err = command(promptConfig)
		default:
			color.Cyan(userPrompt)
			promptConfig.ChatMessages.AddMessage(userPrompt, service.RoleUser)
			err = command.SendPrompt(promptConfig)
		}

		if err != nil {
			fmt.Println("❌", err)
		}

		// case "e":
		// 	lastImagePath = AskForEditImage(lastImagePath)

		fmt.Println("\n✅")
		promptConfig.PreviousPrompt = userPrompt

		// if viper.GetBool("autoSave") {
		// 	service.
		// }
	}
}
