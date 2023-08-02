package command

type PromptConfig struct {
	MdMode         bool
	UserPrompt     string
	PreviousPrompt string
	PreviousRes    string
	FileNumber     int
	SystemPrompts  map[string]string
}

func AddBasicCommand(commandMap map[string]func(*PromptConfig) error) {
	AddFileCommand(commandMap)
	AddConfigCommand(commandMap)
	AddSystemCommand(commandMap)
	AddImageCommand(commandMap)
	AddHuggingFaceCommand(commandMap)
}
