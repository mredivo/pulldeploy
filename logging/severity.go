package logging

import (
	"strings"
)

// The type for logging level.
type severity int

// The valid logging severities.
const (
	kUNCLASSIFIED severity = iota
	kDEBUG
	kINFO
	kWARN
	kERROR
	kFATAL
)

// The text representations of the logging severities.
var severityText = map[severity]string{
	kUNCLASSIFIED: "",
	kDEBUG:        "DEBUG",
	kINFO:         "INFO",
	kWARN:         "WARN",
	kERROR:        "ERROR",
	kFATAL:        "FATAL",
}

// severityToText maps the severity value to a name for printing.
func severityToText(sev severity) string {
	s, ok := severityText[sev]
	if !ok {
		s = "UNKNOWN"
	}
	return s
}

// textToSeverity derives a severity from a text string, ignoring case.
func textToSeverity(s string) severity {
	var sev severity
	switch strings.ToLower(s) {
	case "":
		sev = kUNCLASSIFIED
	case "unclassified":
		sev = kUNCLASSIFIED
	case "debug":
		sev = kDEBUG
	case "info":
		sev = kINFO
	case "warn":
		sev = kWARN
	case "error":
		sev = kERROR
	case "fatal":
		sev = kFATAL
	default:
		sev = kINFO
	}
	return sev
}
