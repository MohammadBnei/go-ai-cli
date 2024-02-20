/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/MohammadBnei/go-ai-cli/service"
	"github.com/MohammadBnei/go-ai-cli/ui/chat"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Start the prompt loop",
	Run: func(cmd *cobra.Command, args []string) {
		promptConfig := &service.PromptConfig{
			ChatMessages: service.NewChatMessages("default"),
		}

		defaulSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM_DEFAULT)
		savedSystemPrompt := viper.GetStringMapString(config.PR_SYSTEM)
		for k := range defaulSystemPrompt {
			promptConfig.ChatMessages.AddMessage(savedSystemPrompt[k], service.RoleSystem)
		}

		if viper.GetBool(config.C_AUTOLOAD) {
			fmt.Println("Loading last chat...")
			configFolder := filepath.Dir(viper.ConfigFileUsed())
			err := promptConfig.ChatMessages.LoadFromFile(configFolder + "/last-chat.yml")
			if err != nil {
				fmt.Printf("An error occured trying to get the last chat : %s\nPress enter to continue", err)
				fmt.Scanln()
			}
		}

		updateChan := make(chan service.ChatMessage)
		defer close(updateChan)
		promptConfig.UpdateChan = updateChan

		chat.Chat(promptConfig)

	},
}

func init() {
	RootCmd.AddCommand(promptCmd)

	promptCmd.PersistentFlags().Bool(config.C_AUTOLOAD, false, "Automatically load the prompt from $CONFIG/last-chat.yml")

	viper.BindPFlag("autoSave", promptCmd.Flags().Lookup(config.C_AUTOLOAD))
}
