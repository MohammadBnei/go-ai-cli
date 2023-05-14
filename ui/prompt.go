package ui

import (
	"context"
	"fmt"
	"os"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/manifoldco/promptui"
)

func OpenAiPrompt() {

	label := "What do you want to ask ? "
	help := `
		q: quit
		h: help
		s: save the response to a file
		
		any other text will be sent to openAI
		`

	fmt.Println("for help type 'h'")

	previousRes := ""
	previousPrompt := ""

PromptLoop:
	for {
		prompt := promptui.Prompt{
			Label:     label,
			AllowEdit: false,
			Default:   previousPrompt,
		}

		userPrompt, err := prompt.Run()
		if err != nil {
			fmt.Println(err)
			return
		}

		switch userPrompt {
		case "q":
			break PromptLoop
		case "h":
			fmt.Println(help)

		case "s":
			filePrompt := promptui.Prompt{
				Label: "specify a filename (with extension)",
			}
			filename, err := filePrompt.Run()
			if err != nil {
				continue PromptLoop
			}
			f, err := os.Create(filename)
			if err != nil {
				fmt.Println(err)
				continue PromptLoop
			}
			defer f.Close()

			f.WriteString(previousRes)
			fmt.Println("saved to", filename)
		default:
			response, err := service.SendPrompt(context.Background(), userPrompt, os.Stdout)
			if err != nil {
				fmt.Println(err)
				return
			}
			previousRes = response
		}

		label = "prompt again "
		previousPrompt = userPrompt
	}
}
