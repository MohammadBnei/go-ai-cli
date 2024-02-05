package api

import (
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/metaphor"
)

func NewWebSearchAgent(llm llms.Model) (*agents.Executor, error) {
	exa, err := metaphor.NewClient()
	if err != nil {
		return nil, err
	}

	t := []tools.Tool{
		exa,
	}

	executor, err := agents.Initialize(llm, t, agents.ConversationalReactDescription,
		agents.WithMaxIterations(10),
	)
	if err != nil {
		return nil, err
	}

	return &executor, nil
}
