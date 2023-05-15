package ui

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/manifoldco/promptui"
	"github.com/sashabaranov/go-openai"
	"github.com/thoas/go-funk"
)

func OpenAiPrompt() {

	var label string
	help := `
		q: quit
		h: help
		s: save the response to a file
		f: add a file to the messages (won't send to openAi until you send a prompt)
		c: clear message list
		
		any other text will be sent to openAI
		`

	fmt.Println("for help type 'h'")

	previousRes := ""
	previousPrompt := ""

	fileNumber := 0
PromptLoop:
	for {
		label = "prompt"
		if fileNumber != 0 {
			label = fmt.Sprintf("%d💾 %s ", fileNumber, label)
		}
		messagesCount := len(service.GetMessages())
		if messagesCount != 0 {
			label = fmt.Sprintf("%d🐦%s ", messagesCount, label)
		}

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

		case "f":
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
				continue PromptLoop
			}
			var selected os.FileInfo

			var files []os.FileInfo

		FileLoop:
			for {
				files, err = ioutil.ReadDir(cwd)

				if err != nil {
					fmt.Println("Error while getting current working directory:", err)
					continue PromptLoop
				}

				if err != nil {
					fmt.Println(err)
					continue PromptLoop
				}
				fileNames := funk.Map(files, func(f os.FileInfo) string {
					return f.Name()
				}).([]string)
				selectPrompt := promptui.Select{
					Label: "File Selection",
					Items: append([]string{"..", "abort"}, fileNames...),
				}

				_, selection, err := selectPrompt.Run()
				if err != nil {
					fmt.Println(err)
					continue PromptLoop
				}
				switch selection {
				case "abort":
					break FileLoop
				case "..":
					cwd = filepath.Dir(cwd)
				default:
					selected = funk.Find(files, func(f os.FileInfo) bool {
						return f.Name() == selection
					}).(os.FileInfo)
					if selected.IsDir() {
						cwd += "/" + selected.Name()
					} else {
						break FileLoop
					}

				}

			}

			fileContent, err := os.ReadFile(cwd + "/" + selected.Name())
			if err != nil {
				fmt.Println(err)
				continue PromptLoop
			}
			service.AddMessage(openai.ChatCompletionMessage{
				Content: string(fileContent),
				Role:    openai.ChatMessageRoleUser,
			})
			fileNumber++

			fmt.Println("added file:", selected.Name())

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

		case "c":
			service.ClearMessages()
			fmt.Println("cleared messages")

		default:
			response, err := service.SendPrompt(context.Background(), userPrompt, os.Stdout)
			if err != nil {
				fmt.Println(err)
				return
			}
			previousRes = response
			fileNumber = 0
		}

		previousPrompt = userPrompt
	}
}
