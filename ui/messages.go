package ui

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/markdown"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/tool"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func SaveChat(chatmessages *service.ChatMessages) error {
	name, err := StringPrompt("Enter a name for this chat")
	if err != nil {
		return err
	}

	chatmessages.SetId(name)

	description, err := StringPrompt("Enter a description for this chat")
	if err != nil {
		return err
	}

	chatmessages.SetDescription(description)

	data, err := yaml.Marshal(chatmessages)
	if err != nil {
		return err
	}

	err = tool.SaveToFile(data, viper.GetString("configPath")+"/"+name+".yaml", false)
	if err != nil {
		return err
	}

	return nil
}

func LoadChat(startPath string) (*service.ChatMessages, error) {
	if startPath == "" {
		startPath = viper.GetString("configPath")
	}
	path, err := PathSelectionFzf(startPath)
	if err != nil {
		return nil, err
	}
	fileContents, err := FileSelectionFzf(path)
	if err != nil {
		return nil, err
	}

	if len(fileContents) != 1 {
		return nil, errors.New("please select only one file")
	}

	loadedChat := &service.ChatMessages{}
	err = yaml.Unmarshal([]byte(fileContents[0]), loadedChat)
	if err != nil {
		return nil, err
	}

	loadedChat.RecountTokens()

	return loadedChat, nil
}

func FilterMessages(messages []service.ChatMessage) (messageIds []int, err error) {
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

			return fmt.Sprintf("%s\n%s", messages[i].Date.String(), AddReturnOnWidth(w/3-1, messages[i].Content))
		}),
	)

	if err != nil {
		return
	}

	for _, i := range idx {
		messageIds = append(messageIds, messages[i].Id)
	}

	if err != nil {
		return
	}

	fmt.Printf("cleared %d messages \n", len(idx))

	return
}

func ReuseMessage(messages []service.ChatMessage) (string, error) {
	messages = lo.Filter[service.ChatMessage](messages, func(item service.ChatMessage, _ int) bool {
		return item.Role == openai.ChatMessageRoleUser
	})

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.After(messages[j].Date)
	})

	id, err := fuzzyfinder.Find(
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
		return "", err
	}

	return messages[id].Content, nil

}

func ShowPreviousMessage(messages []service.ChatMessage, markdownMode bool) (string, error) {
	messages = lo.Filter[service.ChatMessage](messages, func(item service.ChatMessage, _ int) bool {
		return item.Role == openai.ChatMessageRoleAssistant
	})

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.After(messages[j].Date)
	})
	id, err := fuzzyfinder.Find(
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
		return "", err
	}

	if !markdownMode {
		fmt.Println(messages[id].Content)
		return messages[id].Content, nil
	}

	mdWriter := markdown.NewMarkdownWriter()
	mdText, _ := mdWriter.ToMarkdown(messages[id].Content)

	fmt.Println(mdText)
	return mdText, nil
}
