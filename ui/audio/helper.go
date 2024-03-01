package audio

import (
	"sort"

	bList "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/golang-module/carbon"
	"github.com/samber/lo"

	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
)

func getFilesAsItem(files []service.FileMetadata, pc *service.PromptConfig) []bList.Item {
	items := lo.Map(files, func(file service.FileMetadata, _ int) list.Item {
		msg := pc.ChatMessages.FindById(file.MsgID)
		return list.Item{
			ItemId:          file.ID,
			ItemTitle:       msg.Content,
			ItemDescription: file.Timestamp.Format("2020-12-31 15:04:05"),
		}
	})

	sort.Slice(items, func(i, j int) bool {
		return carbon.Parse(items[i].ItemDescription).Gt(carbon.Parse(items[j].ItemDescription))
	})

	return lo.Map(items, func(i list.Item, _ int) bList.Item { return i })
}

func getDelegateFn(pc *service.PromptConfig) *list.DelegateFunctions {
	return &list.DelegateFunctions{
		ChooseFn: func(id string) tea.Cmd {
			return SelectAudioFile(id)
		},
	}
}

type SelectAudioFileEvent struct {
	Id string
}

func SelectAudioFile(id string) tea.Cmd {
	return func() tea.Msg {
		return SelectAudioFileEvent{
			Id: id,
		}
	}
}
