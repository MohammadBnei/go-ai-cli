package service

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/c2h5oh/datasize"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"

	"github.com/MohammadBnei/go-ai-cli/config"
	"github.com/MohammadBnei/go-ai-cli/service/godcontext"
)

func AskImage(prompt string, size string) ([]byte, error) {
	c := openai.NewClient(viper.GetString(config.AI_OPENAI_KEY))

	s := spinner.New(spinner.CharSets[26], 100*time.Millisecond)
	s.Start()
	resp, err := c.CreateImage(godcontext.GodContext, openai.ImageRequest{
		Prompt: prompt,
		User:   "user",

		Size:           size,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	})

	s.Stop()

	if err != nil {
		return nil, err

	}

	b, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func EditImage(filePath, prompt, size string) ([]byte, error) {
	c := openai.NewClient(viper.GetString(config.AI_OPENAI_KEY))

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, _ := file.Stat()

	fmt.Println((datasize.ByteSize(info.Size()) * datasize.B).HR())

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	newImg := image.NewNRGBA(img.Bounds())

	// paste PNG image over to newImage
	draw.Draw(newImg, newImg.Bounds(), img, img.Bounds().Min, draw.Src)
	// set the background color to transparent
	draw.DrawMask(newImg, newImg.Bounds(), newImg, image.Point{}, &image.Uniform{color.Transparent}, image.Point{}, draw.Over)

	tmpFileName := time.Now().String() + ".png"
	tmpPng, err := os.Create(tmpFileName)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFileName)
	defer tmpPng.Close()

	err = png.Encode(tmpPng, newImg)
	if err != nil {
		return nil, err
	}

	info, _ = tmpPng.Stat()

	fmt.Println((datasize.ByteSize(info.Size()) * datasize.B).HR(), info.Name())

	s := spinner.New(spinner.CharSets[26], 100*time.Millisecond)
	s.Start()
	resp, err := c.CreateEditImage(godcontext.GodContext, openai.ImageEditRequest{
		Prompt: prompt,

		Image:          tmpPng,
		Size:           size,
		ResponseFormat: openai.CreateImageResponseFormatB64JSON,
		N:              1,
	})

	s.Stop()

	if err != nil {
		return nil, err

	}

	b, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		return nil, err
	}

	return b, nil
}
