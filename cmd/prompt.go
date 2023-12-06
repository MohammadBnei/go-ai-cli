/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/MohammadBnei/go-openai-cli/prompt"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"moul.io/banner"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Start the prompt loop",
	Run: func(cmd *cobra.Command, args []string) {
		ui.ClearTerminal()
		fmt.Println(banner.Inline("go ai cli - prompt"))
		viper.BindPFlag("md", cmd.Flags().Lookup("md"))

		prompt.OpenAiPrompt()
	},
}

func init() {
	RootCmd.AddCommand(promptCmd)

	promptCmd.PersistentFlags().Int("depth", 2, "the depth of the tree view, when in file mode")
	promptCmd.PersistentFlags().Bool("md", false, "markdown mode enabled")
	promptCmd.PersistentFlags().BoolP("auto-save", "s", false, "Automatically save the prompt to a file")

	viper.BindPFlag("autoSave", promptCmd.Flags().Lookup("auto-save"))
}
