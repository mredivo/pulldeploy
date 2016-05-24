/*
Package signaller sends notifications to running PullDeploy daemons.

The signaller is intended to use Zookeeper for synchronization. If Zookeeper
is not available, a timer is used as a fallback to drive polling.
*/
package signaller

import (
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"

	"github.com/mredivo/pulldeploy/logging"
	"github.com/mredivo/pulldeploy/pdconfig"
)

// Signaller is used to notify running daemons of deploy and release activity.
type Signaller struct {
	self      sync.RWMutex             // Mutex to control access to this struct
	cfg       pdconfig.SignallerConfig // The signaller configuration
	wg        sync.WaitGroup           // A waitgroup to monitor lifetime of all goroutines
	quit      chan struct{}            // Closing this channel causes all goroutines to exit
	zkConn    *zk.Conn                 // The connection to Zookeeper, if used, else nil
	zkLogger  zk.Logger                // A go-zookeeper custom logger
	connState chan bool                // The channel on which we watch session state
	appChange chan Notification        // The channel on which we propagate app events
	watches   map[string]interface{}   // A lookup table of all the watched paths
}

type zkLogger struct {
	lw *logging.Writer
}

func (l zkLogger) Printf(format string, a ...interface{}) {
	if l.lw != nil {
		l.lw.Debug(format, a...)
	}
}

// New returns a new Signaller.
func New(cfg *pdconfig.SignallerConfig, lw *logging.Writer) *Signaller {

	// Create object, apply arguments.
	sgnlr := &Signaller{}
	sgnlr.cfg = *cfg

	// Create internal resources.
	sgnlr.quit = make(chan struct{}, 1)
	sgnlr.zkLogger = &zkLogger{lw}
	sgnlr.connState = make(chan bool, 10)
	sgnlr.appChange = make(chan Notification, 100)
	sgnlr.watches = make(map[string]interface{})

	return sgnlr
}

// Open allocates the resources needed for sending and receiving notifications.
func (sgnlr *Signaller) Open() <-chan Notification {

	// No locking: handled separately in each called method.

	// If we have a Zookeeper server list, open a connection and monitor it.
	if len(sgnlr.cfg.ZK.Servers) > 0 {
		// Only do this once.
		if sgnlr.getZKConnWithLock() == nil {
			var connEvent <-chan zk.Event
			connEvent = sgnlr.connectWithLock()
			sgnlr.wg.Add(1)
			go sgnlr.monitorConnection(connEvent)
		}
	}

	return sgnlr.appChange
}

// Close deallocates the resources allocated by Open.
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

// Monitor is used to ask for notifications for the given environment and application.
func (sgnlr *Signaller) Monitor(envName, appName string) {

	sgnlr.self.RLock()
	defer sgnlr.self.RUnlock()

	// Assemble the path for these notifications.
	watchPath := sgnlr.makeAppWatchPath(envName, appName)

	// Do nothing if this path is already being watched.
	if _, found := sgnlr.watches[watchPath]; found {
		return
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

	// Record the watchPath, and start a watcher for it.
	sgnlr.wg.Add(1)
	sgnlr.watches[watchPath] = nil
	go sgnlr.monitorNode(sgnlr.appChange, envName, appName, watchPath, zkEvent, numSeconds)

	return
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

// GetRegistry retrieves an instance of the hosts registry.
func (sgnlr *Signaller) GetRegistry() *Registry {
	var hr = &Registry{sgnlr}
	return hr
}
