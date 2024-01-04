package prompt

import (
	"fmt"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/atotto/clipboard"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/fatih/color"
)

func OpenAiPrompt() {
	var label string

	if clipboard.Unsupported {
		fmt.Println("clipboard is not avalaible on this os")
	}

	commandMap := make(map[string]func(*command.PromptConfig) error)

	fmt.Println("for help type 'h'")

	command.AddAllCommand(commandMap)

	promptConfig := &command.PromptConfig{
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
		label = "prompt"
		tokens := promptConfig.ChatMessages.TotalTokens
		if tokens != 0 {
			label = fmt.Sprintf("%d🔤 🧠", tokens)
		}

		if promptConfig.MdMode {
			label = fmt.Sprintf("🖥️  %s ", label)
		}

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
