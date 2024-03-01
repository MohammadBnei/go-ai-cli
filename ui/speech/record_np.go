//go:build !portaudio

package speech

import (
	"context"
	"errors"
	"time"
)

type SpeechConfig struct {
	Duration time.Duration
	Lang     string
}

func SpeechToText(ctx context.Context, aiContext context.Context, config *SpeechConfig) (string, error) {
	return "", errors.New("portaudio not found")
}

func recordAudio(ctx context.Context, filename string, maxDuration time.Duration) error {
	return errors.New("portaudio not found")
}
