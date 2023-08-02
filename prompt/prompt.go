package prompt

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/atotto/clipboard"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/tigergraph/promptui"
)

const help = `
Type \ for options prompt.

Available options:

quit: 					quit - Exit the prompt.
help: 					help - Show this help section.

save: 					save the response to a file - Save the last response from OpenAI to a file.
copy: 					copy the last response to the clipboard - Copy the last response from OpenAI to the clipboard.

file: 					add files to the messages - Add files to be included in the conversation messages. These files will not be sent to OpenAI until you send a prompt.
image: 					add an image to the conversation - Add an image to the conversation.
(X) e: 					edit last added image - Edit the last added image.

clear: 					clear messages and files - Clear all conversation messages and files.

system: 				Specify that the next message should be sent as a system message.
filter: 				Remove messages from the conversation history.
reuse: 					Reuse a message.

list: 					List saved system commands.
d-list: 				Delete a saved system command.

default: 				Set the default system commands.
d-default: 			Unset default system commands.

markdown: 			Set output mode to markdown.

mask: 					huggingface model. Find a missing word from a sentence.

Any other text will be sent to OpenAI as the prompt.
`

func OpenAiPrompt() {
	var label string

	if clipboard.Unsupported {
		fmt.Println("clipboard is not avalaible on this os")
	}

	commandMap := make(map[string]func(*command.PromptConfig) error)

	commandMap["quit"] = func(_ *command.PromptConfig) error {
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
		MdMode: viper.GetBool("md"),
	}

	defaulSystemPrompt := viper.GetStringMapString("default-systems")
	savedSystemPrompt := viper.GetStringMapString("systems")
	for k := range defaulSystemPrompt {
		service.AddMessage(service.ChatMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: savedSystemPrompt[k],
			Date:    time.Now(),
		})
	}

PromptLoop:
	for {
		label = "prompt"
		totalCharacters := lo.Reduce[service.ChatMessage, int](service.GetMessages(), func(acc int, elem service.ChatMessage, _ int) int {
			return acc + len(elem.Content)
		}, 0)
		if totalCharacters != 0 {
			label = fmt.Sprintf("%d🔤 🧠", totalCharacters)
		}
		if promptConfig.FileNumber != 0 {
			label = fmt.Sprintf("%d💾 %s ", promptConfig.FileNumber, label)
		}

		if promptConfig.MdMode {
			label = fmt.Sprintf("🖥️  %s ", label)
		}

		prompt := promptui.Prompt{
			Label:     label,
			AllowEdit: false,
			Default:   promptConfig.PreviousPrompt,
		}

		userPrompt, err := prompt.Run()
		if err != nil {
			fmt.Println(err)
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			continue PromptLoop

		}

		cmd := strings.TrimSpace(userPrompt)
		keys := lo.Keys[string](commandMap)

		switch cmd {
		case "\\":
			selection, err2 := fuzzyfinder.Find(keys, func(i int) string {
				return keys[i]
			})
			if err2 != nil {
				fmt.Println(err)
				continue PromptLoop
			}

			err = commandMap[keys[selection]](promptConfig)
		case "h":
			fmt.Println(help)
		default:
			promptConfig.UserPrompt = cmd
			err = command.SendPrompt(promptConfig)
		}

		if err != nil {
			fmt.Println("❌", err)
		}

		// case "e":
		// 	lastImagePath = AskForEditImage(lastImagePath)

		fmt.Println("\n✅")
		promptConfig.PreviousPrompt = userPrompt
	}
}

var clear map[string]func() = make(map[string]func())

func CallClear() {
	if _, ok := clear["linux"]; !ok {
		clear["linux"] = func() {
			cmd := exec.Command("clear") //Linux example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
	}
	if _, ok := clear["windows"]; !ok {
		clear["windows"] = func() {
			cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
	}
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		clear["linux"]()
	}
}
