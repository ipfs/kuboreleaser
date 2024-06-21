package actions

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/ipfs/kuboreleaser/util"
)

type Env struct {}

//go:embed embed/.env.sh
var envScript string
//go:embed embed/.env.template
var envTemplate string

func (ctx Env) Check() error {
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return fmt.Errorf("file .env does not exist yet in the current directory (%w)", ErrIncomplete)
	}
	return nil
}

func (ctx Env) Run() error {
	envScriptFile, err := os.CreateTemp("", ".env.*.sh")
	if err != nil {
		return err
	}
	_, err = envScriptFile.WriteString(envScript)
	if err != nil {
		return err
	}
	err = os.Chmod(envScriptFile.Name(), 0755)
	if err != nil {
		return err
	}
	envTemplateFile, err := os.CreateTemp("", ".env.*.sh")
	if err != nil {
		return err
	}
	_, err = envTemplateFile.WriteString(envTemplate)
	if err != nil {
		return err
	}

	cmd := util.Command{
		Name: envScriptFile.Name(),
		Stdin: os.Stdin,
		Env: append(os.Environ(), fmt.Sprintf("ENV_TEMPLATE=%s", envTemplateFile.Name())),
	}
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
