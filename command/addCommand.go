package command

import (
	"github.com/MohammadBnei/go-openai-cli/service"
)

func AddBasicCommand(commandMap map[string]func(*service.PromptConfig) error) {
	AddFileCommand(commandMap)
	AddConfigCommand(commandMap)
	AddSystemCommand(commandMap)
	AddMiscCommand(commandMap)
	// AddImageCommand(commandMap)
	AddHuggingFaceCommand(commandMap)
}
