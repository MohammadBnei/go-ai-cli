package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/manifoldco/promptui"
)

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

type SendPromptConfig struct {
	ChatMessages *ChatMessages
	Output       io.Writer
	GPTFunc      func(ctx context.Context, messages []ChatMessage, output io.Writer) (string, error)
}

func SendPrompt(cfg *SendPromptConfig) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		_, ok := <-c
		if ok {
			cancel()
		}
	}()

	response, err := cfg.GPTFunc(ctx, cfg.ChatMessages.Messages, cfg.Output)
	signal.Stop(c)
	close(c)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			return "", err
		}
	}

	return response, nil
}

func SaveToFile(content []byte, filename string) error {
	if filename == "" {
		return errors.New("filename cannot be empty")
	}

	if strings.Contains(filename, "/") {
		splitted := strings.Split(filename, "/")
		dw := strings.Join(splitted[:len(splitted)-1], "/")

		if _, err := os.Stat(dw); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dw, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Created directory : " + dw)
		}
	}

	if _, err := os.Stat(filename); err == nil {
		replaceSelect := promptui.Select{
			Label: filename + " exists. Replace ?",
			Items: []string{"Yes", "No"},
		}

		i, _, err := replaceSelect.Run()
		if err != nil {
			return err
		}

		if i == 1 {
			return nil
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(content)
	fmt.Println("saved to", filename)

	return nil
}
