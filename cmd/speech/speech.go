//go:build portaudio
// +build portaudio

package speech

import (
	"fmt"
	"time"

	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/spf13/cobra"
	"moul.io/banner"
)

var advancedFormating string
var systemOptions []string
var speechConfig = &ui.SpeechConfig{}

// SpeechCmd represents the speech command
var SpeechCmd = &cobra.Command{
	Use:   "speech",
	Short: "Convert your speech into text.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(banner.Inline("go ai cli - speech"))
		if speechConfig.MaxMinutes > 4 {
			speechConfig.MaxMinutes = 4
		}

		err := ui.InitSpeech(speechConfig)
		if err != nil {
			fmt.Println(err)
		}

		for {
			err := ui.SpeechLoop(cmd.Context(), speechConfig)
			if err != nil {
				fmt.Println(err)
			}
		}
	},
}

func init() {
	SpeechCmd.PersistentFlags().StringVarP(&speechConfig.Lang, "lang", "l", "en", "language")
	SpeechCmd.RegisterFlagCompletionFunc("lang", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"en", "fr", "fa", "ar", "es", "it"}, cobra.ShellCompDirectiveDefault
	})

	SpeechCmd.PersistentFlags().BoolVarP(&speechConfig.Format, "format", "f", false, "format the output with the carriage return character.")
	SpeechCmd.PersistentFlags().StringVarP(&speechConfig.AdvancedFormating, "advanced-format", "a", "add markdown formating. Add a title and a table of content from the content of the speech, and add the coresponding subtitles. Do not modify the content of the speech", "Add advanced formating that will be sent as system command to openai")
	SpeechCmd.PersistentFlags().BoolVarP(&speechConfig.MarkdownMode, "markdown", "m", false, "Format the output to markdown")
	SpeechCmd.PersistentFlags().StringArrayVarP(&speechConfig.SystemOptions, "system", "s", []string{}, "additionnal system options")
	SpeechCmd.PersistentFlags().IntVarP(&speechConfig.MaxMinutes, "max-minutes", "t", 4, "max record time (in minutes) (max : 4 minutes)")

	SpeechCmd.PersistentFlags().StringVarP(&speechConfig.AutoFilename, "filename", "n", time.Now().Format("2006-01-02_15:04:05")+".txt", "When in auto/continuous/record mode, the name of the file")
	SpeechCmd.PersistentFlags().BoolVar(&speechConfig.AutoMode, "auto", false, "Automatically save the speech to a file.")
	SpeechCmd.PersistentFlags().BoolVar(&speechConfig.Timestamp, "timestamp", true, "Add timestamp on each speech iteration.")

}
