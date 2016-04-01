package command

import (
	"github.com/mredivo/pulldeploy/pdconfig"
)

type Handler interface {
	CheckArgs(pdcfg pdconfig.PDConfig, osArgs []string) bool
	Exec()
}
