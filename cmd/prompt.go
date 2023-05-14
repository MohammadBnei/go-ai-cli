/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/spf13/cobra"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Send a prompt to OpenAI",
	Run: func(cmd *cobra.Command, args []string) {
		ui.OpenAiPrompt()
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	promptCmd.Flags().Bool("md", false, "Write the response to a markdown file")
}
