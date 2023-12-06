package prompt

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/markdown"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/atotto/clipboard"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"moul.io/banner"
)

const help = `
Type \ for options prompt, or \<command_name>.

Available options:

quit: 		quit - Exit the prompt.
help: 		help - Show this help section.

save: 		save the response to a file - Save the last response from OpenAI to a file.
copy: 		copy the last response to the clipboard - Copy the last response from OpenAI to the clipboard.

file: 		add files to the messages - Add files to be included in the conversation messages. These files will not be sent to OpenAI until you send a prompt.
image: 		add an image to the conversation - Add an image to the conversation.
(X) e: 		edit last added image - Edit the last added image.

clear: 		clear messages and files - Clear all conversation messages and files.

system: 	Specify that the next message should be sent as a system message.
filter: 	Remove messages from the conversation history.
reuse: 		Reuse a message.

list: 		List saved system commands.
d-list: 	Delete a saved system command.

default: 	Set the default system commands.
d-default: Unset default system commands.

markdown: Set output mode to markdown.

mask: 		huggingface model. Find a missing word from a sentence.

Any other text will be sent to OpenAI as the prompt.
`

func OpenAiPrompt() {
	var label string

	if clipboard.Unsupported {
		fmt.Println("clipboard is not avalaible on this os")
	}

	commandMap := make(map[string]func(*command.PromptConfig) error)

	commandMap["quit"] = func(_ *command.PromptConfig) error {
		fmt.Println(banner.Inline("bye!"))
		os.Exit(0)
		return nil
	}

	fmt.Println("for help type 'h'")
	commandMap["help"] = func(_ *command.PromptConfig) error {
		fmt.Println(help)
		return nil
	}

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

		// prompt := promptui.Prompt{
		// 	Label:     label,
		// 	AllowEdit: false,
		// 	Default:   promptConfig.PreviousPrompt,
		// }

		userPrompt, err := ui.BasicPrompt(label)
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
			command, ok := commandMap[cmd[1:]]
			if !ok {
				fmt.Println("command not found")
				commandMap["help"](promptConfig)
				continue PromptLoop
			}
			err = command(promptConfig)
		case cmd == "h":
			commandMap["help"](promptConfig)
			continue PromptLoop
		case cmd == "q":
			commandMap["quit"](promptConfig)
		default:
			ui.ClearTerminal()
			promptConfig.ChatMessages.AddMessage(userPrompt, service.RoleUser)

			// writers := io.MultiWriter(os.Stdout)

			mdWriter := markdown.NewMarkdownWriter()
			var writer io.Writer
			writer = os.Stdout
			if promptConfig.MdMode {
				writer = mdWriter
			}
			response, err := service.SendPrompt(&service.SendPromptConfig{
				ChatMessages: promptConfig.ChatMessages,
				Output:       writer,
				GPTFunc:      service.SendPromptToOpenAi,
			})
			if err != nil {
				fmt.Println("❌", err)
			}
			if promptConfig.MdMode {
				mdWriter.Flush(response)
			}

			promptConfig.ChatMessages.AddMessage(response, service.RoleAssistant)
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
