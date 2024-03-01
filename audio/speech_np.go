//go:build !portaudio

package audio

import (
	"context"
	"errors"
	"time"
)

type SpeechConfig struct {
	MaxMinutes time.Duration
	Lang       string
	Detect     bool
}

func SpeechToText(ctx context.Context, config *SpeechConfig) (string, error) {
	return nil, errors.New("portaudio not found")
}

func RecordAudioToFile(ctx context.Context, maxTime time.Duration, filename string) error {
	return errors.New("portaudio not found")

}
