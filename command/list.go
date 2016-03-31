package command

import (
	"fmt"
)

// pulldeploy list
type List struct {
}

func (cmd *List) CheckArgs() bool {
	return true
}

func (cmd *List) Exec() {
	fmt.Printf("list()\n")
}
