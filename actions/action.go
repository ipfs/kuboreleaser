package actions

type IAction interface {
	Run() error
	Check() error
}
