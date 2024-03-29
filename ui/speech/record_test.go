package speech_test

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/MohammadBnei/go-ai-cli/service/godcontext"
	"github.com/MohammadBnei/go-ai-cli/ui/speech"
)

func TestMain(m *testing.M) {
	godcontext.GodContext = context.Background()
	goleak.VerifyTestMain(m)

}

func TestRecordMaxDuration(t *testing.T) {
	viper.BindEnv("OPENAI_KEY", "OPENAI_API_KEY")
	res, err := speech.SpeechToText(context.Background(), context.Background(), &speech.SpeechConfig{Duration: 10 * time.Second, Lang: "en"})

	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, res)

	t.Log(res)
}

func TestRecordCancel(t *testing.T) {
	viper.BindEnv("OPENAI_KEY", "OPENAI_API_KEY")
	ctx, cancelFn := context.WithCancel(context.Background())

	go func() {
		time.Sleep(5 * time.Second)
		cancelFn()
	}()
	res, err := speech.SpeechToText(ctx, context.Background(), &speech.SpeechConfig{Duration: 20 * time.Second, Lang: "en"})

	if err != nil {
		t.Error(err)
	}

	assert.NotEmpty(t, res)

	t.Log(res)
}
