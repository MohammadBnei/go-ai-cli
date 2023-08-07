//go:build portaudio

package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"time"

	"github.com/MohammadBnei/go-openai-cli/markdown"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/atotto/clipboard"
	"github.com/manifoldco/promptui"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
)

type SpeechConfig struct {
	MaxMinutes   int
	Lang         string
	AutoMode     bool
	AutoFilename string
	MarkdownMode bool
	Format       bool
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

	ctx1, closer := LoadContext(ctx)
	speech, err := service.SpeechToText(ctx1, cfg.Lang, time.Duration(cfg.MaxMinutes)*time.Minute, false)
	closer()
	if err != nil {
		return err
	}

	fmt.Print("\n---\n", speech, "\n---\n\n")

	if cfg.AutoMode {
		err := AddToFile([]byte(speech), cfg.AutoFilename)
		if err != nil {
			return err
		}
		return nil
	}

	if lo.SomeBy[service.ChatMessage](service.GetMessages(), func(m service.ChatMessage) bool {
		return m.Role == openai.ChatMessageRoleSystem
	}) {
		var writer io.Writer = os.Stdout
		if cfg.MarkdownMode {
			writer = markdown.NewMarkdownWriter()
		}
		fmt.Print("Formating with openai : \n---\n\n")
		ctx1, closer := LoadContext(ctx)
		text, err := service.SendPrompt(ctx1, speech, writer)
		closer()
		if cfg.MarkdownMode {
			writer.(*markdown.MarkdownWriter).Flush(text)
		}
		if err != nil {
			return err
		} else {
			speech = text
		}
		fmt.Print("\n\n---\n\n")
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
			ctx1, closer := LoadContext(ctx)
			filename, err = service.SpeechToText(ctx1, cfg.Lang, 3*time.Second, false)
			closer()
			filename = strings.TrimSpace(filename)
			fmt.Printf(" Filename : '%s'\n", filename)
			switch {
			case err != nil:
				fmt.Println(err)
				continue filenameLoop
			case filename == "":
				break filenameLoop
			case YesNoPrompt(fmt.Sprintf("Filename : %s", filename)):
				break filenameLoop
			}
		}
		SaveToFile([]byte(speech), filename)
	case 2:
		return nil
	case 3:
		os.Exit(0)
	}

	fmt.Print("\n✅\n\n")

	return nil
}

func LoadContext(ctx context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_, ok := <-c
		if ok {
			cancel()
		}
	}()
	return ctx, func() {
		signal.Stop(c)
		close(c)
	}
}
