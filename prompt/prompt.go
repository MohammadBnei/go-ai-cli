package prompt

import (
	"errors"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
)

func CommandSelectionFactory() func(cmd string, pc *service.PromptConfig) error {
	commandMap := make(map[string]func(*service.PromptConfig) error)

	command.AddAllCommand(commandMap)
	keys := lo.Keys[string](commandMap)

	return func(cmd string, pc *service.PromptConfig) error {

		var err error

		switch {
		case cmd == "":
			commandMap["help"](pc)
		case cmd == "\\":
			selection, err2 := fuzzyfinder.Find(keys, func(i int) string {
				return keys[i]
			})
			if err2 != nil {
				return err2
			}

			err = commandMap[keys[selection]](pc)
		case strings.HasPrefix(cmd, "\\"):
			command, ok := commandMap[cmd[1:]]
			if !ok {
				return errors.New("command not found")
			}
			err = command(pc)
		}

		return err
	}
}
