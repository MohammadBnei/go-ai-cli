package api_test

import (
	"context"
	"testing"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
)

func TestWebSearchAgent(t *testing.T) {
	t.Log("TestWebSearchAgent")
	llm, err := ollama.New(ollama.WithModel("llava"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created llm")

	executor, err := api.NewWebSearchAgent(llm)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created executor")

	result, err := chains.Run(context.Background(), executor, "Was covid the worst scam of all ? Important: Provide your sources.")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Result: " + result)
}
