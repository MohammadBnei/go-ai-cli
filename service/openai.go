package service

import (
	"context"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jinzhu/copier"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

// func SendWithToolCall(ctx context.Context, messages []ChatMessage, functions []function.FunctionDefinition) (string, error) {
// 	c := openai.NewClient(viper.GetString("OPENAI_KEY"))

// 	s := spinner.New(spinner.CharSets[26], 100*time.Millisecond)
// 	s.Start()
// 	defer s.Stop()

// 	model := viper.GetString("model")

// 	if model == "" {
// 		model = openai.GPT4TurboPreview
// 	}

// 	chatMessages := []openai.ChatCompletionMessage{}
// 	err := copier.Copy(&chatMessages, messages)

// 	if err != nil {
// 		return "", err
// 	}

// 	resp, err := c.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
// 		Model:    model,
// 		Messages: chatMessages,
// 		Tools:    tools,
// 	})

// 	if err != nil {
// 		return "", err
// 	}
// 	toolCalled := ""

// 	for _, k := range resp.Choices {
// 		if k.FinishReason == openai.FinishReasonToolCalls {
// 			toolCalled = k.Message.FunctionCall.Name
// 			args := k.Message.FunctionCall.Arguments

// 			data := &ui.SaveFunctionData{}
// 			err := json.Unmarshal([]byte(args), data)
// 			if err != nil {
// 				return "", err
// 			}
// 			err = ui.SaveFileFunctionDef.Function(data)
// 			if err != nil {
// 				return "", err
// 			}
// 		}
// 	}

// 	return toolCalled, nil
// }

func SendPromptToOpenAi(ctx context.Context, request *GPTChanRequest) (<-chan *GPTChanResponse, error) {
	c := openai.NewClient(viper.GetString("OPENAI_KEY"))

	s := spinner.New(spinner.CharSets[26], 100*time.Millisecond)
	s.Start()
	defer s.Stop()

	if request.Model == "" {
		request.Model = viper.GetString("model")
	}

	chatMessages := []openai.ChatCompletionMessage{}
	err := copier.Copy(&chatMessages, request.Messages)
	if err != nil {
		return nil, err
	}

	req := openai.ChatCompletionRequest{
		Model:    request.Model,
		Messages: chatMessages,
		Stream:   true,
	}

	resp, err := c.CreateChatCompletionStream(
		ctx,
		req,
	)
	if err != nil {
		return nil, err
	}

	stream := make(chan *GPTChanResponse)

	go func(resp *openai.ChatCompletionStream) {
		defer resp.Close()
		defer close(stream)
	ResponseLoop:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				data, err := resp.Recv()
				s.Stop()
				if err != nil {
					stream <- &GPTChanResponse{
						Content: nil,
						Err:     err,
					}
					break ResponseLoop
				}
				stream <- &GPTChanResponse{
					Content: []byte(data.Choices[0].Delta.Content),
					Err:     nil,
				}
			}

		}
	}(resp)

	return stream, nil
}

func GetModelList() ([]string, error) {
	c := openai.NewClient(viper.GetString("OPENAI_KEY"))
	models, err := c.ListModels(context.Background())
	if err != nil {
		return nil, err
	}

	modelsList := []string{}
	for _, model := range models.Models {
		modelsList = append(modelsList, model.ID)
	}

	return modelsList, nil
}
