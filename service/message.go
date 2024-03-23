package service

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/samber/lo"
	"github.com/spf13/viper"

	"github.com/MohammadBnei/go-ai-cli/config"
)

type ChatMessage struct {
	Id      snowflake.ID
	Role    ROLES
	Content string
	Tokens  int
	Type    TYPE

	Order uint

	AssociatedMessageId int64

	Date time.Time

	AudioFileId string `json:"-"`

	Meta Meta
}

type Meta struct {
	ApiType, Model, Agent string
}

type CMessages []ChatMessage

func (m CMessages) NewMessage(id snowflake.ID, content string, role ROLES) *ChatMessage {
	if role == "" {
		role = RoleUser
	}
	
	msg := &ChatMessage{
		Id:                  id,
		Role:                role,
		Content:             content,
		Date:                time.Now(),
		Type:                TypeUser,
		AssociatedMessageId: -1,
		Meta: Meta{
			ApiType: viper.GetString(config.AI_API_TYPE),
			Model:   viper.GetString(config.AI_MODEL_NAME),
		},
		Order:   uint(len(m) + 1),
	}

	return msg
}

func (m CMessages) FindById(id int64) *ChatMessage {
	_, index, ok := lo.FindIndexOf(m, func(item ChatMessage) bool {
		return item.Id == snowflake.ParseInt64(id)
	})
	if !ok {
		return nil
	}

	return &m[index]
}

func (m CMessages) FindByOrder(order uint) *ChatMessage {
	_, index, ok := lo.FindIndexOf(m, func(item ChatMessage) bool {
		return item.Order == order
	})
	if !ok {
		return nil
	}

	return &m[index]
}

func (m CMessages) FindMessageByContent(content string) *ChatMessage {
	_, index, ok := lo.FindIndexOf(m, func(item ChatMessage) bool {
		return item.Content == content
	})

	if !ok {
		return nil
	}

	return &m[index]
}

func (c *ChatMessage) AsTypeFile() *ChatMessage {
	c.Type = TypeFile
	return c
}

func (c *ChatMessage) SetAudioFileId(id string) *ChatMessage {
	c.AudioFileId = id
	return c
}
