package service

import (
	"context"
	"errors"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jinzhu/copier"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

var messages []ChatMessage

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`

	Date time.Time
}

func SendPrompt(ctx context.Context, text string, output io.Writer, saveMessage bool) (string, error) {
	c := openai.NewClient(viper.GetString("OPENAI_KEY"))

	s := spinner.New(spinner.CharSets[26], 100*time.Millisecond)
	s.Start()
	defer s.Stop()

	model := viper.GetString("model")

	if model == "" {
		model = openai.GPT3Dot5Turbo
	}
	if saveMessage {

		AddMessage(ChatMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
			Date:    time.Now(),
		})
	}

	chatMessages := []openai.ChatCompletionMessage{}
	err := copier.Copy(&chatMessages, &messages)

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

		output.Write([]byte(msg.Choices[0].Delta.Content))
		fullMsg = strings.Join([]string{fullMsg, msg.Choices[0].Delta.Content}, "")
	}
	if saveMessage {
		AddMessage(ChatMessage{
			Content: fullMsg,
			Role:    openai.ChatMessageRoleAssistant,
			Date:    time.Now(),
		})
	}

	return fullMsg, nil
}

func AddMessage(msg ChatMessage) int {
	messages = append(messages, msg)

	if len(messages) > viper.GetInt("messages-length") {
		messages = messages[1:]
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.Before(messages[j].Date)
	})

	return len(messages)
}

func UpdateMessage(idx int, msg ChatMessage) error {
	if idx >= len(messages) {
		return errors.New("index out of range")
	}
	messages[idx] = msg

	return nil
}

func ClearMessages() {
	messages = []ChatMessage{}
}

func GetMessages() []ChatMessage {
	return messages
}

func SetMessages(msgs []ChatMessage) {
	messages = msgs
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.Before(messages[j].Date)
	})
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
