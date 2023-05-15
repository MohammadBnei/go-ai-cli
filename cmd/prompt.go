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
}
