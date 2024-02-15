package service

import (
	"context"
	"errors"
	"io"
	"os"
	"sort"
	"time"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/audio"
	"github.com/MohammadBnei/go-ai-cli/tool"
	"github.com/jinzhu/copier"
	"github.com/pkoukk/tiktoken-go"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"gopkg.in/yaml.v3"
)

type ROLES string
type TYPE string

const (
	RoleUser      ROLES = "user"
	RoleSystem    ROLES = "system"
	RoleAssistant ROLES = "assistant"
	RoleApp       ROLES = "app"

	TypeFile TYPE = "file"
	TypeUser TYPE = "user"
)

type ChatMessage struct {
	Id      int
	Role    ROLES  `json:"role"`
	Content string `json:"content"`
	Tokens  int    `json:"tokens"`
	Type    TYPE

	AssociatedMessageId int

	ToolCall openai.ToolCall
	Date     time.Time

	Audio io.ReadCloser

	Meta Meta
}

type Meta struct {
	ApiType, Model string
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
		return errors.New("filename cannot be empty")
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = tool.SaveToFile(data, filename, false)
	if err != nil {
		return err
	}

	return nil
}

func (c *ChatMessages) LoadFromFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	marshalledC := &ChatMessages{}

	if err := yaml.Unmarshal(content, marshalledC); err != nil {
		return err
	}

	if err = copier.Copy(c, marshalledC); err != nil {
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

var ErrNotFound = errors.New("not found")

func (c *ChatMessages) FindMessageByContent(content string) (*ChatMessage, error) {
	exists, ok := lo.Find[ChatMessage](c.Messages, func(item ChatMessage) bool {
		return item.Content == content
	})

	if !ok {
		return nil, ErrNotFound
	}

	return &exists, nil
}

var ErrAlreadyExist = errors.New("already exists")

func (c *ChatMessages) AddMessage(content string, role ROLES) (*ChatMessage, error) {
	if role == "" {
		return nil, errors.New("role cannot be empty")
	}

	tokenCount, err := CountTokens(content)
	if err != nil {
		return nil, err
	}

	if exists, ok := lo.Find[ChatMessage](c.Messages, func(item ChatMessage) bool {
		return item.Content == content && item.Role == role && item.Role != RoleUser
	}); ok && content != "" {
		return &exists, ErrAlreadyExist
	}

	msg := ChatMessage{
		Id:                  len(c.Messages),
		Role:                role,
		Content:             content,
		Date:                time.Now(),
		Type:                TypeUser,
		AssociatedMessageId: -1,
		Meta: Meta{
			ApiType: viper.GetString("API_TYPE"),
			Model:   viper.GetString("model"),
		},
	}

	msg.Tokens = tokenCount
	c.Messages = append(c.Messages, msg)

	c.TotalTokens += tokenCount

	sort.SliceStable(c.Messages, func(i, j int) bool {
		return c.Messages[i].Date.Before(c.Messages[j].Date)
	})

	return &msg, nil
}

func (c *ChatMessages) AddMessageFromFile(filename string) (*ChatMessage, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return c.AddMessage(string(content), RoleUser)
}

func (c *ChatMessage) AsTypeFile() *ChatMessage {
	c.Type = TypeFile
	return c
}

func (c *ChatMessage) FetchAudio(ctx context.Context) error {
	data, err := api.TextToSpeech(ctx, c.Content)
	if err != nil {
		return err
	}

	c.Audio = data
	return nil
}

func (c *ChatMessage) SetAudio(data io.ReadCloser) *ChatMessage {
	c.Audio = data
	return c
}

func (c *ChatMessage) PlayAudio(ctx context.Context) error {
	return audio.PlaySound(ctx, c.Audio)
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

func (c *ChatMessages) LastMessage(role *ROLES) *ChatMessage {
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

func (c *ChatMessages) FilterMessages(role ROLES) (messages []ChatMessage, tokens int) {
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
		return lo.Contains[ROLES]([]ROLES{
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
	tkm, err := tiktoken.EncodingForModel(openai.GPT4)
	if err != nil {
		return 0, err
	}

	return len(tkm.Encode(content, nil, nil)), nil
}
