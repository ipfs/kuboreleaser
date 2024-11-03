package util

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func Getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetenvBool(key string) bool {
	return strings.ToLower(Getenv(key, "false")) == "true"
}

func GetenvPrompt(key string, prompt ...string) string {
	value := Getenv(key, "")
	if value == "" {
		if len(prompt) > 0 {
			fmt.Printf("🙋 %s is not set. %s: ", key, prompt[0])
		} else {
			fmt.Printf("🙋 %s is not set. Please enter a value: ", key)
		}
		fmt.Scanln(&value)
	}
	return value
}

func GetenvPromptSecret(key string, prompt ...string) string {
	value := Getenv(key, "")
	if value == "" {
		if len(prompt) > 0 {
			fmt.Printf("🙋 %s is not set. %s: ", key, prompt[0])
		} else {
			fmt.Printf("🙋 %s is not set. Please enter a secret value: ", key)
		}
		bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		fmt.Println()
		value = string(bytes)
	}
	return value
}
