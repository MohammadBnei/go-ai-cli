package agent_test

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms/ollama"
	"go.uber.org/goleak"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/api/agent"
	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/MohammadBnei/go-ai-cli/service/godcontext"
)

func TestMain(m *testing.M) {
	godcontext.GodContext = context.Background()
	goleak.VerifyTestMain(m)
}

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

	executor, err := agent.NewWebSearchAgent(llm, []string{"https://www.francesoir.fr/portraits/les-numeros-d-illusionnistes-du-commissaire-censeur-thierry-breton"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created executor")

	result, err := executor(context.Background(), "Resume the provided articles")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Result: %s", result)
}
