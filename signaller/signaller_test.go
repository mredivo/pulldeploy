package signaller

import (
	"fmt"
	"sync"
	"testing"
	"time"

	config "github.com/mredivo/pulldeploy/configloader"
)

// TestSignaller performs the tests twice, with and without Zookeeper.
func TestSignaller(t *testing.T) {
	withZookeeper(t)
	withoutZookeeper(t)
}

func withZookeeper(t *testing.T) {

	// Instantiate and open the Signaller.
	sgnlr := NewClient(config.SignallerConfig{[]string{"localhost:2181"}, "/pulldeploy", 1, 5})
	sgnlr.Open()
	defer sgnlr.Close()

	testSignalling(t, sgnlr)
}

func withoutZookeeper(t *testing.T) {

	// Instantiate and open the Signaller.
	sgnlr := NewClient(config.SignallerConfig{[]string{}, "/pulldeploy", 1, 5})
	sgnlr.Open()
	defer sgnlr.Close()

	testSignalling(t, sgnlr)
}

func testSignalling(t *testing.T, sgnlr *Signaller) {

	// The conditions we are checking for.
	var eIsZK, eConnected, eDisconnected, eNotified bool

	// Listen for events.
	var unittestWG sync.WaitGroup
	watchPath := sgnlr.makeAppWatchPath("prod", "myapp")
	notifChan := sgnlr.GetNotificationChannel("prod", "myapp")
	unittestWG.Add(1)
	var unittestMutex sync.Mutex // Prevent races in the unit test itself
	go func() {
		for {
			select {
			case connected := <-sgnlr.connState:
				fmt.Printf("Got a connection event: connected=%v\n", connected)
				if connected {
					eConnected = true
				} else {
					eDisconnected = true
				}
			case ns := <-notifChan:
				fmt.Printf("Got a change notification: %q\n", ns)
				unittestMutex.Lock()
				if ns.Source == KNS_ZK {
					eIsZK = true
				}
				eNotified = true
				unittestMutex.Unlock()
				unittestWG.Done()
				return
			case <-time.After(time.Second * 10):
				fmt.Printf("TIMEOUT!\n")
				unittestWG.Done()
				return
			}
		}
	}()

	// Send a notification.
	sgnlr.Notify("prod", "myapp")

	// Exercise the registry.
	sgnlr.Register("prod", "myapp", "myhostname1", "1.1.1")
	sgnlr.Register("prod", "myapp", "myhostname2", "1.1.1")
	sgnlr.Register("prod", "myapp", "myhostname3", "1.1.1")
	fmt.Println(sgnlr.GetRegistry("prod", "myapp"))
	sgnlr.Unregister("prod", "myapp", "myhostname2")
	fmt.Println(sgnlr.GetRegistry("prod", "myapp"))

	// Block until all messages arrive, or there's a timeout.
	unittestWG.Wait()

	// Guard access to the variables affected by the goroutine.
	unittestMutex.Lock()

	// Check whether we received every event.
	if eIsZK && !eConnected {
		t.Errorf("Did not received 'Connected' event")
	}
	if !eDisconnected {
		// Disconnect happens after this function returns, because it was deferred.
		//t.Errorf("Did not received 'Disconnected' event")
	}
	if !eNotified {
		t.Errorf("Did not received notification event for %q", watchPath)
	}

	unittestMutex.Unlock()
}
