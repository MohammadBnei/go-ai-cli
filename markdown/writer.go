package markdown

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/samber/lo"
)

type MarkdownWriter struct {
	Buffer    string
	Raw       []string
	Markdown  []byte
	TermWidth int
}

func NewMarkdownWriter() *MarkdownWriter {
	width, _, err := GetTerminalSize()
	if err != nil {
		return &MarkdownWriter{
			TermWidth: 90,
		}
	}
	return &MarkdownWriter{
		TermWidth: width,
	}
}

func (mw *MarkdownWriter) Flush(raw string) {
	if mw.Buffer == "" {
		mw.Markdown = []byte{}
		mw.Raw = []string{}
		return
	}

	mw.Raw = append(mw.Raw, mw.Buffer)
	newMd := markdown.Render(strings.Join(mw.Raw, "\n"), mw.TermWidth-10, 6)
	if raw != "" {
		newMd = markdown.Render(raw, mw.TermWidth-10, 6)
	}
	alter, found := strings.CutPrefix(string(newMd), string(mw.Markdown))
	if found {
		fmt.Print(alter)
	}

	mw.Buffer = ""
	mw.Markdown = []byte{}
	mw.Raw = []string{}
}

var watchThisChar = []string{
	"`",
	"```\\w*",
}

func (mw *MarkdownWriter) Write(p []byte) (n int, err error) {
	mw.Buffer = mw.Buffer + string(p)
	n = len(p)

	for _, c := range watchThisChar {
		regOpen := regexp.MustCompile(c + "[\\s\\S]*?")
		regClose := regexp.MustCompile(c + "[\\s\\S]*?" + c + "")

		if regOpen.MatchString(mw.Buffer) {
			if !regClose.MatchString(mw.Buffer) {
				return
			}
		}
	}

	splitted := strings.Split(mw.Buffer, "\n")
	if len(splitted) > 1 {
		mw.Raw = append(mw.Raw, splitted[:len(splitted)-1]...)
		mw.Buffer, _ = lo.Last(splitted)
		newMd := markdown.Render(strings.Join(mw.Raw, "\n"), mw.TermWidth-10, 6)
		alter, found := strings.CutPrefix(string(newMd), string(mw.Markdown))
		if found {
			fmt.Print(alter)
			mw.Markdown = newMd
		}
	}

	return
}

func (mw *MarkdownWriter) Print(text string, writer io.Writer) error {
	byteText := markdown.Render(text, mw.TermWidth-10, 6)
	_, err := writer.Write(byteText)
	return err
}

func (mw *MarkdownWriter) ToMarkdown(text string) (string, error) {
	byteText := markdown.Render(text, mw.TermWidth-10, 6)
	return string(byteText), nil
}

func GetTerminalSize() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	size := strings.Split(strings.TrimSpace(string(output)), " ")
	width, _ := strconv.Atoi(size[1])
	height, _ := strconv.Atoi(size[0])
	return width, height, nil
}
