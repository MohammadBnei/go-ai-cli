package service

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/viper"
)

func GetConfig() (map[string]any, error) {
	return viper.AllSettings(), nil
}

func SetConfig(key string, value any) error {
	if !lo.Some[string](viper.AllKeys(), []string{strings.ToLower(key)}) {
		return fmt.Errorf("key %s not found", key)
	}
	viper.Set(key, value)
	return nil
}

func SaveConfigToFile() error {
	return viper.WriteConfig()
}
