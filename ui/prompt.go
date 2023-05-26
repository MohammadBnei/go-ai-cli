package ui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/atotto/clipboard"
	"github.com/manifoldco/promptui"
	"github.com/sashabaranov/go-openai"
	"github.com/thoas/go-funk"
)

const help = `
q: quit
h: help
s: save the response to a file
f: add files to the messages (won't send to openAi until you send a prompt)
c: clear messages and files
c (while getting a response): cancel response
copy: copy the last response to the clipboard

any other text will be sent to openAI
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
			IsVimMode: true,
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
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			go func() {
				_, ok := <-c
				if ok {
					cancel()
				}
			}()

			response, err := service.SendPrompt(ctx, userPrompt, os.Stdout)
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

func keyPressListenerLoop(fn func()) {
	consoleReader := bufio.NewReaderSize(os.Stdin, 1)
	input, _ := consoleReader.ReadByte()
	ascii := input

	fmt.Println(ascii)

	// ESC = 27 and Ctrl-C = 3
	if ascii == 27 {
		fn()
	}
}
