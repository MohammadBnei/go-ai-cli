package service

import (
	"errors"
	"sort"
	"time"

	"github.com/pkoukk/tiktoken-go"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

type Roles string

const (
	RoleUser      Roles = "user"
	RoleSystem    Roles = "system"
	RoleAssistant Roles = "assistant"
)

type ChatMessage struct {
	Id      int
	Role    Roles  `json:"role"`
	Content string `json:"content"`
	Tokens  int    `json:"tokens"`

	Date time.Time
}

type ChatMessages struct {
	id          string
	Messages    []ChatMessage
	TotalTokens int
}

func NewChatMessages(id string) *ChatMessages {
	return &ChatMessages{
		id:          id,
		Messages:    []ChatMessage{},
		TotalTokens: 0,
	}
}

func (c *ChatMessages) AddMessage(content string, role Roles) error {
	if content == "" {
		return errors.New("content cannot be empty")
	}

	if role == "" {
		return errors.New("role cannot be empty")
	}

	tokenCount, err := CountTokens(content)
	if err != nil {
		return err
	}

	msg := ChatMessage{
		Id:      len(c.Messages),
		Role:    role,
		Content: content,
		Date:    time.Now(),
	}

	msg.Tokens = tokenCount
	c.Messages = append(c.Messages, msg)

	c.TotalTokens += tokenCount

	sort.Slice(c.Messages, func(i, j int) bool {
		return c.Messages[i].Date.Before(c.Messages[j].Date)
	})

	return nil
}

func (c *ChatMessages) DeleteMessage(id int) error {
	if _, ok := lo.Find[ChatMessage](c.Messages, func(item ChatMessage) bool {
		return item.Id == id
	}); !ok {
		return errors.New("message not found")
	}

	c.Messages = lo.Filter[ChatMessage](c.Messages, func(item ChatMessage, _ int) bool {
		return item.Id != id
	})
	c.TotalTokens -= c.Messages[id].Tokens

	return nil
}

func (c *ChatMessages) ClearMessages() {
	c.Messages = []ChatMessage{}
}

func (c *ChatMessages) LastMessage(role *Roles) *ChatMessage {
	messages := c.Messages
	if role != nil {
		messages = lo.Filter[ChatMessage](c.Messages, func(item ChatMessage, _ int) bool {
			return item.Role == *role
		})
	}
	if len(messages) == 0 {
		return nil
	}

	return &messages[len(messages)-1]
}

func (c *ChatMessages) FilterMessages(role Roles) (messages []ChatMessage, tokens int) {
	messages = lo.Filter[ChatMessage](messages, func(item ChatMessage, _ int) bool {
		return item.Role == role
	})

	tokens = lo.Reduce[ChatMessage, int](messages, func(acc int, item ChatMessage, _ int) int {
		tokenCount, _ := CountTokens(item.Content)
		return acc + tokenCount
	}, 0)

	return
}

func CountTokens(content string) (int, error) {
	model := viper.GetString("model")

	if model == "" {
		model = openai.GPT3Dot5Turbo
	}
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return 0, err
	}

	return len(tkm.Encode(content, nil, nil)), nil
}
