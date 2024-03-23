/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/chat"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Start the prompt loop",
	Run: func(cmd *cobra.Command, args []string) {
		fileService, err := service.NewFileService()
		if err != nil {
			fmt.Println(err)
			return
		}

		services := &service.Services{
			ChatMessages: service.NewChatMessages("default"),
			Files:        fileService,
			Contexts:     service.NewContextService(),
		}

		defer func() {
			if err := recover(); err != nil {
				services.ChatMessages.SaveToFile(filepath.Dir(viper.ConfigFileUsed()) + "/error-chat.yml")
				fmt.Println(err)
			}
		}()

		defaulSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM_DEFAULT)
		savedSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM)
		for k := range defaulSystemPrompt {
			services.ChatMessages.AddMessage(savedSystemPrompt[k], service.RoleSystem)
		}

		if viper.GetBool(config.C_AUTOLOAD) {
			fmt.Println("Loading last chat...")
			configFolder := filepath.Dir(viper.ConfigFileUsed())
			err := services.ChatMessages.LoadFromFile(configFolder + "/last-chat.yml")
			if err != nil {
				fmt.Printf("An error occured trying to get the last chat : %s\nPress enter to continue", err)
				fmt.Scanln()
			}
		}

		updateChan := make(chan service.ChatMessage)
		defer close(updateChan)
		services.UpdateChan = updateChan

		chatModel, err := chat.NewChatModel(services)
		if err != nil {
			log.Fatal(err)
		}
		p := tea.NewProgram(chatModel,
			tea.WithAltScreen())

		chat.ChatProgram = p

		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	RootCmd.AddCommand(promptCmd)

	promptCmd.PersistentFlags().Bool(config.C_AUTOLOAD, false, "Automatically load the prompt from $CONFIG/last-chat.yml")

	viper.BindPFlag("autoSave", promptCmd.Flags().Lookup(config.C_AUTOLOAD))
}
