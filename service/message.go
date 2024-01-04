package service

import (
	"errors"
	"sort"
	"time"

	"github.com/MohammadBnei/go-openai-cli/tool"
	"github.com/pkoukk/tiktoken-go"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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

	ToolCall openai.ToolCall
	Date     time.Time
}

type ChatMessages struct {
	Id          string
	Description string
	Messages    []ChatMessage
	TotalTokens int
}

func NewChatMessages(id string) *ChatMessages {
	return &ChatMessages{
		Id:          id,
		Messages:    []ChatMessage{},
		TotalTokens: 0,
	}
}

func (c *ChatMessages) SetId(id string) *ChatMessages {
	c.Id = id
	return c
}
func (c *ChatMessages) SetDescription(description string) *ChatMessages {
	c.Description = description

	return c
}

func (c *ChatMessages) SaveToFile(filename string) error {
	if filename == "" {
		filename = c.Id
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = tool.SaveToFile(data, viper.GetString("configPath")+"/"+filename+".yaml")
	if err != nil {
		return err
	}

	return nil
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

	sort.SliceStable(c.Messages, func(i, j int) bool {
		return c.Messages[i].Date.Before(c.Messages[j].Date)
	})

	return nil
}

func (c *ChatMessages) DeleteMessage(id int) error {
	message, ok := lo.Find[ChatMessage](c.Messages, func(item ChatMessage) bool {
		return item.Id == id
	})
	if !ok {
		return errors.New("message not found")
	}

	c.TotalTokens -= message.Tokens

	c.Messages = lo.Filter[ChatMessage](c.Messages, func(item ChatMessage, _ int) bool {
		return item.Id != id
	})

	return nil
}

func (c *ChatMessages) ClearMessages() {
	c.Messages = []ChatMessage{}
	c.TotalTokens = 0
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
	messages = lo.Filter[ChatMessage](c.Messages, func(item ChatMessage, _ int) bool {
		return item.Role == role
	})

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.Before(messages[j].Date)
	})

	tokens = lo.Reduce[ChatMessage, int](messages, func(acc int, item ChatMessage, _ int) int {
		tokenCount, _ := CountTokens(item.Content)
		return acc + tokenCount
	}, 0)

	return
}

func (c *ChatMessages) RecountTokens() *ChatMessages {
	c.TotalTokens = 0
	for _, msg := range c.Messages {
		msg.Tokens, _ = CountTokens(msg.Content)
		c.TotalTokens += msg.Tokens
	}
	return c
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
