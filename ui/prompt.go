package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/markdown"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/atotto/clipboard"
	"github.com/samber/lo"
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
	fmt.Println(banner.Inline("go openai cli"), "\n")
	var label string

	mdWriter := markdown.NewMarkdownWriter()
	md := false

	savedSystemPrompt := viper.GetStringMapString("systems")
	if savedSystemPrompt == nil {
		savedSystemPrompt = make(map[string]string)
	}

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
			label = fmt.Sprintf("%d🔤 🧠", totalCharacters)
		}
		if fileNumber != 0 {
			label = fmt.Sprintf("%d💾 %s ", fileNumber, label)
		}

		if md {
			label = fmt.Sprintf("🖥️  %s ", label)

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

		switch strings.TrimSpace(userPrompt) {

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

		case "\\list":
			err := ListSystemCommand()
			if err != nil {
				fmt.Println(err)
			}
		case "\\d-list":
			err := DeleteSystemCommand()
			if err != nil {
				fmt.Println(err)
			}
		case "\\system":
			err := SendAsSystem(savedSystemPrompt)
			if err != nil {
				fmt.Println(err)
			}

		case "\\filter":
			err := FilterMessages()
			if err != nil {
				fmt.Println(err)
			}

		case "md":
			md = !md
			if md {
				fmt.Println("Markdown mode enabled")
			} else {
				fmt.Println("Markdown mode disabled")
			}

		default:
			if strings.HasPrefix(userPrompt, "!md ") {
				userPrompt = userPrompt[4:]
				md = true
			}

			ctx, cancel := context.WithCancel(context.Background())
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			go func() {
				_, ok := <-c
				if ok {
					cancel()
				}
			}()

			fmt.Print("\033[2J") // Clear screen
			fmt.Printf("\033[%d;%dH", 0, 0)
			var writer io.Writer
			writer = os.Stdout
			if md {
				writer = mdWriter
			}
			response, err := service.SendPrompt(ctx, userPrompt, writer)
			signal.Stop(c)
			close(c)
			if md {
				mdWriter.Flush()
			}
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					fmt.Println("❌", err)
				}
				fmt.Println("↩️")
				previousPrompt = userPrompt
				continue PromptLoop
			}

			previousRes = response
			fileNumber = 0
		}

		fmt.Println("\n✅")

		previousPrompt = userPrompt
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
