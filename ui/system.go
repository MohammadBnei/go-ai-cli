package ui

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"time"

	"github.com/golang-module/carbon/v2"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/tigergraph/promptui"
)

func SendAsSystem() (string, error) {
	systemPrompt := promptui.Prompt{
		Label: "specify model behavior",
	}
	command, err := systemPrompt.Run()
	if err != nil {
		return "", err
	}

	if YesNoPrompt("save prompt ?") {
		AddToSystemList(command, time.Now().Format("2006-01-02 15:04:05"))
	}

	return command, nil
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

func StringPrompt(label string) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
	}

	res, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return res, nil
}

func SetSystemDefault(unset bool) (commandToAdd []string, err error) {
	savedSystemPrompt := viper.GetStringMapString("systems")
	savedDefaultSystemPrompt := viper.GetStringMapString("default-systems")
	keyStringFromSP := lo.Keys[string](savedSystemPrompt)
	sort.Slice(keyStringFromSP, func(i, j int) bool {
		return carbon.Parse(keyStringFromSP[i]).Gt(carbon.Parse(keyStringFromSP[j]))
	})
	if savedDefaultSystemPrompt == nil {
		savedDefaultSystemPrompt = make(map[string]string)
	}
	keys, err := SystemPrompt(savedSystemPrompt, func(i, w, h int) string {
		defaultStr := "❌"
		_, ok := savedDefaultSystemPrompt[keyStringFromSP[i]]

		switch {
		case i == -1:
			return ""
		case ok:
			defaultStr = "✅"
		}

		return fmt.Sprintf("Date: %s\nDefault: %s\n%s", keyStringFromSP[i], defaultStr, AddReturnOnWidth(w/3-1, savedSystemPrompt[keyStringFromSP[i]]))
	})
	if err != nil {
		return
	}

	sendCommands := false
	if !unset {
		sendCommands = YesNoPrompt("send commands ?")
	}

	for _, id := range keys {
		if unset {
			delete(savedDefaultSystemPrompt, id)
		} else {
			savedDefaultSystemPrompt[id] = ""
			if sendCommands {
				commandToAdd = append(commandToAdd, savedSystemPrompt[id])
			}
		}

	}

	viper.Set("default-systems", savedDefaultSystemPrompt)
	viper.GetViper().WriteConfig()

	err = viper.GetViper().WriteConfig()

	return
}

func SelectSystemCommand() ([]string, error) {
	savedSystemPrompt := viper.GetStringMapString("systems")
	keys, err := SystemPrompt(savedSystemPrompt, nil)
	if err != nil {
		return nil, err
	}

	commandToSend := []string{}
	for _, id := range keys {
		commandToSend = append(commandToSend, savedSystemPrompt[id])
	}
	return commandToSend, nil
}

func SystemPrompt(savedSystemPrompt map[string]string, previewWindowFunc func(int, int, int) string) ([]string, error) {
	keyStringFromMap := lo.Keys[string](savedSystemPrompt)
	if len(keyStringFromMap) == 0 {
		return nil, errors.New("no saved systems")
	}
	sort.Slice(keyStringFromMap, func(i, j int) bool {
		return carbon.Parse(keyStringFromMap[i]).Gt(carbon.Parse(keyStringFromMap[j]))
	})
	if previewWindowFunc == nil {
		previewWindowFunc = func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Date: %s\n%s", keyStringFromMap[i], AddReturnOnWidth(w/3-1, savedSystemPrompt[keyStringFromMap[i]]))
		}
	}

	idx, err := fuzzyfinder.FindMulti(
		keyStringFromMap,
		func(i int) string {
			return savedSystemPrompt[keyStringFromMap[i]]
		},
		fuzzyfinder.WithPreviewWindow(previewWindowFunc),
	)
	if err != nil {
		return nil, err
	}

	return lo.Map[int, string](idx, func(i int, _ int) string {
		return keyStringFromMap[i]
	}), nil
}

func DeleteSystemCommand() error {
	savedSystemPrompt := viper.GetStringMapString("systems")
	keys, err := SystemPrompt(savedSystemPrompt, nil)
	if err != nil {
		return err
	}
	for _, id := range keys {
		RemoveFromSystemList(id)
		fmt.Printf("removed %s \n", id)
	}
	return nil
}

func AddToSystemList(command string, key string) {
	if key == "" {
		key = time.Now().Format("2006-01-02 15:04:05")
	}
	systems := viper.GetStringMapString("systems")
	systems[key] = command
	viper.Set("systems", systems)
	viper.GetViper().WriteConfig()
}
func RemoveFromSystemList(time string) {
	systems := viper.GetStringMapString("systems")
	delete(systems, time)
	viper.Set("systems", systems)
	viper.GetViper().WriteConfig()
}

var clear map[string]func() = make(map[string]func())

func ClearTerminal() {
	if _, ok := clear["linux"]; !ok {
		clear["linux"] = func() {
			cmd := exec.Command("clear") //Linux example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
	}
	if _, ok := clear["windows"]; !ok {
		clear["windows"] = func() {
			cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
			cmd.Stdout = os.Stdout
			cmd.Run()
		}
	}
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		clear["linux"]()
	}
}
