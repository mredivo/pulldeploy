/*
Package logging provides logger output.

Features:

	* Can output to either a file, or to stderr
	* Supports multiple writers with different logging levels
	* Supports log rotation
	* Supports clean application termination via logger.Fatal()
	* Provides a panic handler that does special handling when logger.Fatal() is used

Example:

	package main

	import (
		"time"

		"github.com/mredivo/pulldeploy/logging"
	)

	var logger *logging.Logger

	func myFunc() {
		defer logging.OnPanic()
		lw := logger.GetWriter("myFunc", "debug")
		lw.Info("myFunc started")
	}

	func main() {
		defer logging.OnPanic()

		logger = logging.New("demo", "", true)
		defer logger.Close()

		go myFunc()

		lw := logger.GetWriter("", "debug")
		lw.Info("Program started")

		<-time.After(time.Second)

		logger.OnRotate()
		lw.Info("Program ended")
	}

Output:

	2016/05/18 10:18:58.836640 [demo] [INFO] Log opened
	2016/05/18 10:18:58.836828 [demo] [INFO] Program started
	2016/05/18 10:18:58.836841 [demo] [myFunc] [INFO] myFunc started
	2016/05/18 10:18:59.841726 [demo] [INFO] Log closed on signal
	2016/05/18 10:18:59.841777 [demo] [INFO] Log reopened on signal
	2016/05/18 10:18:59.841791 [demo] [INFO] Program ended
	2016/05/18 10:18:59.841802 [demo] [INFO] Log closed

*/
package logging

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type logAction int

// The actions that the logManager() goroutine can act upon.
const (
	kACTION_OPEN logAction = iota
	kACTION_ROTATE
	kACTION_CLOSE
	kACTION_PRINT
)

// The commands and data sent to the logManager() goroutine.
type logCommand struct {
	cmd  logAction // The action to be taken
	text string    // The text to be written to the logger
}

type Logger struct {
	cmd           chan logCommand // The channel for communicating log commands to manager
	identifier    string          // The identifying string written to every log entry
	printSeverity bool            // Whether to print the message severity in the log entries
	wg            *sync.WaitGroup
}

/*
New returns a newly instantiated and initialized logger.

Parameters:

	* identifier    A string to be written at the beginning of each log entry
	* filename      The full path to the log file; when blank, uses stderr
	* printSeverity When true, prints the message severity with each log entry

The Close() method should be deferred after calling New() to ensure proper cleanup.
*/
func New(identifier, filename string, printSeverity bool) *Logger {

	// This may need to be tunable by the client.
	var chanLen int
	if chanLen < 20 {
		chanLen = 20
	}

	l := new(Logger)
	l.identifier = identifier
	l.printSeverity = printSeverity

	// Create the communication channel, and start the logfile manager.
	l.cmd = make(chan logCommand, chanLen)
	go l.logManager(filename)
	l.cmd <- logCommand{cmd: kACTION_OPEN}

	return l
}

// Close releases all resources associated with the logger, and ensures all
// logged entries are flushed.
func (l *Logger) Close() {
	// Kick off the close, and wait for it to finish.
	l.wg = new(sync.WaitGroup)
	l.wg.Add(1)
	l.cmd <- logCommand{cmd: kACTION_CLOSE}
	l.wg.Wait()
}

// OnRotate closes and re-opens the log file.
// This should be called whenever the logs have been rotated.
func (l *Logger) OnRotate() {
	l.cmd <- logCommand{cmd: kACTION_ROTATE}
}

/*
GetWriter gets a logging writer with its own log level and optional log identifier.

This allows logging at different log levels in different parts of the
application. It also allows putting different identifiers onto the log
lines from different parts of the application.
*/
func (l *Logger) GetWriter(writerName string, logLevel string) *Writer {
	var writer *Writer
	writer = &Writer{writerName, textToSeverity(logLevel), l}
	return writer
}

/*
Fatal logs a message and terminates the program.

This is functionally identical to calling writer.Fatal(), and is available
for convenience.
*/
func (l *Logger) Fatal(format string, a ...interface{}) {
	l.fatal("", format, a...)
}

// debug writes a time-stamped Debug message to the log file.
func (l *Logger) debug(writerName, format string, a ...interface{}) {
	l.emit(kDEBUG, writerName, format, a...)
}

// info writes a time-stamped Info message to the log file.
func (l *Logger) info(writerName, format string, a ...interface{}) {
	l.emit(kINFO, writerName, format, a...)
}

// warn writes a time-stamped Warn message to the log file.
func (l *Logger) warn(writerName, format string, a ...interface{}) {
	l.emit(kWARN, writerName, format, a...)
}

// error writes a time-stamped Error message to the log file.
func (l *Logger) error(writerName, format string, a ...interface{}) {
	l.emit(kERROR, writerName, format, a...)
}

// fatal writes a time-stamped Fatal message to the log file, and panics.
func (l *Logger) fatal(writerName, format string, a ...interface{}) {
	l.emit(kFATAL, writerName, format, a...)
	panic(FatalError{fmt.Sprintf(format, a...)})
}

// emit produces a log line.
func (l *Logger) emit(sev severity, writerName, format string, a ...interface{}) {
	l.cmd <- logCommand{cmd: kACTION_PRINT, text: l.format(sev, writerName, format, a...)}
}

// format assembles all the details into a log line.
func (l *Logger) format(sev severity, writerName, format string, a ...interface{}) string {
	prefix := "[" + l.identifier + "] "
	if writerName != "" {
		prefix += "[" + writerName + "] "
	}
	if l.printSeverity {
		prefix += "[" + severityToText(sev) + "] "
	}
	return fmt.Sprintf(prefix+format, a...)
}

// logManager manages the log file and writes log entries into it.
func (l *Logger) logManager(filename string) {

	// mustOpen opens a file for appending, or dies with a fatal error.
	var mustOpen = func(filename string) *os.File {
		if fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err == nil {
			return fp
		} else {
			// Inability to log is a fatal error. We do not run blind.
			l.Fatal("Unable to reopen logfile '%v'. Error: '%s'", filename, err.Error())
			return nil // NOTREACHED
		}
	}

	var opened bool        // Whether the logger has been opened or not
	var logFile *os.File   // The file handle of the opened file
	var logger *log.Logger // The logger that writes to the file

	// Read incoming commands and process them until kACTION_CLOSE received.
	for {

		// Wait for incoming commands and log entries.
		cmd := <-l.cmd

		switch cmd.cmd {

		case kACTION_OPEN:
			if !opened {
				// Initial open of the log file.
				if filename != "" {
					logFile = mustOpen(filename)
				} else {
					logFile = os.Stderr
				}
				logger = log.New(logFile, "", log.Ldate|log.Lmicroseconds)
				logger.Println(l.format(kINFO, "", "Log opened"))
				opened = true
			}

		case kACTION_ROTATE:
			if opened {
				// Close log file, and re-open with the same name.
				logger.Println(l.format(kINFO, "", "Log closed on signal"))
				if filename != "" {
					logFile.Close()
					logFile = mustOpen(filename)
					logger = log.New(logFile, "", log.Ldate|log.Lmicroseconds)
				}
				logger.Println(l.format(kINFO, "", "Log reopened on signal"))
			}

		case kACTION_CLOSE:
			if opened {
				// Close the log file.
				logger.Println(l.format(kINFO, "", "Log closed"))
				if filename != "" {
					logFile.Close()
				}
				opened = false
			}
			// Exit the goroutine.
			l.wg.Done()
			return

		case kACTION_PRINT:
			if opened {
				logger.Println(cmd.text)
			}
		}
	}
}
