package service

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

var messages []openai.ChatCompletionMessage

func SendPrompt(ctx context.Context, text string, output io.Writer) (string, error) {
	c := openai.NewClient(viper.GetString("OPENAI_KEY"))

	s := spinner.New(spinner.CharSets[26], 100*time.Millisecond)
	s.Start()

	addMessage(openai.ChatCompletionMessage{
		Role:    "user",
		Content: text,
	})
	resp, err := c.CreateChatCompletionStream(
		ctx,
		openai.ChatCompletionRequest{
			Model:    viper.GetString("model"),
			Messages: messages,
			Stream:   true,
		},
	)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	fullMsg := ""
	role := ""

	for {
		msg, err := resp.Recv()
		s.Stop()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		output.Write([]byte(msg.Choices[0].Delta.Content))
		fullMsg = strings.Join([]string{fullMsg, msg.Choices[0].Delta.Content}, "")
		if role == "" {
			role = msg.Choices[0].Delta.Role
		}
	}

	addMessage(openai.ChatCompletionMessage{
		Content: fullMsg,
		Role:    role,
	})

	output.Write([]byte("\n"))

	return fullMsg, nil
}

func addMessage(msg openai.ChatCompletionMessage) {
	messages = append(messages, msg)

	if len(messages) > 10 {
		messages = messages[1:]
	}
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
