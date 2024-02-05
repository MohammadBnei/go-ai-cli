//go:build portaudio
// +build portaudio

package speech

import (
	"fmt"

	"github.com/MohammadBnei/go-ai-cli/ui"
	"github.com/spf13/cobra"
	"moul.io/banner"
)

// continuousCmd represents the continuous command
var continuousCmd = &cobra.Command{
	Use:   "continuous",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(banner.Inline("go ai cli - continuous speech"))

		err := ui.InitSpeech(speechConfig)
		if err != nil {
			fmt.Println(err)
		}

		err = ui.ContinuousSpeech(cmd.Context(), speechConfig)
		if err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	SpeechCmd.AddCommand(continuousCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// continuousCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// continuousCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
