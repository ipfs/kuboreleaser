package util

type CheckErrorAction int

const (
	CheckErrorWait CheckErrorAction = iota
	CheckErrorRetry
	CheckErrorFail
)

type CheckError struct {
	Action CheckErrorAction
	Err    error
}

func (e *CheckError) Error() string {
	return e.Err.Error()
}
