// Package signaller provides notifications to running PullDeploy daemons.
package signaller

import (
	"sort"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// Signaller is used to notify running daemons of deploy and release activity.
type Signaller struct {
	self      sync.RWMutex                  // Mutex to control access to this struct
	cfg       pdconfig.SignallerConfig      // The signaller configuration
	wg        sync.WaitGroup                // A waitgroup to monitor lifetime of all goroutines
	quit      chan struct{}                 // Closing this channel causes all goroutines to exit
	zkConn    *zk.Conn                      // The connection to Zookeeper, if used, else nil
	connState chan bool                     // The channel on which we propagate session events
	notifiers map[string]*chan Notification // A map of all the notifiers return, keyed by path
}

// NewCient returns a new Signaller.
func NewClient(cfg pdconfig.SignallerConfig) *Signaller {

	// Create object, apply arguments.
	sgnlr := &Signaller{}
	sgnlr.cfg = cfg

	// Create internal resources.
	sgnlr.quit = make(chan struct{}, 1)
	sgnlr.connState = make(chan bool, 10)
	sgnlr.notifiers = make(map[string]*chan Notification)

	return sgnlr
}

// Open allocates the resources needed for generating notifications.
func (sgnlr *Signaller) Open() {

	// No locking: handled separately in each called method.

	// If we have a Zookeeper server list, open a connection and monitor it.
	if len(sgnlr.cfg.ZK.Servers) > 0 {

		// Return immediately if we already have a connection.
		if sgnlr.getZKConnWithLock() != nil {
			return
		}

		// Open a connection, and monitor it.
		var connEvent <-chan zk.Event
		connEvent = sgnlr.connectWithLock()
		sgnlr.wg.Add(1)
		go sgnlr.monitorConnection(connEvent)
	}
}

// Close deallocates resources allocated by Open.
func (sgnlr *Signaller) Close() {

	sgnlr.self.Lock()
	defer sgnlr.self.Unlock()

	// Shut down the watcher/timer goroutine.
	close(sgnlr.quit)
	sgnlr.wg.Wait()

	// Close Zookeeper (if we connected to it).
	if sgnlr.zkConn != nil {
		sgnlr.zkConn.Close()
	}

	sgnlr.zkConn = nil
}

// GetNotificationChannel returns a channel that delivers notifications for the
// given environment and application.
func (sgnlr *Signaller) GetNotificationChannel(envName, appName string) <-chan Notification {

	sgnlr.self.RLock()
	defer sgnlr.self.RUnlock()

	// Assemble the path for these notifications.
	watchPath := sgnlr.makeAppWatchPath(envName, appName)

	// Allow only one notification channel per path.
	if notifChan, found := sgnlr.notifiers[watchPath]; found {
		return *notifChan
	}

	// Set the regular non-Zookeeper polling interval.
	numSeconds := time.Duration(sgnlr.cfg.PollInterval) * time.Second

	// Use polling at longer intervals as a backup when using Zookeeper.
	var zkEvent <-chan zk.Event
	if sgnlr.zkConn != nil {
		numSeconds = time.Duration(sgnlr.cfg.PollFallback) * time.Second
		sgnlr.makeParentNodes(watchPath)
		// If we have Zookeeper, zkEvent will return Zookeeper notifications.
		_, _, zkEvent, _ = sgnlr.zkConn.ExistsW(watchPath)
	} else {
		// If we do not have Zookeeper, supply a dummy channel for zkEvent.
		zkEvent = make(chan zk.Event, 1)
	}

	// Save the notifier under the watchPath, and start a watcher for it.
	sgnlr.wg.Add(1)
	var notifChan chan Notification = make(chan Notification, 1)
	sgnlr.notifiers[watchPath] = &notifChan
	go sgnlr.monitorNode(notifChan, envName, appName, watchPath, zkEvent, numSeconds)

	return notifChan
}

// Notify sends a notication to all listening daemons in the specified environment.
func (sgnlr *Signaller) Notify(envName, appName string, data []byte) {
	if zkConn := sgnlr.getZKConnWithLock(); zkConn != nil {
		flags := int32(zk.FlagEphemeral)
		acl := zk.WorldACL(zk.PermAll)
		watchPath := sgnlr.makeAppWatchPath(envName, appName)
		if _, err := zkConn.Create(watchPath, data, flags, acl); err == nil {
			zkConn.Delete(watchPath, -1)
		}
	}
}

// Register enters the name of the local machine into the hosts registry,
// along with the currently released version (requires Zookeeper).
func (sgnlr *Signaller) Register(envName, appName, hostName, version string) {
	if zkConn := sgnlr.getZKConnWithLock(); zkConn != nil {
		flags := int32(zk.FlagEphemeral)
		acl := zk.WorldACL(zk.PermAll)
		registryPath := sgnlr.makeRegistryPath(envName, appName, hostName)
		sgnlr.makeParentNodes(registryPath)
		zkConn.Create(registryPath, []byte(version), flags, acl)
	}
}

// Unregister removes the name of the local machine from the hosts registry
// (requires Zookeeper).
func (sgnlr *Signaller) Unregister(envName, appName, hostName string) {
	if zkConn := sgnlr.getZKConnWithLock(); zkConn != nil {
		registryPath := sgnlr.makeRegistryPath(envName, appName, hostName)
		zkConn.Delete(registryPath, -1)
	}
}

// GetRegistry retrieves the information in the hosts registry for the given
// environment and application (requires Zookeeper).
func (sgnlr *Signaller) GetRegistry(envName, appName string) []RegistryInfo {

	var ri registryList = make(registryList, 0)

	if zkConn := sgnlr.getZKConnWithLock(); zkConn != nil {
		registryPath := sgnlr.makeRegistryPath(envName, appName, "")
		data := make([]byte, 100)
		hosts, _, _ := zkConn.Children(registryPath)
		for _, host := range hosts {
			data, _, _ = zkConn.Get(registryPath + "/" + host)
			ri = append(ri, RegistryInfo{host, string(data)})
		}
	}
	sort.Sort(ri)

	return ri
}
