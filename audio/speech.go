//go:build portaudio

package audio

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"os"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/briandowns/spinner"
	"github.com/garlicgarrison/go-recorder/recorder"
	"github.com/garlicgarrison/go-recorder/stream"
)

type SpeechConfig struct {
	MaxMinutes time.Duration
	Lang       string
	Detect     bool
}

func SpeechToText(ctx context.Context, config *SpeechConfig) (string, error) {
	tmpFileName := fmt.Sprintf("speech-%d", time.Now().UnixNano())
	err := RecordAudioToFile(ctx, config.MaxMinutes, tmpFileName)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFileName + ".wav")

	s := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	s.Start()
	defer s.Stop()

	return api.SpeechToText(ctx, tmpFileName+".wav", config.Lang)
}

func RecordAudioToFile(ctx context.Context, maxTime time.Duration, filename string) error {

	stream, err := stream.NewStream(stream.DefaultStreamConfig())
	if err != nil {
		return err
	}
	defer stream.Terminate()

	cfg := recorder.DefaultRecorderConfig()
	cfg.MaxTime = int(maxTime)

	rec, err := recorder.NewRecorder(cfg, stream)

	if err != nil {
		return err
	}

	var recording *bytes.Buffer
	stream.Start()
	defer stream.Close()
	doneChan := make(chan bool)
	go func() {
		select {
		case <-ctx.Done():
			doneChan <- true
		}
	}()
	recording, err = rec.Record(recorder.WAV, doneChan)
	if err != nil {
		return err
	}

	if filename == "" {
		filename = "speech.wav"
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	_, err = file.Write(recording.Bytes())
	if err != nil {
		return err
	}

	return nil
}
