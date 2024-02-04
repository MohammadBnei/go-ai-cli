package api_test

import (
	"context"
	"testing"

	"github.com/MohammadBnei/go-openai-cli/api"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestWebSearchAgent(t *testing.T) {
	t.Log("TestWebSearchAgent")
	llm, err := openai.New(openai.WithModel("gpt-4"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created llm")

	executor, err := api.NewWebSearchAgent(llm)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created executor")

	result, err := chains.Run(context.Background(), executor, "Was covid the worst scam of all ?")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}
