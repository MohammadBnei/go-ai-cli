package service

import (
	"errors"
	"fmt"

	"github.com/TannerKvarfordt/hfapigo"
	"github.com/spf13/viper"
)

func Mask(prompt string) (string, error) {
	apiKey := viper.GetString("HUGGING_KEY") // Your Hugging Face API key
	if apiKey == "" {
		return "", errors.New("Hugging Face API key not found")
	}
	hfapigo.SetAPIKey(apiKey)

	fmt.Println(prompt)

	fmresps, err := hfapigo.SendFillMaskRequest("xlm-roberta-large", &hfapigo.FillMaskRequest{
		Inputs:  []string{prompt},
		Options: *hfapigo.NewOptions().SetWaitForModel(true),
	})
	if err != nil {
		return "", err
	}

	for _, fmresp := range fmresps {
		for _, mask := range fmresp.Masks {
			fmt.Println(mask.Sequence)
		}
	}

	return fmresps[0].Masks[0].Sequence, nil

}
