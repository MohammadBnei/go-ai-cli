package agent_test

import (
	"context"
	"testing"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/api/agent"
	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms/ollama"
)

func TestWebSearch(t *testing.T) {
	viper.Set(config.AI_API_TYPE, api.API_OLLAMA)
	viper.Set(config.AI_MODEL_NAME, "llama2")
	viper.Set(config.AI_OLLAMA_HOST, "http://127.0.0.1:11434")

	t.Log("TestWebSearch")
	llm, err := ollama.New(ollama.WithModel("llama2"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created llm")

	executor, err := agent.NewWebSearchAgent(llm, []string{"https://platform.openai.com/docs/guides/prompt-engineering/strategy-write-clear-instructions"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created executor")

	result, err := executor(context.Background(), "design a system prompt for a golang coder that explains code and writes beautiful code")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Result: %s", result)
}
