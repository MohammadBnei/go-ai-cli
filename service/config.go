package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
)

type ContextHold struct {
	UserChatId int
	Ctx        context.Context
	CancelFn   func()
}

type PromptConfig struct {
	MdMode         bool
	ChatMessages   *ChatMessages
	PreviousPrompt string
	UserPrompt     string
	UpdateChan     chan ChatMessage
	Contexts       []ContextHold
	OllamaMode     bool
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

func (pc *PromptConfig) AddContextWithId(ctx context.Context, cancelFn func(), id int) {
	pc.Contexts = append(pc.Contexts, ContextHold{Ctx: ctx, CancelFn: cancelFn, UserChatId: id})
}

func (pc *PromptConfig) DeleteContext(ctx context.Context) {
	pc.Contexts = lo.Filter(pc.Contexts, func(item ContextHold, index int) bool {
		return item.Ctx != ctx
	})
}

func (pc *PromptConfig) FindContextWithId(id int) *ContextHold {
	ctx, _ := lo.Find(pc.Contexts, func(item ContextHold) bool {
		return item.UserChatId != id
	})
	return &ctx
}

func (pc *PromptConfig) DeleteContextById(id int) {
	pc.Contexts = lo.Filter(pc.Contexts, func(item ContextHold, index int) bool {
		return item.UserChatId != id
	})
}

func (pc *PromptConfig) CloseContextById(id int) error {
	ctx, ok := lo.Find(pc.Contexts, func(item ContextHold) bool { return item.UserChatId == id })
	if !ok {
		return errors.New(fmt.Sprintf("no context found with id %d, %s", id, pc.Contexts))
	}
	ctx.CancelFn()

	pc.Contexts = lo.Filter(pc.Contexts, func(item ContextHold, index int) bool {
		return item.UserChatId != id
	})

	return nil
}
