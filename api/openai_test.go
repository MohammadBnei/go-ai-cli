package api

import (
	"context"
	"io"
	"testing"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSendBasicMessage(t *testing.T) {
	viper.BindEnv("OPENAI_KEY", "OPENAI_API_KEY")

	stream, err := SendPromptToOpenAi(context.Background(), &GPTChanRequest{
		Messages: []service.ChatMessage{{Role: service.RoleUser, Content: "hello world"}},
		Model:    openai.GPT4TurboPreview,
	})

	if err != nil {
		t.Fatal(err)
	}

	content := make([]byte, 20)

	for {
		msg, ok := <-stream
		if !ok || msg.Err == io.EOF {
			break
		}
		if msg.Err != nil {
			t.Fatal(msg.Err)
		}
		content = append(content, msg.Content...)
	}

	assert.NotEmpty(t, content)
}
