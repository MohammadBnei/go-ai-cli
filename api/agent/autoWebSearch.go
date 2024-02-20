package agent

import (
	"context"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/duckduckgo"
	"github.com/tmc/langchaingo/tools/scraper"
)

func NewAutoWebSearchAgent(llm llms.Model) (*agents.Executor, error) {
	ddg, err := duckduckgo.New(10, "en")
	if err != nil {
		return nil, err
	}

	scrap, err := scraper.New()
	if err != nil {
		return nil, err
	}

	t := []tools.Tool{
		NewSearchInputDesigner(),
		ddg,
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

type SearchInputDesigner struct {
}

func NewSearchInputDesigner() tools.Tool {
	return &SearchInputDesigner{}

}

func (s *SearchInputDesigner) Name() string {
	return "SearchInputDesigner"
}

func (s *SearchInputDesigner) Description() string {
	return "SearchInputDesigner is a tool designed to help users design search inputs."
}

func (s *SearchInputDesigner) Call(ctx context.Context, input string) (string, error) {
	llm, err := api.GetLlmModel()
	if err != nil {
		return "", err
	}

	prompt := prompts.NewPromptTemplate(
		"Design a search input, to be used by search engines, for the following text: {{.userInput}}?",
		[]string{"userInput"},
	)

	chain := chains.NewLLMChain(llm, prompt)
	out, err := chains.Run(ctx, chain, input)
	if err != nil {
		return "", err
	}

	return out, nil
}
