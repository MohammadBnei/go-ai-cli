/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Start the prompt loop",
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("md", cmd.Flags().Lookup("md"))

		ui.OpenAiPrompt()
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)

	promptCmd.PersistentFlags().Int("depth", 2, "the depth of the tree view, when in file mode")
	promptCmd.PersistentFlags().Bool("md", false, "markdown mode enabled")

}
