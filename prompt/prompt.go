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
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/tigergraph/promptui"
	"moul.io/banner"
)

const help = `
Available options:

q: quit - Exit the prompt.
h: help - Show this help section.
s: save the response to a file - Save the last response from OpenAI to a file.
f: add files to the messages - Add files to be included in the conversation messages. These files will not be sent to OpenAI until you send a prompt.
c: clear messages and files - Clear all conversation messages and files.
copy: copy the last response to the clipboard - Copy the last response from OpenAI to the clipboard.
i: add an image to the conversation - Add an image to the conversation.
e: edit last added image - Edit the last added image.

Commands that can be used as the prompt:

Any other text will be sent to OpenAI as the prompt.

Additional commands:

\system - Specify that the next message should be sent as a system message.
\filter - Filter messages - Remove messages from the conversation history.
\list - List saved system commands - List all saved system commands.
\d-list - Delete a saved system command - Delete a saved system command.
`

func OpenAiPrompt() {
	fmt.Print(banner.Inline("go openai cli"), "\n\n")
	var label string

	if clipboard.Unsupported {
		fmt.Println("clipboard is not avalaible on this os")
	}

	commandMap := make(map[string]func(*command.PromptConfig) error)

	commandMap["q"] = func(_ *command.PromptConfig) error {
		os.Exit(0)
		return nil
	}
	commandMap["h"] = func(_ *command.PromptConfig) error {
		fmt.Println(help)
		return nil
	}

	command.AddAllCommand(commandMap)

	promptConfig := &command.PromptConfig{
		MdMode: viper.GetBool("md"),
	}

	if promptConfig.MdMode {
		service.AddMessage(service.ChatMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: "respond in markdown format only, with a title.",
			Date:    time.Now(),
		})
	}

	savedSystemPrompt := viper.GetStringMapString("systems")
	if savedSystemPrompt == nil {
		savedSystemPrompt = make(map[string]string)
	}

	promptConfig.SystemPrompts = savedSystemPrompt

	fmt.Println("for help type 'h'")

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
		f, ok := commandMap[cmd]
		if !ok {
			err = command.SendPrompt(promptConfig)
		} else {
			err = f(promptConfig)
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
