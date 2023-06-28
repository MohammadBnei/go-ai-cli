package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/atotto/clipboard"
	"github.com/manifoldco/promptui"
	"github.com/samber/lo"
)

const help = `
Available options:

q: quit - Exit the prompt.
h: help - Show this help section.
s: save the response to a file - Save the last response from OpenAI to a file.
f: add files to the messages - Add files to be included in the conversation messages. These files will not be sent to OpenAI until you send a prompt.
c: clear messages and files - Clear all conversation messages and files.
copy: copy the last response to the clipboard - Copy the last response from OpenAI to the clipboard.

Commands that can be used as the prompt:

Any other text will be sent to OpenAI as the prompt.

Additional commands:

\system - Specify that the next message should be sent as a system message.
\filter - Filter messages - Remove messages from the conversation history.
`

func OpenAiPrompt() {
	var label string

	fmt.Println("for help type 'h'")

	previousRes := ""
	previousPrompt := ""

	// lastImagePath := ""

	fileNumber := 0
PromptLoop:
	for {
		label = "prompt"
		totalCharacters := lo.Reduce[service.ChatMessage, int](service.GetMessages(), func(acc int, elem service.ChatMessage, _ int) int {
			return acc + len(elem.Content)
		}, 0)
		if totalCharacters != 0 {
			label = fmt.Sprintf("%d🔤 follow up", totalCharacters)
		}
		if fileNumber != 0 {
			label = fmt.Sprintf("%d💾 %s ", fileNumber, label)
		}

		prompt := promptui.Prompt{
			Label:     label,
			AllowEdit: false,
			Default:   previousPrompt,
		}

		userPrompt, err := prompt.Run()
		if err != nil {
			fmt.Println(err)
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			continue PromptLoop
		}

		switch userPrompt {
		case "q":
			break PromptLoop
		case "h":
			fmt.Print(help)

		case "s":
			SaveToFile([]byte(previousRes))

		case "i":
			// lastImagePath = AskForImage()
			AskForImage()

		// case "e":
		// 	lastImagePath = AskForEditImage(lastImagePath)

		case "copy":
			if clipboard.Unsupported {
				fmt.Println("clipboard is not avalaible on this os")
				continue PromptLoop
			}
			if previousRes == "" {
				fmt.Println("nothing to copy")
				continue PromptLoop
			}

			clipboard.WriteAll(previousRes)
			fmt.Println("copied to clipboard")

		case "c":
			service.ClearMessages()
			fileNumber = 0
			fmt.Println("cleared messages")

		case "f":
			FileSelectionFzf(&fileNumber)

		case "\\system":
			err := SendAsSystem()
			if err != nil {
				fmt.Println(err)
			}

		case "\\filter":
			err := FilterMessages()
			if err != nil {
				fmt.Println(err)
			}

		default:
			ctx, cancel := context.WithCancel(context.Background())
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			go func() {
				_, ok := <-c
				if ok {
					cancel()
				}
			}()

			response, err := service.SendPrompt(ctx, userPrompt, os.Stdout)
			signal.Stop(c)
			close(c)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					fmt.Println(err)
				}
				previousPrompt = userPrompt
				continue PromptLoop
			}
			previousRes = response
			fileNumber = 0
		}

		previousPrompt = userPrompt
	}
}
