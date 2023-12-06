package ui

import (
	"bufio"
	"fmt"
	"os"
)

func BasicPrompt(label string) (string, error) {
	fmt.Print(label + ": ")
	scanner := bufio.NewScanner(os.Stdin)
	var userPrompt string
	ok := scanner.Scan()
	if err := scanner.Err(); !ok && err != nil {
		return "", err
	}
	userPrompt = scanner.Text()

	return userPrompt, nil
}
