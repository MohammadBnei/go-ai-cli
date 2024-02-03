/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MohammadBnei/go-openai-cli/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set the configuration in a file",
	Run: func(cmd *cobra.Command, args []string) {
		if l, _ := cmd.Flags().GetBool("list-model"); l {
			modelList, err := api.GetOpenAiModelList()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(strings.Join(modelList, "\n"))
			return
		}

		filePath := viper.GetString("config-path")
		folders := strings.Split(filePath, "/")
		created := false
		if _, err := os.Stat(filePath); folders[0] != filePath && errors.Is(err, os.ErrNotExist) {
			path := strings.Join(folders[:len(folders)-1], "/")
			err := os.MkdirAll(path, os.ModePerm)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Created config directory : " + path)
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
	RootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	configCmd.Flags().BoolP("list-model", "l", false, "list the avalaible models")
	configCmd.Flags().IntP("messages-length", "d", 20, "the number of messages to remember (all messages will be sent for every requests)")

	configCmd.Flags().Bool("md", false, "markdown mode enabled")

	viper.BindPFlags(configCmd.Flags())

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
