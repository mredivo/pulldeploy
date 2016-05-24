package command

import (
	"fmt"
)

func NewResult(cmdName string) *Result {
	return &Result{cmdName, "", make([]error, 0)}
}

type Result struct {
	cmdName string
	msg     string
	errs    []error
}

func (result *Result) Messagef(format string, a ...interface{}) {
	result.msg = fmt.Sprintf(format, a...)
}

func (result *Result) Message() string {
	return result.msg
}

func (result *Result) AppendError(err error) {
	result.errs = append(result.errs, err)
}

func (result *Result) Errorf(format string, a ...interface{}) {
	result.errs = append(result.errs, fmt.Errorf(format, a...))
}

func (result *Result) ErrorCount() int {
	return len(result.errs)
}

func (result *Result) Errors() []string {
	var s []string = []string{}
	for _, err := range result.errs {
		s = append(s, result.cmdName+": "+err.Error())
	}
	return s
}
