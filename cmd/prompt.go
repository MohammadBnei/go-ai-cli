/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/MohammadBnei/go-ai-cli/command"
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
		viper.BindPFlag("md", cmd.Flags().Lookup("md"))

		commandMap := make(map[string]func(*service.PromptConfig) error)

		command.AddAllCommand(commandMap)

		promptConfig := &service.PromptConfig{
			ChatMessages: service.NewChatMessages("default"),
		}

		defaulSystemPrompt := viper.GetStringMapString("default-systems")
		savedSystemPrompt := viper.GetStringMapString("systems")
		for k := range defaulSystemPrompt {
			promptConfig.ChatMessages.AddMessage(savedSystemPrompt[k], service.RoleSystem)
		}

		updateChan := make(chan service.ChatMessage)
		defer close(updateChan)
		promptConfig.UpdateChan = updateChan

		chat.Chat(promptConfig)

		recover()

	},
}

func init() {
	RootCmd.AddCommand(promptCmd)

	promptCmd.PersistentFlags().Int("depth", 2, "the depth of the tree view, when in file mode")
	promptCmd.PersistentFlags().Bool("md", false, "markdown mode enabled")
	promptCmd.PersistentFlags().BoolP("auto-load", "s", false, "Automatically save the prompt to a file")

	viper.BindPFlag("autoSave", promptCmd.Flags().Lookup("auto-load"))
}
