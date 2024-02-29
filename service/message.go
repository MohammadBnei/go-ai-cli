package service

import (
	"time"

	"github.com/bwmarrin/snowflake"
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

func (c *ChatMessage) AsTypeFile() *ChatMessage {
	c.Type = TypeFile
	return c
}

func (c *ChatMessage) SetAudioFileId(id string) *ChatMessage {
	c.AudioFileId = id
	return c
}
