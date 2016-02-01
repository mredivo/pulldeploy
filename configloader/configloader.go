// Package configloader loads and manages all application configuration.
package configloader

type SignallerConfig struct {
	ZKServers    []string // Zookeeper servers: host[:port]
	BaseNode     string   // The path to the base node of all Zookeeper nodes
	PollInterval int      // Seconds between repository polls when not using Zookeeper
	PollFallback int      // Seconds between repository polls when Zookeeper is available
}
