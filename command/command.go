// Package command implements the handlers for the PullDeploy sub-commands.
package command

import (
	"github.com/mredivo/pulldeploy/pdconfig"
	"github.com/mredivo/pulldeploy/repo"
	"github.com/mredivo/pulldeploy/storage"
)

// Handler is the interface to which every command handler must conform.
type Handler interface {
	CheckArgs(cmdName string, pdcfg pdconfig.PDConfig, osArgs []string) *Result
	Exec() *Result
}

func getRepoIndex(stg storage.Storage, appName string) (*repo.Index, error) {
	ri := repo.NewIndex(appName)
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

func setRepoIndex(stg storage.Storage, ri *repo.Index) error {
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

func subtractArray(minuend, subtrahend []string) []string {
	var difference []string = []string{}
	for _, s1 := range minuend {
		found := false
		for _, s2 := range subtrahend {
			if s1 == s2 {
				found = true
				break
			}
		}
		if !found {
			difference = append(difference, s1)
		}
	}
	return difference
}
