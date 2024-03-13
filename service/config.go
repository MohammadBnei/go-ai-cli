package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/samber/lo"
)

type ContextHold struct {
	UserChatId snowflake.ID
	Ctx        context.Context
	CancelFn   func()
}

type PromptConfig struct {
	ChatMessages   *ChatMessages
	PreviousPrompt string
	UserPrompt     string
	UpdateChan     chan *ChatMessage
	Contexts       []ContextHold
	*FileService
}

func (pc *PromptConfig) CloseLastContext() error {
	if len(pc.Contexts) == 0 {
		return errors.New("no context")
	}
	pc.Contexts[len(pc.Contexts)-1].CancelFn()
	return nil
}

func (pc *PromptConfig) AddContext(ctx context.Context, cancelFn func()) {
	pc.Contexts = append(pc.Contexts, ContextHold{Ctx: ctx, CancelFn: cancelFn})
}

func (pc *PromptConfig) AddContextWithId(ctx context.Context, cancelFn func(), id int64) {
	pc.Contexts = append(pc.Contexts, ContextHold{Ctx: ctx, CancelFn: cancelFn, UserChatId: snowflake.ParseInt64(id)})
}

func (pc *PromptConfig) CloseContext(ctx context.Context) error {
	ctxHlod, ok := lo.Find(pc.Contexts, func(item ContextHold) bool { return item.Ctx == ctx })
	if !ok {
		return errors.New("context not found")
	}
	ctxHlod.CancelFn()

	pc.Contexts = lo.Filter(pc.Contexts, func(item ContextHold, index int) bool {
		return item.Ctx != ctx
	})

	return nil
}

func (pc *PromptConfig) FindContextWithId(id int64) *ContextHold {
	ctx, _ := lo.Find(pc.Contexts, func(item ContextHold) bool {
		return item.UserChatId != snowflake.ParseInt64(id)
	})
	return &ctx
}

func (pc *PromptConfig) CloseContextById(id int64) error {
	ctx, _, ok := lo.FindLastIndexOf(pc.Contexts, func(item ContextHold) bool { return item.UserChatId == snowflake.ParseInt64(id) })
	if !ok {
		return fmt.Errorf("no context found with id %d, %s", id, pc.Contexts)
	}
	ctx.CancelFn()

	pc.Contexts = lo.Filter(pc.Contexts, func(item ContextHold, index int) bool {
		return item.UserChatId != snowflake.ParseInt64(id)
	})

	return nil
}

func (c ContextHold) String() string {
	return fmt.Sprintf("[%d %s]", c.UserChatId, c.Ctx)
}
