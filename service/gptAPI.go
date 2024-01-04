package service

import (
	"context"
	"errors"
	"io"
)

type GPTChanResponse struct {
	Content []byte
	Err     error
}

type GPTChanRequest struct {
	Messages []ChatMessage
	Stream   bool
	Model    string
}

type GPTApiFunc func(context.Context, *GPTChanRequest) (<-chan *GPTChanResponse, error)

func PrintTo(stream <-chan *GPTChanResponse, fn func([]byte) (int, error)) (response string, err error) {
	response = ""
	for {
		msg, ok := <-stream
		switch {
		case msg.Err == io.EOF:
			return
		case !ok:
			err = errors.New("channel closed")
			return
		case msg.Err != nil:
			err = msg.Err
			return
		default:
			response += string(msg.Content)
			_, err = fn(msg.Content)
			if err != nil {
				return
			}
		}
	}
}

func GetFullResponse(stream <-chan *GPTChanResponse) (string, error) {
	response := ""
	for r := range stream {
		if r.Err == io.EOF {
			break
		}
		if r.Err != nil {
			return "", r.Err
		}
		response += string(r.Content)
	}

	return response, nil
}
