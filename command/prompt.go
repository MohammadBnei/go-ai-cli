package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/MohammadBnei/go-openai-cli/markdown"
	"github.com/MohammadBnei/go-openai-cli/service"
)

func SendPrompt(cfg *PromptConfig) error {
	mdWriter := markdown.NewMarkdownWriter()
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_, ok := <-c
		if ok {
			cancel()
		}
	}()

	var writer io.Writer
	writer = os.Stdout
	if cfg.MdMode {
		writer = mdWriter
	}
	response, err := service.SendPrompt(ctx, cfg.UserPrompt, writer)
	signal.Stop(c)
	close(c)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			return err
		}
		fmt.Println("↩️")
		cfg.PreviousPrompt = cfg.UserPrompt
	}
	if cfg.MdMode {
		mdWriter.Flush(response)
	}

	cfg.PreviousRes = response

	return nil
}
