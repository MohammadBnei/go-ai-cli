package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"github.com/MohammadBnei/go-ai-cli/config"
)

func SpeechToText(ctx context.Context, filename string, lang string) (string, error) {
	c := openai.NewClient(viper.GetString(config.AI_OPENAI_KEY))

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

func GenerateImage(ctx context.Context, prompt string, size string) ([]byte, error) {
	c := openai.NewClient(viper.GetString(config.AI_OPENAI_KEY))
	model := viper.GetString(config.AI_OPENAI_IMAGE_MODEL)
	if model == "" {
		model = "dall-e-3"
	}

	resp, err := c.CreateImage(ctx, openai.ImageRequest{
		Prompt: prompt,
		User:   "user",
		Model:  model,

		Size:           size,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	})
	if err != nil {
		return nil, err

	}

	b, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func TextToSpeech(ctx context.Context, content string) (io.ReadCloser, error) {
	c := openai.NewClient(viper.GetString(config.AI_OPENAI_KEY))

	parts := []string{""}
	if len(content) >= 4096 {
		splitted := strings.SplitAfter(content, "\n")
		for _, s := range splitted {
			if len(parts[len(parts)-1])+len(s) >= 4096 {
				parts = append(parts, "")
			}
			parts[len(parts)-1] = parts[len(parts)-1] + s
		}
	} else {
		s, err := c.CreateSpeech(ctx, openai.CreateSpeechRequest{
			Model:          openai.TTSModel1,
			ResponseFormat: openai.SpeechResponseFormatMp3,
			Input:          content,
			Voice:          openai.VoiceNova,
		})
		if err != nil {
			return nil, err
		}
		return s, nil
	}

	var g errgroup.Group

	responses := make(map[int]io.ReadCloser)

	for index, p := range parts {
		currentIndex := index
		currentTextPart := p
		g.Go(func() error {

			s, err := c.CreateSpeech(ctx, openai.CreateSpeechRequest{
				Model:          openai.TTSModel1,
				ResponseFormat: openai.SpeechResponseFormatMp3,
				Input:          currentTextPart,
				Voice:          openai.VoiceNova,
			})
			if err != nil {
				return err
			}
			responses[currentIndex] = s

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	response := io.NopCloser(io.MultiReader(bytes.NewReader([]byte(""))))

	for i := range len(parts) {
		originalData, err := io.ReadAll(response)
		if err != nil {
			return nil, err
		}
		newData, err := io.ReadAll(responses[i])
		if err != nil {
			return nil, err
		}

		updatedData := append(originalData, newData...)
		response = io.NopCloser(io.MultiReader(bytes.NewReader(updatedData)))
	}

	return response, nil
}
