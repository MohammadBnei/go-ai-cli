package tool

import (
	"errors"
	"os"
	"strings"
)

func ReadFile(filename string) ([]byte, error) {
	if filename == "" {
		return nil, errors.New("filename cannot be empty")
	}

	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	return os.ReadFile(filename)
}

func SaveToFile(content []byte, filename string, append bool) error {
	if filename == "" {
		return errors.New("filename cannot be empty")
	}

	if strings.Contains(filename, "/") {
		splitted := strings.Split(filename, "/")
		dw := strings.Join(splitted[:len(splitted)-1], "/")

		_, err := os.Stat(dw)

		if errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dw, os.ModePerm)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

	}

	if append {
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.WriteString(string(content)); err != nil {
			return err
		}

		return nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	f.Write(content)

	return nil

}
