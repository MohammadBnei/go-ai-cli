package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
)

func FilterMessages() error {
	messages := service.GetMessages()

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.After(messages[j].Date)
	})
	idx, err := fuzzyfinder.FindMulti(
		messages,
		func(i int) string {
			content := messages[i].Content
			if len(content) > 50 {
				content = content[:50] + "..."
			}
			return fmt.Sprintf("%s : %s", messages[i].Role, content)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}

			splitted := strings.Split(messages[i].Content, " ")
			acc := 0
			for i, word := range splitted {
				if acc > w*2/5 {
					splitted = append(splitted[:i], "\n")
					splitted = append(splitted, splitted[i+1:]...)
					acc = 0
				}
				acc += lo.RuneLength(word) + 1
			}

			return AddReturnOnWidth(w/3-1, fmt.Sprintf("%s\n%s", messages[i].Date.String(), strings.Join(splitted, " ")))
		}),
	)

	if err != nil {
		return err
	}

	messages = lo.Filter[service.ChatMessage](messages, func(_ service.ChatMessage, i int) bool {
		return !lo.Contains[int](idx, i)
	})

	service.SetMessages(messages)

	fmt.Printf("cleared %d messages \n", len(idx))

	return nil
}
