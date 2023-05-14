/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set the configuration in a file",
	Run: func(cmd *cobra.Command, args []string) {
		if l, _ := cmd.Flags().GetBool("list-model"); l {
			modelList, err := service.GetModelList()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(strings.Join(modelList, "\n"))
			return
		}

		err := viper.BindPFlag("OPENAI_KEY", cmd.Flags().Lookup("OPENAI_KEY"))
		if err != nil {
			fmt.Println(err)
			return
		}
		err = viper.BindPFlag("model", cmd.Flags().Lookup("model"))
		if err != nil {
			fmt.Println(err)
			return
		}
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		if err := viper.WriteConfigAs(home + "/.go-openai-cli.yaml"); err != nil {
			fmt.Printf("Error creating config file: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	configCmd.PersistentFlags().StringP("OPENAI_KEY", "o", "", "the open ai key to be added to config")
	configCmd.PersistentFlags().StringP("model", "m", openai.GPT3Dot5Turbo, "the model to use")
	configCmd.PersistentFlags().BoolP("list-model", "l", false, "list the avalaible models")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
