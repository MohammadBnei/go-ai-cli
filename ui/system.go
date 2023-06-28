package ui

import (
	"time"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/manifoldco/promptui"
	"github.com/sashabaranov/go-openai"
)

func SendAsSystem() error {
	systemPrompt := promptui.Prompt{
		Label: "specify model behavior",
	}
	command, err := systemPrompt.Run()
	if err != nil {
		return err
	}

	service.AddMessage(service.ChatMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: command,
		Date:    time.Now(),
	})

	return nil
}
