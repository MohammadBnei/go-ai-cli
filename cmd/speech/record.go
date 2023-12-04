// +build portaudio

package speech

import (
	"fmt"
	"time"

	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/spf13/cobra"
)

// recordCmd represents the record command
var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record audio to file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := service.RecordAudioToFile(1*time.Minute, false, cmd.Flag("filename").Value.String()); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%s.wav saved\n", cmd.Flag("filename").Value.String())
	},
}

func init() {
	SpeechCmd.AddCommand(recordCmd)

	recordCmd.MarkFlagRequired("filename")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// recordCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// recordCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
