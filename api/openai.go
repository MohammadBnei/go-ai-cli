package api

import (
	"context"
	"io"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

func SpeechToText(ctx context.Context, filename string, lang string) (string, error) {
	c := openai.NewClient(viper.GetString("OPENAI_KEY"))

	if lang == "" {
		lang = "en"
	}

	response, err := c.CreateTranscription(ctx, openai.AudioRequest{
		Model:    openai.Whisper1,
		Format:   openai.AudioResponseFormatJSON,
		FilePath: filename,
		Language: lang,
	})
	if err != nil {
		return "", err
	}

	return response.Text, nil
}

func TextToSpeech(ctx context.Context, content string) (io.ReadCloser, error) {
	c := openai.NewClient(viper.GetString("OPENAI_KEY"))

	response, err := c.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		ResponseFormat: openai.SpeechResponseFormatMp3,
		Input:          content,
		Voice:          openai.VoiceNova,
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}
