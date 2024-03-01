package helper

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/manifoldco/promptui"
)

func CheckedStringHelper(yes bool) string {
	if yes {
		return "✅"
	}
	return "❌"
}

func YesNoPrompt(label string) bool {
	prompt := promptui.Select{
		Label: label,
		Items: []string{"yes", "no"},
	}

	_, choice, err := prompt.Run()
	if err != nil || choice == "no" {
		return false
	}

	return true
}

var ChatProgram *tea.Program
