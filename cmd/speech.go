//go:build portaudio

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MohammadBnei/go-openai-cli/markdown"
	"github.com/MohammadBnei/go-openai-cli/service"
	"github.com/MohammadBnei/go-openai-cli/ui"
	"github.com/atotto/clipboard"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"github.com/tigergraph/promptui"
)

var format bool
var markdownMode bool
var carriageReturnC []string
var advancedFormating string
var systemOptions []string
var maxMinutes int

// speechCmd represents the speech command
var speechCmd = &cobra.Command{
	Use:   "speech",
	Short: "Convert your speech into text.",
	Run: func(cmd *cobra.Command, args []string) {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-quit
			os.Exit(0)
		}()
		if maxMinutes > 5 {
			maxMinutes = 5
		}
		for {

			fmt.Println("Press enter to start")
			fmt.Scanln()
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				print := true
				go func(print *bool) {
					<-ctx.Done()
					*print = false
				}(&print)
				time.Sleep(time.Duration(maxMinutes) * time.Minute - 15 * time.Second)
				if print {
					fmt.Print("15 seconds remaining...")
				} else {
					return
				}
				time.Sleep(10 * time.Second)
				for i := 5; i > 0; i-- {
					if print {
						fmt.Printf("%d seconds remaining...", i)
					} else {
						return
					}
					time.Sleep(1 * time.Second)
				}
			}()

			speech, err := service.SpeechToText(ctx, cmd.Flag("lang").Value.String(), time.Duration(maxMinutes)*time.Minute, false)
			cancel()
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Print("\n---\n", speech, "\n---\n\n")

			for _, opt := range systemOptions {
				service.AddMessage(service.ChatMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: opt,
					Date:    time.Now(),
				})
			}

			if format {
				fmt.Print("Formating with openai : \n---\n\n")
				service.AddMessage(service.ChatMessage{
					Role: openai.ChatMessageRoleSystem,
					Content: fmt.Sprintf(
						"You will be prompted with a speech converted to text. Format it by changing occurences of '%s' with a carriage return, and correct puntucation. Do not translate.",
						strings.Join(carriageReturnC, ", "),
					),
					Date: time.Now(),
				})
				if advancedFormating != "" {
					service.AddMessage(service.ChatMessage{
						Role:    openai.ChatMessageRoleSystem,
						Content: advancedFormating,
						Date:    time.Now(),
					})
				}
				text, err := service.SendPrompt(cmd.Context(), speech, os.Stdout)
				if markdownMode {
					fmt.Print("\n\n---\n\n Markdown : \n\n")
					writer := markdown.NewMarkdownWriter()
					writer.Print(text, os.Stdout)
				}
				if err != nil {
					fmt.Println(err)
				} else {
					speech = text
				}
				fmt.Print("\n\n---\n\n")
			}

			selectionPrompt := promptui.Select{
				Label: "Speech converted to text. What do you want to do with it ?",
				Items: []string{"Copy to clipboard", "Save in file", "quit"},
			}

			id, _, err := selectionPrompt.Run()
			if err != nil {
				fmt.Println(err)
				return
			}

			switch id {
			case 0:
				clipboard.WriteAll(speech)
			case 1:
				filename := ""
			filenameLoop:
				for {
					fmt.Println("Specify the filename orally. If you don't want to specify, press enter twice.")
					fmt.Println("Press enter to record")
					fmt.Scanln()
					filename, err = service.SpeechToText(context.Background(), cmd.Flag("lang").Value.String(), 3*time.Second, false)
					filename = strings.TrimSpace(filename)
					fmt.Printf(" Filename : '%s'\n", filename)
					switch {
					case err != nil:
						fmt.Println(err)
						continue filenameLoop
					case filename == "":
						break filenameLoop
					case ui.YesNoPrompt(fmt.Sprintf("Filename : %s", filename)):
						break filenameLoop
					}
				}
				ui.SaveToFile([]byte(speech), filename)
			case 2:
				os.Exit(0)
			}

			fmt.Print("\n✅\n\n")
		}
	},
}

func init() {
	rootCmd.AddCommand(speechCmd)

	speechCmd.Flags().StringP("lang", "l", "en", "language")
	speechCmd.Flags().StringArrayVarP(&carriageReturnC, "carriage-return", "n", []string{"carriage return"}, "The carriage return character.")
	speechCmd.Flags().BoolVarP(&format, "format", "f", false, "format the output with the carriage return character.")
	speechCmd.Flags().StringVarP(&advancedFormating, "advanced-format", "a", "add markdown formating. Add a title and a table of content from the content of the speech, and add the coresponding subtitles.", "Add advanced formating that will be sent as system command to openai")
	speechCmd.Flags().BoolVarP(&markdownMode, "markdown", "m", false, "Format the output to markdown")
	speechCmd.Flags().StringArrayVarP(&systemOptions, "system", "s", []string{}, "additionnal system options")
	speechCmd.Flags().IntVarP(&maxMinutes, "max-minutes", "t", 5, "max record time (in minutes) (max : 5 minutes)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// speechCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// speechCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
