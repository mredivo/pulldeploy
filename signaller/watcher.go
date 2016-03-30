package signaller

import (
	"path"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

// connectWithLock wraps the connect funtionality with a mutex.
func (sgnlr *Signaller) connectWithLock() <-chan zk.Event {

	sgnlr.self.Lock()
	defer sgnlr.self.Unlock()

	if sgnlr.zkConn != nil {
		sgnlr.zkConn.Close()
		sgnlr.zkConn = nil
	}

	var connEvent <-chan zk.Event
	sgnlr.zkConn, connEvent, _ = zk.Connect(sgnlr.cfg.ZK.Servers, time.Second)

	return connEvent
}

// getZKConnWithLock wraps retrieving the Zookeeper connection with a mutex.
func (sgnlr *Signaller) getZKConnWithLock() *zk.Conn {

	sgnlr.self.RLock()
	defer sgnlr.self.RUnlock()

	return sgnlr.zkConn
}

// makeAppWatchPath builds the Zookeeper path corresponding to the name env and app.
//   /<base>/<env>/changed/<app>
func (sgnlr *Signaller) makeAppWatchPath(envName, appName string) string {
	return path.Join(sgnlr.cfg.ZK.BaseNode, envName, "changed", appName)
}

// makeParentNodes ensures all leading elements of the supplied path are present.
func (sgnlr *Signaller) makeParentNodes(watchPath string) {

	// Check each segment but the last, creating a permanent node as necessary.
	segs := strings.Split(watchPath, "/")
	newPath := ""
	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)
	for _, k := range segs[1 : len(segs)-1] {
		newPath += "/" + k
		if found, _, _ := sgnlr.zkConn.Exists(newPath); !found {
			sgnlr.zkConn.Create(newPath, []byte{}, flags, acl)
		}
	}
}

// monitorConnection monitors the state of the Zookeeper connection.
func (sgnlr *Signaller) monitorConnection(connEvent <-chan zk.Event) {

	var inSession bool

	for {
		select {
		case <-sgnlr.quit:
			sgnlr.wg.Done()
			return
		case e := <-connEvent:
			if e.Type == zk.EventSession {
				send := false
				switch e.State {
				case zk.StateConnecting:
					//logger.Sink.Log("Zookeeper attempting to connect")
					inSession = false
				case zk.StateConnected:
					//logger.Sink.Log("Zookeeper connected")
				case zk.StateHasSession:
					//logger.Sink.Log("Zookeeper session started")
					inSession = true
					send = true
				case zk.StateDisconnected:
					//logger.Sink.Log("Zookeeper connection lost")
					inSession = false
					send = true
				case zk.StateExpired:
					//logger.Sink.Log("Zookeeper session expired")
					inSession = false
					send = true
					// Expiry is a special case; recovery is not automatic.
					connEvent = sgnlr.connectWithLock()
				}
				if send {
					// Non-blocking channel write.
					select {
					case sgnlr.connState <- inSession:
					default:
					}
				}
			}
		}
	}
}

// monitorNode monitors the state of a particular Zookeeper node.
func (sgnlr *Signaller) monitorNode(
	notifChan chan Notification,
	envName, appName, watchPath string,
	zkEvent <-chan zk.Event,
	numSeconds time.Duration,
) {
	for {
		select {
		case <-sgnlr.quit:
			sgnlr.wg.Done()
			return
		case e := <-zkEvent:
			if e.Type == zk.EventNodeCreated {
				data, _, _ := sgnlr.zkConn.Get(e.Path)
				notifChan <- Notification{KNS_ZK, appName, data}
			}
			_, _, zkEvent, _ = sgnlr.zkConn.ExistsW(watchPath)
		case <-time.After(numSeconds):
			notifChan <- Notification{KNS_TIMER, appName, []byte{}}
		}
	}
}
