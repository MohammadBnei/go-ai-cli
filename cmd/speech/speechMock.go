// +build !portaudio

package speech

import (
	"fmt"

	"github.com/spf13/cobra"
	"moul.io/banner"
)

var SpeechCmd = &cobra.Command{
	Use:   "speech",
	Short: "Convert your speech into text.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(banner.Inline("go ai cli - speech"))

		fmt.Println("Speech is not supported on this platform")
	},
}
