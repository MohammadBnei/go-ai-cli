package api

import (
	"fmt"

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

	scrap, err := scraper.New()
	if err != nil {
		return nil, err
	}

	t := []tools.Tool{
		ddg,
		scrap,
	}

	executor, err := agents.Initialize(llm, t, agents.ConversationalReactDescription,
		agents.WithMaxIterations(10),
		agents.WithParserErrorHandler(agents.NewParserErrorHandler(func(s string) string {
			fmt.Println("\n\n" + s)
			return s
		})),
	)
	if err != nil {
		return nil, err
	}

	return &executor, nil
}
