package service_test

import (
	"testing"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	viper.Set("get", "config")
	config, _ := service.GetConfig()

	assert.NotEmpty(t, config)
	assert.Equal(t, "config", config["get"])
}

func TestSetConfig(t *testing.T) {
	viper.Set("KEY", "asjdoha_ mcxhfiua")
	err := service.SetConfig("nokey", "value")
	assert.Error(t, err)

	err = service.SetConfig("KEY", "new_key")
	assert.NoError(t, err)

	assert.Equal(t, "new_key", viper.GetString("KEY"))
}
