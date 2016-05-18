package logging

import (
	"fmt"
	"os"
)

// FatalError is a custom error that can be distinguished in the global panic handler.
type FatalError struct {
	msg string
}

func (e *FatalError) Error() string {
	return fmt.Sprintf("FATAL %s", e.msg)
}

/*
OnPanic should be deferred at the beginning of every goroutine that uses logger.Fatal().

If logger.Fatal() is called, OnPanic will log the message to stderr and
exit the application. All other panics are forwarded unchanged.
*/
func OnPanic() {
	if err := recover(); err != nil {
		switch err.(type) {
		case FatalError:
			// This is an error that we generated to terminate the program.
			e := err.(FatalError)
			fmt.Fprintf(os.Stderr, "%s\n", e.Error())
			os.Exit(1) // Let OS know we aborted
		default:
			// This is an error due to a bug; print full details and terminate.
			// Note: panic() writes to stderr.
			panic(err)
		}
	}
}
