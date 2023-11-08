// +build !portaudio

package command

func AddAllCommand(commandMap map[string]func(*PromptConfig) error) {
	AddBasicCommand(commandMap)
}
