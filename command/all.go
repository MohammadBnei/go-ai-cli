//go:build !portaudio
// +build !portaudio

package command

import "github.com/MohammadBnei/go-ai-cli/service"

func AddAllCommand(commandMap map[string]func(*service.PromptConfig) error) {
	AddBasicCommand(commandMap)
}
