package service

import (
	"errors"
	"sort"
	"time"

	"github.com/MohammadBnei/go-openai-cli/tool"
	"github.com/jinzhu/copier"
	"github.com/pkoukk/tiktoken-go"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"gopkg.in/yaml.v3"
)

type Roles string

const (
	RoleUser      Roles = "user"
	RoleSystem    Roles = "system"
	RoleAssistant Roles = "assistant"
	RoleApp       Roles = "app"
)

type ChatMessage struct {
	Id      int
	Role    Roles  `json:"role"`
	Content string `json:"content"`
	Tokens  int    `json:"tokens"`

	AssociatedMessageId int

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

func (c *ChatMessages) FindById(id int) *ChatMessage {
	_, index, ok := lo.FindIndexOf[ChatMessage](c.Messages, func(item ChatMessage) bool {
		return item.Id == id
	})
	if !ok {
		return nil
	}

	return &c.Messages[index]
}
func (c *ChatMessages) AddMessage(content string, role Roles) (*ChatMessage, error) {

	if role == "" {
		return nil, errors.New("role cannot be empty")
	}

	tokenCount, err := CountTokens(content)
	if err != nil {
		return nil, err
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

	return &msg, nil
}

func (c *ChatMessages) SetAssociatedId(idUser, idAssistant int) error {
	msgUser := c.FindById(idUser)
	if msgUser == nil {
		return errors.New("user message not found")
	}

	msgAssistant := c.FindById(idAssistant)
	if msgAssistant == nil {
		return errors.New("assistant message not found")
	}

	msgUser.AssociatedMessageId = idAssistant
	msgAssistant.AssociatedMessageId = idUser

	return nil
}

func (c *ChatMessages) UpdateMessage(m ChatMessage) error {
	if m.Content == "" {
		return errors.New("content cannot be empty")
	}

	if m.Role == "" {
		return errors.New("role cannot be empty")
	}

	tokenCount, err := CountTokens(m.Content)
	if err != nil {
		return err
	}

	msg := c.FindById(m.Id)
	if msg == nil {
		c.AddMessage(m.Content, m.Role)
		return nil
	}

	m.Tokens = tokenCount

	copier.Copy(msg, m)

	c.RecountTokens()

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

func (c *ChatMessages) ToLangchainMessage() []llms.MessageContent {
	return lo.Map[ChatMessage, llms.MessageContent](c.FilterByOpenAIRoles(), func(item ChatMessage, index int) llms.MessageContent {
		switch item.Role {
		case RoleSystem:
			return llms.TextParts(schema.ChatMessageTypeSystem, item.Content)
		case RoleAssistant:
			return llms.TextParts(schema.ChatMessageTypeAI, item.Content)
		case RoleUser:
			return llms.TextParts(schema.ChatMessageTypeGeneric, item.Content)
		}
		return llms.TextParts(schema.ChatMessageTypeGeneric, item.Content)
	})
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

func (c *ChatMessages) FilterByOpenAIRoles() []ChatMessage {
	return lo.Filter[ChatMessage](c.Messages, func(item ChatMessage, _ int) bool {
		return lo.Contains[Roles]([]Roles{
			openai.ChatMessageRoleUser,
			openai.ChatMessageRoleAssistant,
			openai.ChatMessageRoleSystem,
		}, item.Role)
	})
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
