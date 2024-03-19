package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/samber/lo"
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

func SendImageToOpenAI(ctx context.Context, prompt string, images ...[]byte) (chan string, error) {
	respChan := make(chan string)

	imagesData := []string{}

	for _, img := range images {
		contentType := http.DetectContentType(img)
		allowedTypes := []string{"image/jpeg", "image/jpg", "image/png"}
		if !slices.Contains(allowedTypes, contentType) {
			return nil, fmt.Errorf("invalid image type: %s", contentType)
		}

		imageStr := ""
		switch contentType {
		case "image/jpeg":
			imageStr = "data:image/jpeg;base64,"
		case "image/jpg":
			imageStr = "data:image/jpeg;base64,"
		case "image/png":
			imageStr = "data:image/png;base64,"
		}

		imageStr += base64.StdEncoding.EncodeToString(img)

		imagesData = append(imagesData, imageStr)
	}

	messages := append([]openai.ChatMessagePart{
		{
			Type: openai.ChatMessagePartTypeText,
			Text: prompt,
		},
	}, lo.Map(imagesData, func(imageStr string, _ int) openai.ChatMessagePart {
		return openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL:    imageStr,
				Detail: openai.ImageURLDetailAuto,
			},
		}
	})...)

	c := openai.NewClient(viper.GetString(config.AI_OPENAI_KEY))
	resp, err := c.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: viper.GetString(config.AI_MODEL_NAME),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:         openai.ChatMessageRoleUser,
				MultiContent: messages,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	go func() {
		defer close(respChan)
		for {
			resp, err := resp.Recv()
			if err != nil {
				respChan <- fmt.Sprintf("\nerror: %s", err.Error())
				return
			}
			respChan <- resp.Choices[0].Delta.Content
		}
	}()

	return respChan, nil
}
