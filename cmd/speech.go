//go:build portaudio

package cmd

import (
	"fmt"
	"time"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"moul.io/banner"
)

var format bool
var markdownMode bool
var advancedFormating string
var systemOptions []string
var maxMinutes int
var autoFilename string
var autoMode bool

// speechCmd represents the speech command
var speechCmd = &cobra.Command{
	Use:   "speech",
	Short: "Convert your speech into text.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(banner.Inline("go ai cli - speech"))
		if maxMinutes > 4 {
			maxMinutes = 4
		}

		for _, opt := range systemOptions {
			service.AddMessage(service.ChatMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: opt,
				Date:    time.Now(),
			})
		}

		if format {
			service.AddMessage(service.ChatMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You will be prompted with a speech converted to text. Format it by adding line return between ideas and correct puntucation. Do not translate.",
				Date:    time.Now(),
			})
			if advancedFormating != "" {
				service.AddMessage(service.ChatMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: advancedFormating,
					Date:    time.Now(),
				})
			}
		}

		cfg := &ui.SpeechConfig{
			MaxMinutes:   maxMinutes,
			Lang:         cmd.Flag("lang").Value.String(),
			Format:       format,
			MarkdownMode: markdownMode,
			AutoMode:     autoMode,
			AutoFilename: autoFilename,
		}

		for {
			err := ui.SpeechLoop(cmd.Context(), cfg)
			if err != nil {
				fmt.Println(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(speechCmd)

	speechCmd.PersistentFlags().StringP("lang", "l", "en", "language")
	speechCmd.Flags().BoolVarP(&format, "format", "f", false, "format the output with the carriage return character.")
	speechCmd.Flags().StringVarP(&advancedFormating, "advanced-format", "a", "add markdown formating. Add a title and a table of content from the content of the speech, and add the coresponding subtitles. Do not modify the content of the speech", "Add advanced formating that will be sent as system command to openai")
	speechCmd.Flags().BoolVarP(&markdownMode, "markdown", "m", false, "Format the output to markdown")
	speechCmd.Flags().StringArrayVarP(&systemOptions, "system", "s", []string{}, "additionnal system options")
	speechCmd.Flags().IntVarP(&maxMinutes, "max-minutes", "t", 4, "max record time (in minutes) (max : 4 minutes)")

	speechCmd.Flags().StringVarP(&autoFilename, "filename", "n", time.Now().Format("2006-01-02_15:04:05")+".txt", "When in auto mode, the name of the file")
	speechCmd.Flags().BoolVar(&autoMode, "auto", false, "Automatically save the speech to a file.")

}
