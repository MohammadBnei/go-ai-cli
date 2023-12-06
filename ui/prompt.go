package ui

import (
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

func BasicPrompt(label, previousPrompt string) (string, error) {
	prompt := promptui.Prompt{
		Label:     label,
		AllowEdit: false,
		Default:   previousPrompt,
	}

	userPrompt, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			os.Exit(0)
		}
		return "", err

	}

	return strings.TrimSpace(userPrompt), nil
}
