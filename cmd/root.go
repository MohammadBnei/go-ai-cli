/*
Copyright © 2023 Mohammad-Amine BANAEI mohammadamine.banaei@pm.me

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/MohammadBnei/go-ai-cli/api"
	"github.com/MohammadBnei/go-ai-cli/cmd/speech"
	"github.com/fsnotify/fsnotify"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "go-ai-cli",
	Short: "go-ai-cli is a command-line interface that allows users to generate text using OpenAI's GPT-3 language generation service.",
	Long:  `go-ai-cli is a command-line interface tool that provides users with convenient access to OpenAI's GPT-3 language generation service. With this app, users can easily send prompts to the OpenAI API and receive generated responses, which can then be printed on the command-line or saved to a markdown file. go-ai-cli is an excellent tool for creatives, content creators, chatbot developers and virtual assistants, as they can use it to quickly generate text for various purposes. By configuring their OpenAI API key and model, users can customize the behavior of the app to suit their specific needs. Moreover, go-ai-cli is an open-source project that welcomes contributions from the community, and it is licensed under the MIT License.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "configfile", home+"/.config/go-ai-cli/config.yaml", "config file (default is $HOME/.config/go-ai-cli/config.yaml)")
	RootCmd.PersistentFlags().StringP("OPENAI_KEY", "o", "", "the open ai key to be added to config")
	RootCmd.PersistentFlags().String("HUGGINGFACE_KEY", "", "the hugging face key to be added to config")

	RootCmd.PersistentFlags().StringP("API_TYPE", "t", api.API_OLLAMA, "the api type to be added to config")
	RootCmd.RegisterFlagCompletionFunc("API_TYPE", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{api.API_HUGGINGFACE, api.API_OLLAMA, api.API_OPENAI}, cobra.ShellCompDirectiveDefault
	})

	RootCmd.PersistentFlags().StringP("model", "m", openai.GPT4, "the model to use")
	RootCmd.RegisterFlagCompletionFunc("model", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		apiType, err := cmd.Flags().GetString("API_TYPE")
		if err != nil || apiType == "" {
			apiType = api.API_OLLAMA
		}
		switch apiType {
		case api.API_OLLAMA:
			models, err := api.GetOllamaModelList()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			return models, cobra.ShellCompDirectiveDefault
		case api.API_OPENAI:
			models, err := api.GetOpenAiModelList()
			if err != nil {
				return nil, cobra.ShellCompDirectiveError
			}
			return models, cobra.ShellCompDirectiveDefault
		}
		return nil, cobra.ShellCompDirectiveDefault
	})

	RootCmd.AddCommand(speech.SpeechCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if os.Getenv("CONFIG") != "" {
		cfgFile = os.Getenv("CONFIG")
	}
	// Use config file from the flag.
	viper.SetConfigFile(cfgFile)

	viper.BindPFlags(RootCmd.PersistentFlags())

	viper.SetDefault("OLLAMA_HOST", "http://127.0.0.1:11434")
	viper.SetDefault("auto-save", true)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err = viper.WriteConfig(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}

		}
		fmt.Fprintln(os.Stderr, err)
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		viper.ReadInConfig()
	})
	go viper.WatchConfig()
}
