package api_test

import (
	"testing"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetOllamaModelList(t *testing.T) {

	// Set the API type to "OLLAMA"
	viper.Set("API_TYPE", api.API_OLLAMA)

	// Set the OLLAMA_HOST to your test server URL
	viper.Set("OLLAMA_HOST", "127.0.0.1:11434")

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
