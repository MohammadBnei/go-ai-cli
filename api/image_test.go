package api_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/config"
)

func TestOllamaImageRead(t *testing.T) {
	// Set the API type to "OLLAMA"
	viper.Set(config.AI_API_TYPE, api.API_OLLAMA)

	// Set the OLLAMA_HOST to your test server URL
	viper.Set(config.AI_OLLAMA_HOST, "http://127.0.0.1:11434")

	viper.Set(config.AI_MODEL_NAME, "llava")

	imageFile, err := os.ReadFile("openclose-inn.jpg")
	if err != nil {
		t.Error(err)
		return
	}

	// Call the function
	resChan, err := api.SendImageToOllama(context.Background(), "Describe precisely this image", imageFile)

	if err != nil {
		t.Error(err)
		return
	}

	for r := range resChan {
		fmt.Print(r)
	}
}

func TestOpenAIImageRead(t *testing.T) {
	// Set the API type to "OLLAMA"
	viper.Set(config.AI_API_TYPE, api.API_OPENAI)

	viper.Set(config.AI_MODEL_NAME, "gpt-4-vision-preview")

	viper.BindEnv(config.AI_OPENAI_KEY, "OPENAI_API_KEY")

	imageFile, err := os.ReadFile("openclose-inn.jpg")
	if err != nil {
		t.Error(err)
		return
	}

	// Call the function
	resChan, err := api.SendImageToOpenAI(context.Background(), "Describe precisely this image", imageFile)

	if err != nil {
		t.Error(err)
		return
	}

	for r := range resChan {
		fmt.Print(r)
	}
}
