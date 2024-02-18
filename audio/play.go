//go:build portaudio

package audio

import (
	"context"
	"io"
	"time"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func PlaySound(ctx context.Context, data io.ReadCloser) error {
	streamer, format, err := mp3.Decode(data)
	if err != nil {
		return err
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/60))

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()

	speech := buffer.Streamer(0, buffer.Len())
	speaker.Play(speech)

	return nil
}

func PlayTextToSpeech(ctx context.Context, text string) error {
	data, err := api.TextToSpeech(ctx, text)
	if err != nil {
		return err
	}
	return PlaySound(ctx, data)
}
