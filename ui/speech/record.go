package speech

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/garlicgarrison/go-recorder/recorder"
	"github.com/garlicgarrison/go-recorder/stream"
)

type SpeechConfig struct {
	Duration time.Duration
	Lang     string
}

func SpeechToText(ctx context.Context, aiContext context.Context, config *SpeechConfig) (string, error) {
	tmpFileName := fmt.Sprintf("speech-%d.wav", time.Now().UnixNano())
	err := recordAudio(ctx, tmpFileName, config.Duration)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFileName)

	if f, err := os.Open(tmpFileName); err == nil {
		stat, _ := f.Stat()
		if stat.Size() >= 26214400 {
			return "", errors.New("file too big")
		}
	}

	return api.SpeechToText(aiContext, tmpFileName, config.Lang)
}

func recordAudio(ctx context.Context, filename string, maxDuration time.Duration) error {
	stream, err := stream.NewStream(stream.DefaultStreamConfig())
	if err != nil {
		return err
	}
	defer stream.Terminate()

	cfg := recorder.DefaultRecorderConfig()
	cfg.MaxTime = int(maxDuration / time.Millisecond)

	rec, err := recorder.NewRecorder(cfg, stream)

	if err != nil {
		return err
	}

	var recording *bytes.Buffer
	stream.Start()
	defer stream.Close()
	doneChan := make(chan bool)
	go func() {
		<-ctx.Done()
		doneChan <- true
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
