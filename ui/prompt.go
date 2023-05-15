package ui

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/ktr0731/go-fuzzyfinder"
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
		f: add files to the messages (won't send to openAi until you send a prompt)
		c: clear messages and files
		
		any other text will be sent to openAI
		`

	fmt.Println("for help type 'h'")

	previousRes := ""
	previousPrompt := ""

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

		case "c":
			service.ClearMessages()
			fileNumber = 0
			fmt.Println("cleared messages")

		case "f":
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
				continue PromptLoop
			}

			selected := []os.FileInfo{}

		FileLoop:
			for {
				files, err := ioutil.ReadDir(cwd)
				if err != nil {
					fmt.Println("Error while getting current working directory:", err)
					continue PromptLoop
				}
				files = append(files, &myFileInfo{"..", 0, 0, time.Now(), true})

				idx, err := fuzzyfinder.FindMulti(
					files,
					func(i int) string {
						return files[i].Name()
					},
					fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
						if i == -1 {
							return ""
						}
						if files[i].IsDir() {
							return "📁 "
						}
						fileContent, err := os.ReadFile(cwd + "/" + files[i].Name())
						if err != nil {
							return fmt.Sprintf("Error while reading file: %s\n", err)
						}
						return fmt.Sprintf("File: %s\nLength: %d",
							files[i].Name(),
							len(string(fileContent)),
						)
					}))

				if err != nil {
					fmt.Println(err)
					continue PromptLoop
				}
				if len(idx) == 1 {
					file := files[idx[0]]

					switch {
					case file.Name() == "..":
						cwd = filepath.Dir(cwd)
					case file.IsDir():
						cwd += "/" + file.Name()
					default:
						selected = funk.Map(idx, func(i int) os.FileInfo {
							return files[i]
						}).([]os.FileInfo)
						break FileLoop
					}
				} else {
					selected = funk.Map(idx, func(i int) os.FileInfo {
						return files[i]
					}).([]os.FileInfo)
					break FileLoop
				}
			}

			for _, file := range selected {
				if file.IsDir() {
					fmt.Printf("%s is a directory, not adding it.\n", file.Name())
					continue
				}

				fileContent, err := os.ReadFile(cwd + "/" + file.Name())
				if err != nil {
					fmt.Println(err)
					continue PromptLoop
				}
				service.AddMessage(openai.ChatCompletionMessage{
					Content: string(fileContent),
					Role:    openai.ChatMessageRoleUser,
				})
				fileNumber++

				fmt.Println("added file:", file.Name())
			}

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

type myFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi myFileInfo) Name() string {
	return fi.name
}

func (fi myFileInfo) Size() int64 {
	return fi.size
}

func (fi myFileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi myFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi myFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi myFileInfo) Sys() interface{} {
	return nil
}
