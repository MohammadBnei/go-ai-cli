//go:build portaudio
// +build portaudio

package speech

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Convert an audio file to text.",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

		if _, err := os.Open(args[0]); err != nil {
			return err
		}

		extension := filepath.Ext(args[0])
		if extension != ".mp3" && extension != ".wav" {
			return fmt.Errorf("only mp3 and wav files are supported")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, closer := service.LoadContext(cmd.Context())
		text, err := service.SendAudio(ctx, args[0], cmd.Flag("lang").Value.String())
		closer()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("---\n\n", text, "\n\n---")
		if ui.YesNoPrompt("Copy to clipboard?") {
			clipboard.WriteAll(text)
		}
	},
}

func init() {
	SpeechCmd.AddCommand(fileCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fileCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fileCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
