package command

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui"
	"github.com/atotto/clipboard"
	"github.com/manifoldco/promptui"
)

func SendPrompt(pc *service.PromptConfig, streamFunc ...func(*api.GPTChanResponse)) error {
	userMsg, _ := pc.ChatMessages.AddMessage(pc.UserPrompt, service.RoleUser)
	assistantMessage, _ := pc.ChatMessages.AddMessage("", service.RoleAssistant)

	pc.ChatMessages.SetAssociatedId(userMsg.Id, assistantMessage.Id)

	ctx, cancel := context.WithCancel(context.Background())
	pc.AddContextWithId(ctx, cancel, userMsg.Id)

	stream, err := api.SendPromptToOpenAi(ctx, &api.GPTChanRequest{
		Messages: pc.ChatMessages.FilterByOpenAIRoles(),
	})
	if err != nil {
		return err
	}

	go func(stream <-chan *api.GPTChanResponse) {
		defer pc.DeleteContext(ctx)
		for v := range stream {
			for _, fn := range streamFunc {
				fn(v)
			}
			previous := pc.ChatMessages.FindById(assistantMessage.Id)
			if previous == nil {
				log.Fatalln("previous message not found")
			}
			previous.Content += string(v.Content)
			pc.ChatMessages.UpdateMessage(*previous)
			if pc.UpdateChan != nil {
				pc.UpdateChan <- *previous
			}
		}
	}(stream)

	return nil
}

func AddFileCommand(commandMap map[string]func(*service.PromptConfig) error) {
	commandMap["save"] = func(pc *service.PromptConfig) error {
		assistantRole := service.RoleAssistant
		lastMessage := pc.ChatMessages.LastMessage(&assistantRole)
		if lastMessage == nil {
			return errors.New("no assistant message found")
		}
		return ui.SaveToFile([]byte(lastMessage.Content), "")
	}

	commandMap["save-chat"] = func(pc *service.PromptConfig) error {
		return ui.SaveChat(pc.ChatMessages)
	}

	commandMap["load-chat"] = func(pc *service.PromptConfig) error {
		startPath, err := ui.StringPrompt("Enter a path to start from")
		if err != nil {
			return err
		}
		fmt.Println(startPath)
		loadedChat, err := ui.LoadChat(startPath)
		if err != nil {
			return err
		}
		pc.ChatMessages = loadedChat
		return nil
	}

	commandMap["file"] = func(pc *service.PromptConfig) error {
		fileContents, err := ui.FileSelectionFzf("")
		if err != nil {
			return err
		}
		for _, fileContent := range fileContents {
			m, err := pc.ChatMessages.AddMessage(fileContent, service.RoleUser)
			if err != nil {
				return err
			}
			m.AsTypeFile()
		}
		return nil
	}
}

func AddSystemCommand(commandMap map[string]func(*service.PromptConfig) error) {
	commandMap["copy"] = func(pc *service.PromptConfig) error {
		assistantMessages, _ := pc.ChatMessages.FilterMessages(service.RoleAssistant)
		if len(assistantMessages) < 1 {
			return errors.New("no messages to copy")
		}

		clipboard.WriteAll(assistantMessages[len(assistantMessages)-1].Content)
		fmt.Println("copied to clipboard")
		return nil
	}

	commandMap["clear"] = func(pc *service.PromptConfig) error {
		pc.ChatMessages.ClearMessages()
		fmt.Println("cleared messages")
		return nil
	}
}

func AddHuggingFaceCommand(commandMap map[string]func(*service.PromptConfig) error) {
	commandMap["mask"] = func(cfg *service.PromptConfig) error {
		maskPrompt := promptui.Prompt{
			Label: "Write a sentance with the character !! as the token to replace",
		}
		pr, err := maskPrompt.Run()
		if err != nil {
			return err
		}
		result, err := api.Mask(strings.Replace(pr, "!!", "[MASK]", -1))
		if err != nil {
			return err
		}
		fmt.Println("Result : ", result)

		return nil
	}
}

const HELP = `
Type \ for options prompt, or \<command_name>.

Available options:

quit: 		quit - Exit the prompt.
help: 		help - Show this help section.

save: 		save the response to a file - Save the last response from OpenAI to a file.
copy: 		copy the last response to the clipboard - Copy the last response from OpenAI to the clipboard.

file: 		add files to the messages - Add files to be included in the conversation messages. These files will not be sent to OpenAI until you send a prompt.
image: 		add an image to the conversation - Add an image to the conversation.
(X) e: 		edit last added image - Edit the last added image.

clear: 		clear messages and files - Clear all conversation messages and files.

system: 	Specify that the next message should be sent as a system message.
filter: 	Remove messages from the conversation history.
reuse: 		Reuse a message.

list: 		List saved system commands.
d-list: 	Delete a saved system command.

default: 	Set the default system commands.
d-default: Unset default system commands.

markdown: Set output mode to markdown.

mask: 		huggingface model. Find a missing word from a sentence.

Any other text will be sent to OpenAI as the prompt.
`
