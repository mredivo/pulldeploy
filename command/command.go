// Package command implements the handlers for the PullDeploy sub-commands.
package command

import (
	"fmt"

	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/repo"
	"github.com/mredivo/pulldeploy/storage"
)

// Handler is the interface to which every command handler must conform.
type Handler interface {
	CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *ErrorList
	Exec() *ErrorList
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

func getRepoIndex(stg storage.Storage, appName string) (*repo.RepoIndex, error) {
	ri := repo.NewRepoIndex(appName)
	if text, err := stg.Get(ri.IndexPath()); err == nil {
		if err := ri.FromJSON(text); err == nil {
			return ri, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func setRepoIndex(stg storage.Storage, ri *repo.RepoIndex) error {
	ri.Canary++
	if text, err := ri.ToJSON(); err == nil {
		if err := stg.Put(ri.IndexPath(), text); err == nil {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func placeHolder(format string, a ...interface{}) {
	fmt.Printf("NOTIMPLEMENTED: "+format, a...)
}
