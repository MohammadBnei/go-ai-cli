package service

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jinzhu/copier"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

type SendPromptConfig struct {
	ChatMessages *ChatMessages
	Output       io.Writer
}

func SendPrompt(ctx context.Context, req *SendPromptConfig) (string, error) {
	c := openai.NewClient(viper.GetString("OPENAI_KEY"))

	s := spinner.New(spinner.CharSets[26], 100*time.Millisecond)
	s.Start()
	defer s.Stop()

	model := viper.GetString("model")

	if model == "" {
		model = openai.GPT3Dot5Turbo
	}

	chatMessages := []openai.ChatCompletionMessage{}
	err := copier.Copy(&chatMessages, req.ChatMessages.Messages)

	if err != nil {
		return "", err
	}
	resp, err := c.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: chatMessages,
			Stream:   true,
		},
	)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	fullMsg := ""

	for {
		msg, err := resp.Recv()
		s.Stop()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		req.Output.Write([]byte(msg.Choices[0].Delta.Content))
		fullMsg = strings.Join([]string{fullMsg, msg.Choices[0].Delta.Content}, "")
	}

	return fullMsg, nil
}

func GetModelList() ([]string, error) {
	c := openai.NewClient(viper.GetString("OPENAI_KEY"))
	models, err := c.ListModels(context.Background())
	if err != nil {
		return nil, err
	}

	modelsList := []string{}
	for _, model := range models.Models {
		modelsList = append(modelsList, model.ID)
	}

	return modelsList, nil
}
