package agent

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
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
				data, err := fetchHTML(url)
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
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT  10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36"),
	)
	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	scrap, err := scraper.New()
	if err != nil {
		return "", err
	}
	return scrap.Call(context.Background(), url)
}

func fetchHTML(url string) (string, error) {
	// Initialize a new browser context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Navigate to the URL and fetch the rendered HTML
	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return "", err
	}

	return htmlContent, nil
}
func parseHtml(ctx context.Context, data string) (string, error) {

	llm, err := api.GetLlmModel()
	if err != nil {
		return "", err
	}

	prompt := prompts.NewPromptTemplate(
		`You are an HTML parser. You job is to extract meaningful information from the HTML, removing the html code while keeping the meaning. 
		If you see a XSS security, a "is human" verification, it means that the site is not accessible to you. If you see something like this, this is the output you must produce : site not accessible : [SHOW ERROR HERE].		Here is the content : {{.html}}`,
		[]string{"html"},
	)

	chain := chains.NewLLMChain(llm, prompt)
	out, err := chains.Run(ctx, chain, data)
	if err != nil {
	}

	fmt.Println("data :"+data, "\nout :"+out)

	return out, nil
}
