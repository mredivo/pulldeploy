package command

import (
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// Handler is the interface to which every command handler must conform.
type Handler interface {
	CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList
	Exec()
}

func NewErrorList(cmdName string) *ErrorList {
	return &ErrorList{cmdName, make([]error, 0)}
}

type ErrorList struct {
	cmdName string
	errs    []error
}

func (el *ErrorList) Append(err error) {
	el.errs = append(el.errs, err)
}

func (el *ErrorList) Errorf(format string, a ...interface{}) {
	el.errs = append(el.errs, fmt.Errorf(format, a...))
}

func (el *ErrorList) Len() int {
	return len(el.errs)
}

func (el *ErrorList) Errors() []string {
	var s []string = []string{}
	for _, err := range el.errs {
		s = append(s, el.cmdName+": "+err.Error())
	}
	return s
}

func placeHolder(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}
