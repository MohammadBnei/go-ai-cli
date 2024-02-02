/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/MohammadBnei/go-openai-cli/command"
	"github.com/MohammadBnei/go-openai-cli/prompt"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Start the prompt loop",
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("md", cmd.Flags().Lookup("md"))

		commandMap := make(map[string]func(*command.PromptConfig) error)

		command.AddAllCommand(commandMap)

		promptConfig := &command.PromptConfig{
			MdMode:       viper.GetBool("md"),
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

		prompt.Chat(promptConfig)

	},
}

func init() {
	RootCmd.AddCommand(promptCmd)

	promptCmd.PersistentFlags().Int("depth", 2, "the depth of the tree view, when in file mode")
	promptCmd.PersistentFlags().Bool("md", false, "markdown mode enabled")
	promptCmd.PersistentFlags().BoolP("auto-save", "s", false, "Automatically save the prompt to a file")

	viper.BindPFlag("autoSave", promptCmd.Flags().Lookup("auto-save"))
}
