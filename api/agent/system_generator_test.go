package agent_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/api/agent"
	"github.com/MohammadBnei/go-ai-cli/config"
)

func TestSystemGenerator(t *testing.T) {
	viper.Set(config.AI_API_TYPE, api.API_OPENAI)
	viper.Set(config.AI_MODEL_NAME, "gpt-4-turbo-preview")
	viper.BindEnv(config.AI_OPENAI_KEY, "OPENAI_API_KEY")

	scg := &agent.UserExchangeChans{
		In:  make(chan string),
		Out: make(chan string),
	}

	go func() {
		for input := range scg.In {
			res := ""
			t.Log("Input: " + input)
			fmt.Scanln(res)
			scg.Out <- res
		}
	}()

	t.Log("TestSystemGenerator")
	executor, err := agent.NewSystemGeneratorExecutor(scg)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Created executor")

	result, err := executor.LLM.GenerateContent(context.Background(), []llms.MessageContent{
		llms.TextParts(schema.ChatMessageTypeSystem, agent.SystemGeneratorPrompt),
		llms.TextParts(schema.ChatMessageTypeHuman, "Create a system prompt for golang code generation."),
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Result: %s", result.Choices[0].Content)
}
