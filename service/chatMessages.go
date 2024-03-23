package service

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

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

type MessagesService struct {
	Id          string
	Description string
	Messages    CMessages
	TotalTokens int

	node *snowflake.Node
}

func NewChatMessages(id string) *MessagesService {
	node, _ := snowflake.NewNode(1)
	return &MessagesService{
		Id:          id,
		Messages:    make(CMessages, 0),
		TotalTokens: 0,

		node: node,
	}
}

func (c *MessagesService) SetId(id string) *MessagesService {
	c.Id = id

	return c
}
func (c *MessagesService) SetDescription(description string) *MessagesService {
	c.Description = description

	return c
}

func (c *MessagesService) SaveToFile(filename string) error {
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

func (c *MessagesService) SaveChatInModelfileFormat(filename string) error {
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

func (c *MessagesService) LoadFromFile(filename string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	marshalledC := &MessagesService{}

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

func (c *MessagesService) FindById(id int64) *ChatMessage {
	return c.Messages.FindById(id)
}

func (c *MessagesService) FindByOrder(order uint) *ChatMessage {
	return c.Messages.FindByOrder(order)
}

func (c *MessagesService) FindMessageByContent(content string) *ChatMessage {
	return c.Messages.FindMessageByContent(content)
}

var ErrAlreadyExist = errors.New("already exists")

func (c *MessagesService) AddMessage(content string, role ROLES) (*ChatMessage, error) {
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

	msg := c.Messages.NewMessage(c.node.Generate(), content, role)

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
func (c *MessagesService) AddMessageFromFile(filename string) (*ChatMessage, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return c.AddMessage(fmt.Sprintf("(Filename : %s)\n\n%s", filename, content), RoleUser)
}

func (c *MessagesService) SetAssociatedId(idUser, idAssistant int64) error {
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

// AppendToMessageContent appends the given appendix to the content of the message with the specified ID.
//
// Parameters:
// - id: The ID of the message to append to.
// - appendix: The string to append to the message content.
//
// Returns:
// - ChatMessage: The updated chat message with the appended content.
// - error: An error if the message with the specified ID is not found.
func (c *MessagesService) AppendToMessageContent(id int64, appendix string) (ChatMessage, error) {
	msg := c.FindById(id)
	if msg == nil {
		return ChatMessage{}, errors.New("message not found")
	}

	msg.Content += appendix

	c.RecountTokens()

	return *msg, nil
}

func (c *MessagesService) UpdateMessage(m ChatMessage) error {
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

func (c *MessagesService) DeleteMessage(id int64) error {
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

func (c *MessagesService) ToLangchainMessage() []llms.MessageContent {
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

func (c *MessagesService) ClearMessages() {
	c.Messages = []ChatMessage{}
	c.TotalTokens = 0
}

func (c *MessagesService) LastMessage(role ROLES) *ChatMessage {
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

func (c *MessagesService) FilterMessages(role ROLES) (messages []ChatMessage, tokens int) {
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

func (c *MessagesService) FilterByOpenAIRoles() []ChatMessage {
	return lo.Filter(c.Messages, func(item ChatMessage, _ int) bool {
		return lo.Contains([]ROLES{
			openai.ChatMessageRoleUser,
			openai.ChatMessageRoleAssistant,
			openai.ChatMessageRoleSystem,
		}, item.Role)
	})
}

func (c *MessagesService) RecountTokens() *MessagesService {
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

func (c *MessagesService) SetMessagesOrder() *MessagesService {
	c.Messages = lo.Map(c.Messages, func(item ChatMessage, order int) ChatMessage {
		item.Order = uint(order + 1)
		return item
	})

	return c
}
