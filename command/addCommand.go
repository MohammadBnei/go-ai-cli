package command

import "github.com/MohammadBnei/go-openai-cli/service"


type PromptConfig struct {
	MdMode         bool
	ChatMessages   *service.ChatMessages
	PreviousPrompt string
	UserPrompt     string
}

func AddBasicCommand(commandMap map[string]func(*PromptConfig) error) {
	AddFileCommand(commandMap)
	AddConfigCommand(commandMap)
	AddSystemCommand(commandMap)
	AddImageCommand(commandMap)
	AddHuggingFaceCommand(commandMap)
}
