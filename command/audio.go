//go:build portaudio

package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MohammadBnei/go-openai-cli/service"
)

func AddAudioCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["r"] = func(cfg *PromptConfig) error {
		text, err := service.SpeechToText(context.Background(), "", 30*time.Second)
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		fmt.Println("Speech: ", text)
		cfg.PreviousPrompt = text

		return nil
	}

	commandMap["rs"] = func(cfg *PromptConfig) error {
		text, err := service.SpeechToText(context.Background(), "", 30*time.Second)
		if err != nil {
			return err
		}
		fmt.Print("\033[2J") // Clear screen
		fmt.Printf("\033[%d;%dH", 0, 0)
		text = strings.TrimSpace(text)
		fmt.Println("Speech: ", text)

		cfg.UserPrompt = text
		return SendPrompt(cfg)
	}
}

func AddAllCommand(commandMap map[string]func(*PromptConfig) error) {
	AddBasicCommand(commandMap)
	AddAudioCommand(commandMap)
}
