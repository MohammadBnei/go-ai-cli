package agent

import (
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/scraper"
)

func NewWebSearchAgent(llm llms.Model, urls []string) (*agents.Executor, error) {
	scrap, err := scraper.New()
	if err != nil {
		return nil, err
	}

	t := []tools.Tool{
		scrap,
	}

	executor, err := agents.Initialize(llm, t, agents.ZeroShotReactDescription,
		agents.WithMaxIterations(5),
	)
	if err != nil {
		return nil, err
	}

	return &executor, nil
}
