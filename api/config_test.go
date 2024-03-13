package api_test

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/MohammadBnei/go-ai-cli/service/godcontext"
)

func TestMain(m *testing.M) {
	godcontext.GodContext = context.Background()
	goleak.VerifyTestMain(m)
}

func TestGetOllamaModelList(t *testing.T) {

	// Set the API type to "OLLAMA"
	viper.Set(config.AI_API_TYPE, api.API_OLLAMA)

	// Set the OLLAMA_HOST to your test server URL
	viper.Set(config.AI_OLLAMA_HOST, "http://127.0.0.1:11434")

	// Call the function
	models, err := api.GetOllamaModelList()

	if err != nil {
		t.Error(err)
	}

	// Assure the returned models is not nil
	assert.NotNil(t, models)

	// Do a check for empty slice if you wish
	assert.NotEqual(t, 0, len(models))
}
