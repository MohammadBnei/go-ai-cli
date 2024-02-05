package command

import (
	"github.com/MohammadBnei/go-ai-cli/service"
)

func AddBasicCommand(commandMap map[string]func(*service.PromptConfig) error) {
	AddFileCommand(commandMap)
	AddSystemCommand(commandMap)
	// AddImageCommand(commandMap)
	AddHuggingFaceCommand(commandMap)
}
