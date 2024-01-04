//go:build portaudio
// +build portaudio

package command

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MohammadBnei/go-openai-cli/audio"
	"github.com/MohammadBnei/go-openai-cli/service"
)

func AddAudioCommand(commandMap map[string]func(*PromptConfig) error) {
	commandMap["r"] = func(cfg *PromptConfig) error {
		text, err := audio.SpeechToText(context.Background(), &audio.SpeechConfig{MaxMinutes: time.Minute, Lang: "", Detect: false})
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		fmt.Println("Speech: ", text)
		cfg.PreviousPrompt = text

		return nil
	}

	commandMap["rs"] = func(cfg *PromptConfig) error {
		text, err := audio.SpeechToText(context.Background(), &audio.SpeechConfig{MaxMinutes: time.Minute, Lang: "", Detect: false})
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		fmt.Println("Speech: ", text)

		cfg.ChatMessages.AddMessage(text, service.RoleUser)

		return SendPrompt(cfg)

	}
}

func AddAllCommand(commandMap map[string]func(*PromptConfig) error) {
	AddBasicCommand(commandMap)
	AddAudioCommand(commandMap)
}
