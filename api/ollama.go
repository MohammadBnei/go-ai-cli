package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/spf13/viper"

	"github.com/MohammadBnei/go-ai-cli/config"
)

func SendImageToOllama(ctx context.Context, prompt string, images ...[]byte) (chan string, error) {
	respChan := make(chan string)

	options := make(map[string]string)

	if v := viper.GetFloat64(config.AI_TEMPERATURE); v > 0 {
		options["temperature"] = fmt.Sprintf("%f", v)
	}
	if v := viper.GetInt(config.AI_TOP_K); v > 0 {
		options["top_k"] = fmt.Sprintf("%d", v)
	}
	if v := viper.GetFloat64(config.AI_TOP_P); v > 0 {
		options["top_p"] = fmt.Sprintf("%f", v)
	}

	for _, img := range images {
		contentType := http.DetectContentType(img)
		allowedTypes := []string{"image/jpeg", "image/jpg", "image/png"}
		if !slices.Contains(allowedTypes, contentType) {
			return nil, fmt.Errorf("invalid image type: %s", contentType)
		}
	}

	data := make(map[string]any)
	data["prompt"] = prompt
	data["images"] = images
	data["model"] = viper.GetString(config.AI_MODEL_NAME)

	if len(options) > 0 {
		data["options"] = options
	}

	bts, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(bts)

	request, err := http.NewRequestWithContext(ctx, "POST", viper.GetString(config.AI_OLLAMA_HOST)+"/api/generate", buf)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/x-ndjson")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("request failed with status code: " + fmt.Sprintf("%d", resp.StatusCode))
	}

	dec := json.NewDecoder(resp.Body)
	go func() {
		defer resp.Body.Close()
		defer close(respChan)

		// While the array contains values
		for dec.More() {
			var m GenerateResponse
			// Decode an array value (Message)
			err := dec.Decode(&m)
			if err != nil {
				respChan <- fmt.Sprintf("\nError: %s", err)
				return
			}

			if m.Done {
				return
			}

			respChan <- m.Response
		}
	}()

	return respChan, nil
}

type GenerateResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	SampleCount        int    `json:"sample_count,omitempty"`
	SampleDuration     int64  `json:"sample_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
	Done               bool   `json:"done"`
}

func isBase64Encoded(data []byte) bool {
	// Check if the length of the byte array is divisible by 4
	if len(data)%4 != 0 {
		return false
	}

	// Attempt to decode the byte array
	_, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return false
	}

	return true
}
