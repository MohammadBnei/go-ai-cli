package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/api"
	"github.com/MohammadBnei/go-openai-cli/markdown"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/atotto/clipboard"
	"github.com/manifoldco/promptui"
	"moul.io/banner"
)

func SendPrompt(pc *PromptConfig) error {
	mdWriter := markdown.NewMarkdownWriter()
	var writer io.Writer
	writer = os.Stdout
	if pc.MdMode {
		writer = mdWriter
	}
	ctx, closer := service.LoadContext(context.Background())
	defer closer()
	stream, err := api.SendPromptToOpenAi(ctx, &api.GPTChanRequest{
		Messages: pc.ChatMessages.Messages,
	})
	if err != nil {
		return err
	}
	response, err := api.PrintTo(stream, writer.Write)
	if err != nil {
		return err
	}
	if pc.MdMode {
		mdWriter.Flush(response)
	}

	pc.ChatMessages.AddMessage(response, service.RoleAssistant)
	return nil
}

func AddFileCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["save"] = func(pc *PromptConfig) error {
		assistantRole := service.RoleAssistant
		lastMessage := pc.ChatMessages.LastMessage(&assistantRole)
		if lastMessage == nil {
			return errors.New("no assistant message found")
		}
		return ui.SaveToFile([]byte(lastMessage.Content), "")
	}

	commandMap["save-chat"] = func(pc *PromptConfig) error {
		return ui.SaveChat(pc.ChatMessages)
	}

	commandMap["load-chat"] = func(pc *PromptConfig) error {
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

	commandMap["meta"] = func(pc *PromptConfig) error {
		system := ""
		yes := ui.YesNoPrompt("Use predefined system ?")
		if yes {
			systems, err := ui.SelectSystemCommand()
			if err != nil {
				return err
			}
			system = strings.Join(systems, "\n")
		}
		additionalSystem, err := ui.StringPrompt("additional system")
		if err != nil {
			return err
		}
		if additionalSystem != "" {
			system = system + "\n" + additionalSystem
		}

		command, err := ui.StringPrompt("command")
		if err != nil {
			return err
		}

		ui.SendCommandOnChat(system, command)
		return nil
	}

	commandMap["file"] = func(pc *PromptConfig) error {
		fileContents, err := ui.FileSelectionFzf("")
		if err != nil {
			return err
		}
		for _, fileContent := range fileContents {
			pc.ChatMessages.AddMessage(fileContent, service.RoleUser)
		}
		return nil
	}
}

func AddConfigCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["markdown"] = func(pc *PromptConfig) error {
		pc.MdMode = !pc.MdMode
		return nil
	}
}

func AddSystemCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["list"] = func(pc *PromptConfig) error {
		messages, err := ui.SelectSystemCommand()
		if err != nil {
			return err
		}
		for _, message := range messages {
			pc.ChatMessages.AddMessage(message, service.RoleAssistant)
		}
		return nil
	}

	commandMap["d-list"] = func(pc *PromptConfig) error {
		return ui.DeleteSystemCommand()
	}

	commandMap["system"] = func(pc *PromptConfig) error {
		message, err := ui.SendAsSystem()
		if err != nil {
			return err
		}
		pc.ChatMessages.AddMessage(message, service.RoleAssistant)
		return nil
	}

	commandMap["filter"] = func(pc *PromptConfig) error {
		messageIds, err := ui.FilterMessages(pc.ChatMessages.Messages)
		if err != nil {
			return err
		}

		for _, id := range messageIds {
			_err := pc.ChatMessages.DeleteMessage(id)
			if _err != nil {
				err = errors.Join(err, _err)
			}
		}

		return err
	}

	commandMap["cli-clear"] = func(pc *PromptConfig) error {
		ui.ClearTerminal()
		return nil
	}

	commandMap["reuse"] = func(pc *PromptConfig) error {
		message, err := ui.ReuseMessage(pc.ChatMessages.Messages)
		if err != nil {
			return err
		}
		pc.PreviousPrompt = message
		return nil
	}

	commandMap["responses"] = func(pc *PromptConfig) error {
		_, err := ui.ShowPreviousMessage(pc.ChatMessages.Messages, pc.MdMode)
		return err
	}

	commandMap["default"] = func(pc *PromptConfig) error {
		commandToAdd, err := ui.SetSystemDefault(false)
		if err != nil {
			return err
		}
		for _, command := range commandToAdd {
			pc.ChatMessages.AddMessage(command, service.RoleAssistant)
		}
		return nil
	}
	commandMap["d-default"] = func(pc *PromptConfig) error {
		commandToAdd, err := ui.SetSystemDefault(true)
		if err != nil {
			return err
		}
		for _, command := range commandToAdd {
			pc.ChatMessages.AddMessage(command, service.RoleAssistant)
		}
		return nil
	}

	commandMap["copy"] = func(pc *PromptConfig) error {
		assistantMessages, _ := pc.ChatMessages.FilterMessages(service.RoleAssistant)
		if len(assistantMessages) < 1 {
			return errors.New("no messages to copy")
		}

		clipboard.WriteAll(assistantMessages[len(assistantMessages)-1].Content)
		fmt.Println("copied to clipboard")
		return nil
	}

	commandMap["clear"] = func(pc *PromptConfig) error {
		pc.ChatMessages.ClearMessages()
		fmt.Println("cleared messages")
		return nil
	}
}

func AddHuggingFaceCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["mask"] = func(cfg *PromptConfig) error {
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

func AddMiscCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["help"] = func(_ *PromptConfig) error {
		fmt.Println(help)
		return nil
	}

	commandMap["quit"] = func(_ *PromptConfig) error {
		fmt.Println(banner.Inline("bye!"))
		os.Exit(0)
		return nil
	}
}

const help = `
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
