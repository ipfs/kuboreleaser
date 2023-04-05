package util

import (
	"fmt"

	"github.com/google/go-github/v48/github"
)

func Confirm(prompt string) bool {
	var confirmation string
	fmt.Printf(`%s

Only 'yes' will be accepted to approve.

Enter a value: `, prompt)
	fmt.Scanln(&confirmation)
	return confirmation == "yes"
}

func ConfirmPR(pr *github.PullRequest) bool {
	prompt := fmt.Sprintf(`Go to %s, ensure the CI checks pass, and merge the PR. Use merge commit if possible.

Please approve once the PR is merged.`, pr.GetHTMLURL())
	return Confirm(prompt)
}
