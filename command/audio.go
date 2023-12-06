//go:build portaudio
// +build portaudio

package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/MohammadBnei/go-openai-cli/markdown"
	"github.com/MohammadBnei/go-openai-cli/service"
)

func AddAudioCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["r"] = func(cfg *PromptConfig) error {
		text, err := service.SpeechToText(context.Background(), &service.SpeechConfig{MaxMinutes: time.Minute, Lang: "", Detect: false})
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		fmt.Println("Speech: ", text)
		cfg.PreviousPrompt = text

		return nil
	}

	commandMap["rs"] = func(cfg *PromptConfig) error {
		text, err := service.SpeechToText(context.Background(), &service.SpeechConfig{MaxMinutes: time.Minute, Lang: "", Detect: false})
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		fmt.Println("Speech: ", text)

		cfg.ChatMessages.AddMessage(text, service.RoleUser)

		mdWriter := markdown.NewMarkdownWriter()
		var writer io.Writer
		writer = os.Stdout
		if cfg.MdMode {
			writer = mdWriter
		}
		response, err := service.SendPrompt(&service.SendPromptConfig{
			ChatMessages: cfg.ChatMessages,
			Output:       writer,
			GPTFunc:      service.SendPromptToOpenAi,
		})
		if err != nil {
			return err
		}
		if cfg.MdMode {
			mdWriter.Flush(response)
		}

		return nil
	}
}

func AddAllCommand(commandMap map[string]func(*PromptConfig) error) {
	AddBasicCommand(commandMap)
	AddAudioCommand(commandMap)
}
