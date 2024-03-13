package agent_test

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/api/agent"
	"github.com/MohammadBnei/go-ai-cli/config"
)

func TestWebSearchAgent(t *testing.T) {
	viper.Set(config.AI_API_TYPE, api.API_OLLAMA)
	viper.Set(config.AI_MODEL_NAME, "based-dolphin-mistral")
	viper.Set(config.AI_OLLAMA_HOST, "http://127.0.0.1:11434")

	t.Log("TestWebSearchAgent")
	llm, err := openai.New(openai.WithModel("gpt-4-turbo-preview"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created llm")

	executor, err := agent.NewAutoWebSearchAgent(llm)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created executor")

	result, err := chains.Run(context.Background(), executor, "I want you to design a system prompt for a gpt-4 model. The system prompt will make the ai model tell persian mythology. I want the model to respond with true facts, not to invent anything. The model will respond in the format of an engaging story, presenting itself as a character inside the story.")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Result: " + result)
}
