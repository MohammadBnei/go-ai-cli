package service

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/jinzhu/copier"
	"github.com/pkoukk/tiktoken-go"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"gopkg.in/yaml.v3"

	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/MohammadBnei/go-ai-cli/tool"
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

type ChatMessages struct {
	Id          string
	Description string
	Messages    []ChatMessage
	TotalTokens int

	node *snowflake.Node
}

func NewChatMessages(id string) *ChatMessages {
	node, _ := snowflake.NewNode(1)
	return &ChatMessages{
		Id:          id,
		Messages:    []ChatMessage{},
		TotalTokens: 0,

		node: node,
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

	copyOfC := *c
	copyOfC.Messages = lo.Map(copyOfC.Messages, func(item ChatMessage, _ int) ChatMessage {
		item.AudioFileId = ""
		return item
	})

	data, err := yaml.Marshal(copyOfC)
	if err != nil {
		return err
	}

	return tool.SaveToFile(data, filename, false)
}

func (c *ChatMessages) SaveChatInModelfileFormat(filename string) error {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("FROM %s\n\n", viper.GetString(config.AI_MODEL_NAME)))

	for _, m := range c.Messages {
		switch m.Role {
		case RoleUser:
			builder.WriteString("MESSAGE user ")
		case RoleAssistant:
			builder.WriteString("MESSAGE assistant ")
		default:
			continue
		}

		content := m.Content
		content = strings.ReplaceAll(m.Content, "\\\"", "\"")
		content = strings.ReplaceAll(m.Content, "\"", "\\\"")

		builder.WriteString(fmt.Sprintf("\"\"\"\n%s\"\"\"\n\n", content))
	}

	return tool.SaveToFile([]byte(builder.String()), filename, false)
}

func (c *ChatMessages) LoadFromFile(filename string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	marshalledC := &ChatMessages{}

	if err := yaml.Unmarshal(content, marshalledC); err != nil {
		return err
	}

	c.Messages = marshalledC.Messages

	c.RecountTokens()
	c.Id = marshalledC.Id
	c.Description = marshalledC.Description

	c.SetMessagesOrder()

	lastMessage := c.LastMessage(RoleAssistant)
	if lastMessage != nil && lastMessage.Meta.ApiType != "" && lastMessage.Meta.Model != "" {
		viper.Set(config.AI_API_TYPE, lastMessage.Meta.ApiType)
		viper.Set(config.AI_MODEL_NAME, lastMessage.Meta.Model)
	}

	return nil
}

func (c *ChatMessages) FindById(id int64) *ChatMessage {
	_, index, ok := lo.FindIndexOf(c.Messages, func(item ChatMessage) bool {
		return item.Id == snowflake.ParseInt64(id)
	})
	if !ok {
		return nil
	}

	return &c.Messages[index]
}

func (c *ChatMessages) FindByOrder(order uint) *ChatMessage {
	_, index, ok := lo.FindIndexOf(c.Messages, func(item ChatMessage) bool {
		return item.Order == order
	})
	if !ok {
		return nil
	}

	return &c.Messages[index]
}

var ErrNotFound = errors.New("not found")

func (c *ChatMessages) FindMessageByContent(content string) (*ChatMessage, error) {
	exists, ok := lo.Find(c.Messages, func(item ChatMessage) bool {
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

	if exists, ok := lo.Find(c.Messages, func(item ChatMessage) bool {
		return item.Content == content && item.Role == role && item.Role != RoleUser
	}); ok && content != "" {
		return &exists, ErrAlreadyExist
	}

	msg := &ChatMessage{
		Id:                  c.node.Generate(),
		Role:                role,
		Content:             content,
		Date:                time.Now(),
		Type:                TypeUser,
		AssociatedMessageId: -1,
		Meta: Meta{
			ApiType: viper.GetString(config.AI_API_TYPE),
			Model:   viper.GetString(config.AI_MODEL_NAME),
		},
		Order: uint(len(c.Messages)) + 1,
	}

	msg.Tokens = tokenCount

	c.Messages = append(c.Messages, *msg)

	c.TotalTokens += tokenCount

	sort.SliceStable(c.Messages, func(i, j int) bool {
		return c.Messages[i].Date.Before(c.Messages[j].Date)
	})

	c.SetMessagesOrder()

	return msg, nil
}

// AddMessageFromFile reads the content of a file using os.ReadFile, then adds a new message with the file content and filename to the ChatMessages using the AddMessage method.
// If there's an error reading the file, it returns the error.
func (c *ChatMessages) AddMessageFromFile(filename string) (*ChatMessage, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return c.AddMessage(fmt.Sprintf("(Filename : %s)\n\n%s", filename, content), RoleUser)
}

func (c *ChatMessages) SetAssociatedId(idUser, idAssistant int64) error {
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

	msg := c.FindById(m.Id.Int64())
	if msg == nil {
		c.AddMessage(m.Content, m.Role)
		return nil
	}

	m.Tokens = tokenCount

	copier.Copy(msg, m)

	c.RecountTokens()

	return nil
}

func (c *ChatMessages) DeleteMessage(id int64) error {
	message, ok := lo.Find[ChatMessage](c.Messages, func(item ChatMessage) bool {
		return item.Id == snowflake.ParseInt64(id)
	})
	if !ok {
		return errors.New("message not found")
	}

	c.TotalTokens -= message.Tokens

	c.Messages = lo.Filter[ChatMessage](c.Messages, func(item ChatMessage, _ int) bool {
		return item.Id != snowflake.ParseInt64(id)
	})

	c.SetMessagesOrder()

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

func (c *ChatMessages) LastMessage(role ROLES) *ChatMessage {
	messages := c.Messages
	if role != "" {
		messages = lo.Filter(c.Messages, func(item ChatMessage, _ int) bool {
			return item.Role == role
		})
	}
	if len(messages) == 0 {
		return nil
	}

	return &messages[len(messages)-1]
}

func (c *ChatMessages) FilterMessages(role ROLES) (messages []ChatMessage, tokens int) {
	messages = lo.Filter(c.Messages, func(item ChatMessage, _ int) bool {
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
	return lo.Filter(c.Messages, func(item ChatMessage, _ int) bool {
		return lo.Contains([]ROLES{
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

func (c *ChatMessages) SetMessagesOrder() *ChatMessages {
	c.Messages = lo.Map(c.Messages, func(item ChatMessage, order int) ChatMessage {
		item.Order = uint(order + 1)
		return item
	})

	return c
}
