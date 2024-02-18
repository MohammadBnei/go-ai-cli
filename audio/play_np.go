//go:build !portaudio

package audio

import (
	"context"
	"errors"
	"io"
	"time"
)

type SpeechConfig struct {
	MaxMinutes time.Duration
	Lang       string
	Detect     bool
}

func SpeechToText(ctx context.Context, config *SpeechConfig) (string, error) {
	return "", errors.New("portaudio not present")
}

func RecordAudioToFile(ctx context.Context, maxTime time.Duration, filename string) error {
	return errors.New("portaudio not present")
}

func PlaySound(ctx context.Context, data io.ReadCloser) error {
	return errors.New("portaudio not present")

}

func PlayTextToSpeech(ctx context.Context, text string) error {
	return errors.New("portaudio not present")

}
