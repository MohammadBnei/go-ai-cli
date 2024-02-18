package agent

import (
	"fmt"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/tools"
)

type model struct {
	agentMap      map[string]*agents.Executor
	selectedAgent *agents.Executor

	title string

	viewport viewport.Model
	textarea textarea.Model

	promptConfig *service.PromptConfig

	agentFancyList tea.Model

	userPrompt string
}

func NewAgentModel(promptConfig *service.PromptConfig) (tea.Model, error) {
	agentMap := map[string]*agents.Executor{}

	llm, err := api.GetLlmModel()
	if err != nil {
		return nil, err
	}

	agent, err := api.NewWebSearchAgent(llm)
	if err != nil {
		return nil, err
	}
	agentMap["Web Search"] = agent

	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.Focus()
	ta.ShowLineNumbers = false

	return list.NewFancyListModel("Agents", lo.MapToSlice(agentMap, func(k string, v *agents.Executor) list.Item {
		return list.Item{
			ItemId:          k,
			ItemTitle:       k,
			ItemDescription: lo.Reduce(v.Tools, func(agg string, r tools.Tool, _ int) string { return fmt.Sprintf("%s %s", agg, r.Name()) }, ""),
		}
	}), &list.DelegateFunctions{
		ChooseFn: func(id string) tea.Cmd {
			selected, ok := agentMap[id]
			if !ok {
				return event.Error(fmt.Errorf("agent %s not found", id))
			}
			return tea.Sequence(event.RemoveStack(nil), event.AgentSelection(selected, id))
		},
	}), nil

}
