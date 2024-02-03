//go:build !portaudio
// +build !portaudio

package command

import "github.com/MohammadBnei/go-openai-cli/service"

func AddAllCommand(commandMap map[string]func(*service.PromptConfig) error) {
	AddBasicCommand(commandMap)
}
