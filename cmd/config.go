/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
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
		err = viper.BindPFlag("HUGGING_KEY", cmd.Flags().Lookup("HUGGINGFACE_KEY"))
		if err != nil {
			fmt.Println(err)
			return
		}
		err = viper.BindPFlag("model", cmd.Flags().Lookup("model"))
		if err != nil {
			fmt.Println(err)
			return
		}
		err = viper.BindPFlag("messages-length", cmd.Flags().Lookup("messages-length"))
		if err != nil {
			fmt.Println(err)
			return
		}
		err = viper.BindPFlag("config-path", cmd.Flags().Lookup("config"))
		if err != nil {
			fmt.Println(err)
			return
		}

		err = viper.BindPFlag("md", cmd.Flags().Lookup("md"))
		if err != nil {
			fmt.Println(err)
			return
		}

		filePath := viper.GetString("config-path")
		created := false
		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			path, _ := strings.CutSuffix(filePath, "/config.yaml")
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Created config directory : " + filePath)
			created = true
		}

		viper.SetConfigFile(filePath)
		if err := viper.WriteConfigAs(filePath); err != nil {
			fmt.Printf("Error creating config file: %s", err)
			return
		}
		if created {
			fmt.Println("Created config file : " + filePath)
		} else {
			fmt.Println("Updated config file : " + filePath)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	configCmd.PersistentFlags().StringP("model", "m", openai.GPT3Dot5Turbo, "the model to use")
	configCmd.PersistentFlags().BoolP("list-model", "l", false, "list the avalaible models")

	rootCmd.PersistentFlags().StringP("OPENAI_KEY", "o", "", "the open ai key to be added to config")
	rootCmd.PersistentFlags().String("HUGGINGFACE_KEY", "", "the hugging face key to be added to config")
	rootCmd.PersistentFlags().IntP("messages-length", "d", 20, "the number of messages to remember (all messages will be sent for every requests)")

	configCmd.PersistentFlags().Bool("md", false, "markdown mode enabled")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
