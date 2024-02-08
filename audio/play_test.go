package audio_test

import (
	"context"
	"testing"

	"github.com/MohammadBnei/go-ai-cli/audio"
	"github.com/spf13/viper"
)

func TestTextToSpeech(t *testing.T) {
	viper.BindEnv("OPENAI_KEY", "OPENAI_API_KEY")

	err := audio.PlayTextToSpeech(context.Background(), "hello world")
	if err != nil {
		t.Fatal(err)
	}

}
