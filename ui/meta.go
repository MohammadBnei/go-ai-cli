package ui

import (
	"context"
	"errors"

	"github.com/MohammadBnei/go-openai-cli/api"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/samber/lo"
)

func SendCommandOnChat(system string, command string) (*service.ChatMessages, error) {
	if command == "" {
		return nil, errors.New("no command to send")
	}

	loadedChat, err := LoadChatOnlyAssistant("")
	if err != nil {
		return nil, err
	}

	messagesToString := lo.Reduce[service.ChatMessage, string](loadedChat.Messages, func(acc string, item service.ChatMessage, _ int) string {
		return acc + item.Content
	}, "")

	metaChatMessages := service.NewChatMessages("meta")
	if system != "" {
		metaChatMessages.AddMessage(system, service.RoleSystem)
	}

	metaChatMessages.AddMessage(messagesToString, service.RoleUser)
	metaChatMessages.AddMessage(command, service.RoleUser)
	ctx, closer := service.LoadContext(context.Background())
	defer closer()
	stream, err := api.SendPromptToOpenAi(ctx, &api.GPTChanRequest{
		Messages: metaChatMessages.Messages,
	})
	if err != nil {
		return nil, err
	}

	response, err := api.GetFullResponse(stream)
	if err != nil {
		return nil, err
	}

	metaChatMessages.AddMessage(response, service.RoleAssistant)

	return metaChatMessages, nil
}

func LoadChatOnlyAssistant(startPath string) (*service.ChatMessages, error) {
	loadedChat, err := LoadChat(startPath)
	if err != nil {
		return nil, err
	}

	loadedChat.Messages, _ = loadedChat.FilterMessages(service.RoleAssistant)
	if len(loadedChat.Messages) == 0 {
		return nil, errors.New("no assistant message found")
	}

	return loadedChat, nil
}
