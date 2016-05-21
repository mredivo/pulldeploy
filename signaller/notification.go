package signaller

import (
	"fmt"
)

// NotifySource indicates what mechanism triggered the event.
type NotifySource int

// String provides a printable representation of a NotifySource.
func (ns NotifySource) String() string {
	switch ns {
	case KNS_FORCED:
		return "forced"
	case KNS_TIMER:
		return "timer"
	case KNS_ZK:
		return "zk"
	default:
		return fmt.Sprintf("unknown(%d)", ns)
	}
}

// The values that may appear as the Source in a Notification.
const (
	KNS_FORCED NotifySource = iota // Notification was created externally to signaller
	KNS_TIMER                      // Notification was triggered by a timer
	KNS_ZK                         // Notification was triggered by Zookeeper
)

/*
A Notification identifies which monitored application has changed.
They are obtained from the channel returned by Open().
*/
type Notification struct {
	Source  NotifySource // The mechanism that caused the event
	Appname string       // The name of the application that caused the event
	Data    []byte       // Optional data associated with the event
}
