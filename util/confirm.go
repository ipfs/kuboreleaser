package util

import "fmt"

func Confirm(prompt string) bool {
	var confirmation string
	fmt.Printf(`%s

Only 'yes' will be accepted to approve.

Enter a value: `, prompt)
	fmt.Scanln(&confirmation)
	return confirmation == "yes"
}
