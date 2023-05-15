package ui

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/disiqueira/gotree"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/manifoldco/promptui"
	"github.com/sashabaranov/go-openai"
	"github.com/thoas/go-funk"
)

func FileSelectionFzf(fileNumber *int) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}

	selected := []os.FileInfo{}

FileLoop:
	for {
		files, err := ioutil.ReadDir(cwd)
		if err != nil {
			fmt.Println("Error while getting current working directory:", err)
			return
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
					root := gotree.New(files[i].Name())
					subFiles, err := ioutil.ReadDir(cwd + "/" + files[i].Name())
					if err != nil {
						return "📁"
					}
					for _, f := range subFiles {
						sub := root.Add(f.Name())
						if f.IsDir() {
							subFiles, err := ioutil.ReadDir(cwd + "/" + files[i].Name())
							if err == nil {
								for _, f := range subFiles {
									sub.Add(f.Name())
								}
							}
						}
					}

					return root.Print()
				}
				fileContent, err := os.ReadFile(cwd + "/" + files[i].Name())
				if err != nil {
					return fmt.Sprintf("Error while reading file: %s\n", err)
				}
				return fmt.Sprintf("File: %s\nType: %s\nLength: %d",
					files[i].Name(),
					http.DetectContentType(fileContent),
					len(string(fileContent)),
				)
			}))

		if err != nil {
			fmt.Println(err)
			return
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
			return
		}
		service.AddMessage(openai.ChatCompletionMessage{
			Content: string(fileContent),
			Role:    openai.ChatMessageRoleUser,
		})
		*fileNumber++

		fmt.Println("added file:", file.Name())
	}
}

func SaveToFile(previousRes string) {
	filePrompt := promptui.Prompt{
		Label: "specify a filename (with extension)",
	}
	filename, err := filePrompt.Run()
	if err != nil {
		return
	}

	if strings.Contains(filename, "/") {
		splitted := strings.Split(filename, "/")
		dw := strings.Join(splitted[:len(splitted)-1], "/")

		if _, err := os.Stat(dw); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dw, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Created directory : " + dw)
		}
	}

	if _, err := os.Stat(filename); err == nil {
		replaceSelect := promptui.Select{
			Label: filename + " exists. Replace ?",
			Items: []string{"Yes", "No"},
		}

		i, _, err := replaceSelect.Run()
		if err != nil {
			fmt.Println(err)
			return
		}

		if i == 1 {
			fmt.Println("abort")
			return
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	f.WriteString(previousRes)
	fmt.Println("saved to", filename)
}
