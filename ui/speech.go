//go:build portaudio
// +build portaudio

package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/audio"
	"github.com/MohammadBnei/go-ai-cli/markdown"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/tool"
	"github.com/MohammadBnei/go-ai-cli/ui/helper"
	"github.com/atotto/clipboard"
	"github.com/manifoldco/promptui"
)

type SpeechConfig struct {
	MaxMinutes        int
	Lang              string
	AutoMode          bool
	AutoFilename      string
	MarkdownMode      bool
	Format            bool
	Timestamp         bool
	SystemOptions     []string
	AdvancedFormating string
	ChatMessages      *service.ChatMessages
}

func SpeechLoop(ctx context.Context, cfg *SpeechConfig) error {
	done := atomic.Bool{}
	done.Store(false)
	defer func(done *atomic.Bool) {
		done.Store(true)
	}(&done)

	fmt.Println("Press enter to start")
	fmt.Scanln()
	go func() {
		time.Sleep(time.Duration(cfg.MaxMinutes)*time.Minute - 15*time.Second)
		if done.Load() {
			return
		}
		if done.Load() {
			return
		}
		fmt.Print("15 seconds remaining...")
		time.Sleep(10 * time.Second)
		for i := 5; i > 0; i-- {
			if done.Load() {
				return
			}
			fmt.Printf("%d seconds remaining...", i)
			time.Sleep(1 * time.Second)
		}
	}()

	ctx1, closer := service.LoadContext(ctx)
	speech, err := audio.SpeechToText(ctx1, &audio.SpeechConfig{Lang: cfg.Lang, MaxMinutes: time.Duration(cfg.MaxMinutes) * time.Minute, Detect: false})
	closer()
	if err != nil {
		return err
	}

	fmt.Print("\n---\n", speech, "\n---\n\n")

	if msgs, _ := cfg.ChatMessages.FilterMessages(service.RoleAssistant); len(msgs) != 0 {
		speech, err = FormatWithOpenai(ctx, cfg)
		if err != nil {
			return err
		}
	}

	if cfg.AutoMode {
		err := AddToFile([]byte(speech), cfg.AutoFilename, true)
		if err != nil {
			return err
		}
		return nil
	}

	selectionPrompt := promptui.Select{
		Label: "Speech converted to text. What do you want to do with it ?",
		Items: []string{"Copy to clipboard", "Save in file", "another speech", "quit"},
	}

	id, _, err := selectionPrompt.Run()
	if err != nil {
		return err
	}

	switch id {
	case 0:
		clipboard.WriteAll(speech)
	case 1:
		filename := ""
	filenameLoop:
		for {
			fmt.Println("Specify the filename orally. If you don't want to specify, press enter twice.")
			fmt.Println("Press enter to record")
			fmt.Scanln()
			ctx1, closer := service.LoadContext(ctx)
			filename, err = audio.SpeechToText(ctx1, &audio.SpeechConfig{Lang: cfg.Lang, MaxMinutes: 5 * time.Second, Detect: false})
			closer()
			filename = strings.TrimSpace(filename)
			fmt.Printf(" Filename : '%s'\n", filename)
			switch {
			case err != nil:
				fmt.Println(err)
				continue filenameLoop
			case filename == "":
				break filenameLoop
			case helper.YesNoPrompt(fmt.Sprintf("Filename : %s", filename)):
				break filenameLoop
			}
		}
		tool.SaveToFile([]byte(speech), filename, false)
	case 2:
		return nil
	case 3:
		os.Exit(0)
	}

	fmt.Print("\n✅\n\n")

	return nil
}

func ContinuousSpeech(ctx context.Context, cfg *SpeechConfig) error {
	speech := make(chan string)
	defer close(speech)
	go func() {
		for {
			txt, ok := <-speech
			if !ok {
				return
			}

			if msgs, _ := cfg.ChatMessages.FilterMessages(service.RoleAssistant); len(msgs) != 0 {
				var err error
				txt, err = FormatWithOpenai(ctx, cfg)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			err := AddToFile([]byte(txt), cfg.AutoFilename, false)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}()

	for {
		go func(speech chan<- string) {
			txt, err := audio.SpeechToText(ctx, &audio.SpeechConfig{Lang: cfg.Lang, MaxMinutes: time.Minute, Detect: false})
			if err != nil {
				fmt.Println(err)
				return
			}
			speech <- txt
		}(speech)

		time.Sleep(50 * time.Second)
	}

}

func FormatWithOpenai(ctx context.Context, cfg *SpeechConfig) (speech string, err error) {
	var writer io.Writer = os.Stdout
	if cfg.MarkdownMode {
		writer = markdown.NewMarkdownWriter()
	}
	fmt.Print("Formating with openai : \n---\n\n")
	ctx1, closer := service.LoadContext(ctx)
	defer closer()
	stream, err := api.SendPromptToOpenAi(ctx1, &api.GPTChanRequest{
		Messages: cfg.ChatMessages.Messages,
	})
	if err != nil {
		return
	}
	_, err = api.PrintTo(stream, writer.Write)
	if err != nil {
		return
	}
	if cfg.MarkdownMode {
		writer.(*markdown.MarkdownWriter).Flush(speech)
	}
	fmt.Print("\n\n---\n\n")

	return
}

func InitSpeech(speechConfig *SpeechConfig) error {
	speechConfig.ChatMessages = service.NewChatMessages("speech")
	for _, opt := range speechConfig.SystemOptions {
		speechConfig.ChatMessages.AddMessage(opt, service.RoleSystem)
	}

	fmt.Println(speechConfig.SystemOptions)

	if speechConfig.Format {
		speechConfig.ChatMessages.AddMessage("You will be prompted with a speech converted to text. Format it by adding line return between ideas and correct puntucation. Do not translate.", service.RoleSystem)

		if speechConfig.AdvancedFormating != "" {
			speechConfig.ChatMessages.AddMessage(speechConfig.AdvancedFormating, service.RoleSystem)
		}
	}

	return nil
}
