package command

type Handler interface {
	CheckArgs(osArgs []string) bool
	Exec()
}
