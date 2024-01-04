//go:build portaudio
// +build portaudio

package audio

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"os"

	"github.com/MohammadBnei/go-openai-cli/api"
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
	err := RecordAudioToFile(config.MaxMinutes, config.Detect, tmpFileName)
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFileName + ".wav")

	s := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	s.Start()
	defer s.Stop()

	return api.SendAudio(ctx, tmpFileName+".wav", config.Lang)
}

func RecordAudioToFile(maxTime time.Duration, detect bool, filename string) error {
	quit := make(chan bool, 2)
	go func(quit chan bool) {
		fmt.Println("Press enter to stop recording")
		fmt.Scanln()
		select {
		case _, ok := <-quit:
			if !ok {
				return
			}
		default:
			quit <- true

		}
	}(quit)
	go func(quit chan bool) {
		time.Sleep(maxTime)
		quit <- true
		close(quit)
	}(quit)

	stream, err := stream.NewStream(stream.DefaultStreamConfig())
	if err != nil {
		log.Fatalf("stream error -- %s", err)
	}
	defer stream.Terminate()

	cfg := recorder.DefaultRecorderConfig()
	cfg.MaxTime = int(maxTime)

	rec, err := recorder.NewRecorder(cfg, stream)

	if err != nil {
		log.Fatalf("recorder error -- %s", err)
	}

	fmt.Print("Recording...")
	var recording *bytes.Buffer
	if detect {
		recording, err = rec.RecordVAD(recorder.WAV)
	} else {
		stream.Start()
		defer stream.Close()
		recording, err = rec.Record(recorder.WAV, quit)
	}
	if err != nil {
		return err
	}
	fmt.Println(" done.")

	if filename == "" {
		filename = "speech"
	}
	file, err := os.Create(filename + ".wav")
	if err != nil {
		return err
	}

	_, err = file.Write(recording.Bytes())
	if err != nil {
		return err
	}

	return nil
}
