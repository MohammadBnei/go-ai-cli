package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools/scraper"
	"golang.org/x/sync/errgroup"
)

func NewWebSearchAgent(llm llms.Model, urls []string) (func(ctx context.Context, question string) (map[string]any, error), error) {
	var g errgroup.Group

	if len(urls) > 3 {
		return nil, errors.New("too many urls (3 max for now)")
	}

	docs := make([]schema.Document, len(urls))
	for i, url := range urls {
		g.Go(func(url string, i int) func() error {
			return func() error {
				data, err := getHtmlContent(url)
				if err != nil {
					return err
				}

				parsedData, err := parseHtml(context.Background(), data)
				if err != nil {
					return err
				}

				docs[i] = schema.Document{
					PageContent: parsedData,
					Metadata: map[string]any{
						"url": url,
					},
				}

				return nil
			}
		}(url, i))
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	stuffQAChain := chains.LoadStuffQA(llm)

	callFunc := func(ctx context.Context, question string) (map[string]any, error) {
		return chains.Call(context.Background(), stuffQAChain, map[string]any{
			"input_documents": docs,
			"question":        question,
		})
	}

	return callFunc, nil
}

func getHtmlContent(url string) (string, error) {
	scrap, err := scraper.New()
	if err != nil {
		return "", err
	}
	return scrap.Call(context.Background(), url)
}

func parseHtml(ctx context.Context, data string) (string, error) {

	llm, err := api.GetLlmModel()
	if err != nil {
		return "", err
	}

	prompt := prompts.NewPromptTemplate(
		"You will be given a raw html content of a web page. You have to parse the html, and extract the content. Do not modify anything, simply remove the html structure and extract the content. Here is the content : {{.html}}",
		[]string{"html"},
	)

	chain := chains.NewLLMChain(llm, prompt)
	out, err := chains.Run(ctx, chain, data)
	if err != nil {
		return "", err
	}

	fmt.Println("data :"+data, "out :"+out)

	return out, nil
}
