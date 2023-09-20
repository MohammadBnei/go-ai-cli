package command

import (
	"errors"
	"fmt"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/atotto/clipboard"
	"github.com/manifoldco/promptui"
)

func AddFileCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["save"] = func(cfg *PromptConfig) error {
		return ui.SaveToFile([]byte(cfg.PreviousRes), "")
	}

	commandMap["file"] = func(cfg *PromptConfig) error {
		return ui.FileSelectionFzf(&cfg.FileNumber)
	}
}

func AddConfigCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["markdown"] = func(cfg *PromptConfig) error {
		cfg.MdMode = !cfg.MdMode
		return nil
	}
}

func AddSystemCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["list"] = func(pc *PromptConfig) error {
		return ui.SelectSystemCommand()
	}

	commandMap["d-list"] = func(pc *PromptConfig) error {
		return ui.DeleteSystemCommand()
	}

	commandMap["system"] = func(pc *PromptConfig) error {
		return ui.SendAsSystem()
	}

	commandMap["filter"] = func(pc *PromptConfig) error {
		return ui.FilterMessages()
	}

	commandMap["cli-clear"] = func(pc *PromptConfig) error {
		return ui.ClearTerminal()
	}

	commandMap["reuse"] = func(pc *PromptConfig) error {
		message, err := ui.ReuseMessage()
		if err != nil {
			return err
		}
		pc.PreviousPrompt = message
		return nil
	}

	commandMap["responses"] = func(pc *PromptConfig) error {
		_, err := ui.ShowPreviousMessage(pc.MdMode)
		return err
	}

	commandMap["default"] = func(pc *PromptConfig) error {
		return ui.SetSystemDefault(false)
	}
	commandMap["d-default"] = func(pc *PromptConfig) error {
		return ui.SetSystemDefault(true)
	}

	commandMap["copy"] = func(pc *PromptConfig) error {
		if pc.PreviousRes == "" {
			return errors.New("nothing to copy")
		}
		clipboard.WriteAll(pc.PreviousRes)
		fmt.Println("copied to clipboard")
		return nil
	}

	commandMap["clear"] = func(pc *PromptConfig) error {
		service.ClearMessages()
		pc.FileNumber = 0
		fmt.Println("cleared messages")
		return nil
	}
}

func AddImageCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["image"] = func(cfg *PromptConfig) error {
		return ui.AskForImage()
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
		result, err := service.Mask(strings.Replace(pr, "!!", "[MASK]", -1))
		if err != nil {
			return err
		}
		fmt.Println("Result : ", result)

		return nil
	}
}
