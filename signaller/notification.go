package signaller

import (
	"fmt"
)

// NotifySource indicate what mechanism triggered the event.
type NotifySource int

// String provides a printable representation of a NotifySource.
func (ns NotifySource) String() string {
	switch ns {
	case KNS_TIMER:
		return "timer"
	case KNS_ZK:
		return "zk"
	default:
		return fmt.Sprintf("unknown(%d)", ns)
	}
}

const (
	KNS_TIMER NotifySource = iota // Event was triggered by a timer
	KNS_ZK                        // Event was triggered by Zookeeper
)

type Notification struct {
	Source  NotifySource // The mechanism that caused the event
	Appname string       // The name of the application that caused the event
	Data    []byte       // Optional data associated with the event
}
