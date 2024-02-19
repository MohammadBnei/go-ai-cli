package service

import (
	"context"
	"io"
	"time"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/audio"
	"github.com/bwmarrin/snowflake"
)

type ChatMessage struct { Id      snowflake.ID
	Role    ROLES
	Content string
	Tokens  int
	Type    TYPE

	AssociatedMessageId int64

	Date time.Time

	Audio io.ReadCloser `json:"-,omitempty"`

	Meta Meta
}

type Meta struct {
	ApiType, Model, Agent string
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
