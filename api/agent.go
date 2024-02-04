package api

import (
	"context"
	"time"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/duckduckgo"
	"github.com/tmc/langchaingo/tools/scraper"
)

func NewWebSearchAgent(llm llms.Model) (*agents.Executor, error) {
	ddg, err := duckduckgo.New(10, "en")
	if err != nil {
		return nil, err
	}

	scraper, err := scraper.New()
	if err != nil {
		return nil, err
	}

	t := []tools.Tool{
		ddg,
		scraper,
		NewGetTime(),
	}

	executor, err := agents.Initialize(llm, t, agents.ConversationalReactDescription, agents.WithMaxIterations(3))
	if err != nil {
		return nil, err
	}

	return &executor, nil
}

type GetTime struct {
}

func NewGetTime() tools.Tool {
	return &GetTime{}
}

func (t *GetTime) Name() string {
	return "Get Current Time"
}

func (t *GetTime) Description() string {
	return "Returns the current time in the following format : YYYY-MM-DD HH:MM:SS"
}

func (t *GetTime) Call(ctx context.Context, input string) (string, error) {

	return time.Now().Format("2006-01-22 15:04:05"), nil

}
