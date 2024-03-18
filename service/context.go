package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/snowflake"
	"github.com/samber/lo"
)

type ContextHolder struct {
	UserChatId snowflake.ID
	Ctx        context.Context
	CancelFn   func()
}

type ContextService struct {
	Contexts []ContextHolder
}

func NewContextService() *ContextService {
	return &ContextService{
		Contexts: []ContextHolder{},
	}
}

func (cs *ContextService) CloseLastContext() error {
	if len(cs.Contexts) == 0 {
		return errors.New("no context")
	}
	cs.Contexts[len(cs.Contexts)-1].CancelFn()
	return nil
}

func (cs *ContextService) AddContext(ctx context.Context, cancelFn func()) {
	cs.Contexts = append(cs.Contexts, ContextHolder{Ctx: ctx, CancelFn: cancelFn})
}

func (cs *ContextService) AddContextWithId(ctx context.Context, cancelFn func(), id int64) {
	cs.Contexts = append(cs.Contexts, ContextHolder{Ctx: ctx, CancelFn: cancelFn, UserChatId: snowflake.ParseInt64(id)})
}

func (cs *ContextService) CloseContext(ctx context.Context) error {
	ctxHlod, ok := lo.Find(cs.Contexts, func(item ContextHolder) bool { return item.Ctx == ctx })
	if !ok {
		return errors.New("context not found")
	}
	ctxHlod.CancelFn()

	cs.Contexts = lo.Filter(cs.Contexts, func(item ContextHolder, index int) bool {
		return item.Ctx != ctx
	})

	return nil
}

func (cs *ContextService) FindContextWithId(id int64) *ContextHolder {
	ctx, _ := lo.Find(cs.Contexts, func(item ContextHolder) bool {
		return item.UserChatId != snowflake.ParseInt64(id)
	})
	return &ctx
}

func (cs *ContextService) CloseContextById(id int64) error {
	ctx, _, ok := lo.FindLastIndexOf(cs.Contexts, func(item ContextHolder) bool { return item.UserChatId == snowflake.ParseInt64(id) })
	if !ok {
		return fmt.Errorf("no context found with id %d, %s", id, cs.Contexts)
	}
	ctx.CancelFn()

	cs.Contexts = lo.Filter(cs.Contexts, func(item ContextHolder, index int) bool {
		return item.UserChatId != snowflake.ParseInt64(id)
	})

	return nil
}

func (cs *ContextService) Length() int {
	return len(cs.Contexts)
}
func (c ContextHolder) String() string {
	return fmt.Sprintf("[%d %s]", c.UserChatId, c.Ctx)
}
