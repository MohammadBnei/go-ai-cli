package agent

import (
	"fmt"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/api/agent"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/event"
	"github.com/MohammadBnei/go-ai-cli/ui/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/tools"
)

const (
	WEB_SEARCH = "Web Search"
)

func NewAgentModel(promptConfig *service.PromptConfig) (tea.Model, error) {
	agentMap := map[string]*agents.Executor{}

	llm, err := api.GetLlmModel()
	if err != nil {
		return nil, err
	}

	agent, err := agent.NewAutoWebSearchAgent(llm)
	if err != nil {
		return nil, err
	}
	agentMap[WEB_SEARCH] = agent

	return list.NewFancyListModel("Agents", lo.MapToSlice(agentMap, func(k string, v *agents.Executor) list.Item {
		return list.Item{
			ItemId:          k,
			ItemTitle:       k,
			ItemDescription: lo.Reduce(v.Tools, func(agg string, r tools.Tool, _ int) string { return fmt.Sprintf("%s %s", agg, r.Name()) }, ""),
		}
	}), &list.DelegateFunctions{
		ChooseFn: getAgentFn(agentMap),
	}), nil
}

func getAgentFn(agentMap map[string]*agents.Executor) func(id string) tea.Cmd {
	return func(id string) tea.Cmd {
		selected, ok := agentMap[id]
		if !ok {
			return event.Error(fmt.Errorf("agent %s not found", id))
		}
		switch id {
		case WEB_SEARCH:
			return tea.Sequence(event.RemoveStack(nil), event.AgentSelection(selected, id))

		}

		return nil
	}
}
