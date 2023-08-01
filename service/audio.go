package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"os"

	"github.com/briandowns/spinner"
	"github.com/garlicgarrison/go-recorder/recorder"
	"github.com/garlicgarrison/go-recorder/stream"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

func SpeechToText(ctx context.Context, lang string, maxTime time.Duration) (string, error) {
	c := openai.NewClient(viper.GetString("OPENAI_KEY"))

	if lang == "" {
		lang = "en"
	}

	err := RecordAudioToFile(maxTime)
	if err != nil {
		return "", err
	}
	defer os.Remove("speech.wav")

	s := spinner.New(spinner.CharSets[35], 100*time.Millisecond)
	s.Start()
	defer s.Stop()

	response, err := c.CreateTranscription(ctx, openai.AudioRequest{
		Model:    openai.Whisper1,
		Format:   "text",
		FilePath: "speech.wav",
		Language: lang,
	})
	if err != nil {
		return "", err
	}

	return response.Text, nil
}

func RecordAudioToFile(maxTime time.Duration) error {
	quit := make(chan bool)
	go func(quit chan bool) {
		fmt.Println("Press enter to stop recording")
		fmt.Scanln()
		quit <- true
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

	stream.Start()
	defer stream.Close()
	fmt.Print("Recording...")
	recording, err := rec.Record(recorder.WAV, quit)
	fmt.Println(" done.")
	if err != nil {
		return err
	}

	file, err := os.Create("speech.wav")
	if err != nil {
		return err
	}

	_, err = file.Write(recording.Bytes())
	if err != nil {
		return err
	}

	return nil
}
