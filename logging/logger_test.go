package logging

import (
	"bufio"
	"os"
	"sync"
	"testing"
	"time"
)

func testLoggerWriters(t *testing.T) {

	logger := New("TestLoggerWriters", "", true)
	defer logger.Close()

	lw1 := logger.GetWriter("foo", severityToText(kINFO))

	lw1.Debug("debug message")
	lw1.Info("info message")
	lw1.Warn("warn message")
	lw1.Error("error message")

	lw2 := logger.GetWriter("bar", severityToText(kWARN))

	lw2.Debug("debug message")
	lw2.Info("info message")
	lw2.Warn("warn message")
	lw2.Error("error message")

	lw3 := logger.GetWriter("", severityToText(kERROR))

	lw3.Debug("debug message")
	lw3.Info("info message")
	lw3.Warn("warn message")
	lw3.Error("error message")

	lw1.Error("error message")
	lw2.Error("error message")
	lw3.Error("error message")

	lw4 := logger.GetWriter("", severityToText(kERROR))
	lw4.Error("error message")

	lw5 := logger.GetWriter("foo", severityToText(kERROR))
	lw5.Error("error message")
}

func TestLoggerWriters(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(1)

	// Put body of tests into sub-function so deferred close() happens.
	testLoggerWriters(t)

	go func() {
		<-time.After(time.Second)
		wg.Done()
	}()

	wg.Wait()
}

func testLogger(t *testing.T, logfile string) {

	// Instantiate a logger.
	logger := New("TestLogger", logfile, true)
	defer logger.Close()

	lw := logger.GetWriter("", severityToText(kDEBUG))
	logger.emit(kFATAL, "", "Logging at level DEBUG")
	lw.Debug("Hello %s!", "debug")
	lw.Info("Hello %s!", "info")
	lw.Warn("Hello %s!", "warn")
	lw.Error("Hello %s!", "error")

	lw = logger.GetWriter("", severityToText(kINFO))
	logger.emit(kFATAL, "", "Logging at level INFO")
	lw.Debug("Hello %s!", "debug")
	lw.Info("Hello %s!", "info")
	lw.Warn("Hello %s!", "warn")
	lw.Error("Hello %s!", "error")

	lw = logger.GetWriter("", severityToText(kWARN))
	logger.emit(kFATAL, "", "Logging at level WARN")
	lw.Debug("Hello %s!", "debug")
	lw.Info("Hello %s!", "info")
	lw.Warn("Hello %s!", "warn")
	lw.Error("Hello %s!", "error")

	logger.OnRotate()

	lw = logger.GetWriter("", severityToText(kERROR))
	logger.emit(kFATAL, "", "Logging at level ERROR")
	lw.Debug("Hello %s!", "debug")
	lw.Info("Hello %s!", "info")
	lw.Warn("Hello %s!", "warn")
	lw.Error("Hello %s!", "error")

	lw = logger.GetWriter("", severityToText(kFATAL))
	logger.emit(kFATAL, "", "Logging at level FATAL")
	lw.Debug("Hello %s!", "debug")
	lw.Info("Hello %s!", "info")
	lw.Warn("Hello %s!", "warn")
	lw.Error("Hello %s!", "error")

	lw = logger.GetWriter("", severityToText(kDEBUG))
	lw.Debug("Hello %s!", "debug")
	lw.Info("Hello %s!", "info")
	lw.Warn("Hello %s!", "warn")
	lw.Error("Hello %s!", "error")
}

func TestLogger(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(1)

	const LOGFILE string = "./test.log"

	// Clean up the results of any failed prior test.
	os.Remove(LOGFILE)

	// Put body of tests into sub-function so deferred close() happens.
	testLogger(t, LOGFILE)

	go func() {
		<-time.After(time.Second)
		wg.Done()
	}()

	// Wait for the run() goroutine to process its messages.
	wg.Wait()

	// Validate the results by reading the logfile.
	var lines []string
	var expected = 23 // 19 let through above, plus 4 for opened/closed/reopened messages

	fp, err := os.Open(LOGFILE)
	if err != nil {
		t.Errorf("Unable to open test log file: %s", err.Error())
	}
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(lines) != expected {
		// Also keep the output file for inspection.
		t.Errorf("Wrong number of log lines were emitted. Found %d, expected %d", len(lines), expected)
	} else {
		os.Remove(LOGFILE)
	}
}
