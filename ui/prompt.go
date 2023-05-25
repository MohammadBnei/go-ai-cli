package ui

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/atotto/clipboard"
	"github.com/manifoldco/promptui"
	"github.com/mattn/go-tty"
	"github.com/sashabaranov/go-openai"
	"github.com/thoas/go-funk"
)

func OpenAiPrompt() {

	var label string
	help := `
		q: quit
		h: help
		s: save the response to a file
		f: add files to the messages (won't send to openAi until you send a prompt)
		c: clear messages and files
		c (while getting a response): cancel response
		copy: copy the last response to the clipboard
		
		any other text will be sent to openAI
		`

	fmt.Println("for help type 'h'")

	previousRes := ""
	previousPrompt := ""

	// lastImagePath := ""

	fileNumber := 0
PromptLoop:
	for {
		label = "prompt"
		totalCharacters := funk.Reduce(service.GetMessages(), func(acc int, elem openai.ChatCompletionMessage) int {
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
			continue PromptLoop
		}

		switch userPrompt {
		case "q":
			break PromptLoop
		case "h":
			fmt.Println(help)

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
			} else {
				clipboard.WriteAll(previousRes)
				fmt.Println("copied to clipboard")
			}

		case "c":
			service.ClearMessages()
			fileNumber = 0
			fmt.Println("cleared messages")

		case "f":
			FileSelectionFzf(&fileNumber)

		default:
			ctx, cancel := context.WithCancel(context.Background())
			go func(ctx context.Context, cancel context.CancelFunc) {
				tty, err := tty.Open()
				if err != nil {
					fmt.Println(err)
					return
				}
				defer tty.Close()

				for {
					select {
					case <-ctx.Done():
						return
					default:
						r, err := tty.ReadRune()
						if err != nil {
							fmt.Println(err)
							return
						}
						if r == 'c' {
							cancel()
							return
						}
					}
				}
			}(ctx, cancel)

			response, err := service.SendPrompt(ctx, userPrompt, os.Stdout)
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
