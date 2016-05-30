package signaller

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mredivo/pulldeploy/pdconfig"
)

// TestSignaller performs the tests twice, with and without Zookeeper.
func TestSignaller(t *testing.T) {
	withZookeeper(t)
	withoutZookeeper(t)
}

func withZookeeper(t *testing.T) {

	// Instantiate and open the Signaller.
	sgnlr := New(&pdconfig.SignallerConfig{
		1,
		5,
		pdconfig.ZookeeperConfig{[]string{"localhost:2181"}, "/pulldeploy"},
	}, nil)
	notifChan := sgnlr.Open()
	defer sgnlr.Close()

	testSignalling(t, sgnlr, notifChan, "with")
}

func withoutZookeeper(t *testing.T) {

	// Instantiate and open the Signaller.
	sgnlr := New(&pdconfig.SignallerConfig{
		1,
		5,
		pdconfig.ZookeeperConfig{[]string{}, ""},
	}, nil)
	notifChan := sgnlr.Open()
	defer sgnlr.Close()

	testSignalling(t, sgnlr, notifChan, "without")
}

func testSignalling(t *testing.T, sgnlr *Signaller, notifChan <-chan Notification, mode string) {

	// The conditions we are checking for.
	var eIsZK, eConnected, eDisconnected, eNotified bool

	// Listen for events.
	var unittestWG sync.WaitGroup
	watchPath := sgnlr.makeAppWatchPath("prod", "myapp")
	sgnlr.Monitor("prod", "myapp")
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
	sgnlr.Notify("prod", "myapp", []byte("My, what big teeth you have, granny!"))

	// Exercise the registry.
	if mode == "with" {
		hr := sgnlr.GetRegistry()
		hr.Register("prod", "myapp", "clienthost-1", "1.1.1")
		hr.Register("prod", "myapp", "clienthost-2", "1.1.1")
		hr.Register("prod", "myapp", "clienthost-3", "1.1.1")
		if count := dumpRegistryInfo(hr.Hosts("prod", "myapp")); count != 3 {
			t.Errorf("Wrong number of registered hosts: have %d, expected %d", count, 3)
		}
		hr.Unregister("prod", "myapp", "clienthost-2")
		if count := dumpRegistryInfo(hr.Hosts("prod", "myapp")); count != 2 {
			t.Errorf("Wrong number of registered hosts: have %d, expected %d", count, 2)
		}
		// Update an entry (twice).
		hr.Register("prod", "myapp", "clienthost-1", "1.1.2")
		if count := dumpRegistryInfo(hr.Hosts("prod", "myapp")); count != 2 {
			t.Errorf("Wrong number of registered hosts: have %d, expected %d", count, 2)
		}
		for _, v := range hr.Hosts("prod", "myapp") {
			if v.Hostname == "clienthost-1" && v.AppVersion != "1.1.2" {
				t.Errorf("Wrong version for registered host %s: have %s, expected %s",
					v.Hostname, v.AppVersion, "1.1.2")
			}
		}
		hr.Register("prod", "myapp", "clienthost-1", "1.1.3")
		if count := dumpRegistryInfo(hr.Hosts("prod", "myapp")); count != 2 {
			t.Errorf("Wrong number of registered hosts: have %d, expected %d", count, 2)
		}
		for _, v := range hr.Hosts("prod", "myapp") {
			if v.Hostname == "clienthost-1" && v.AppVersion != "1.1.3" {
				t.Errorf("Wrong version for registered host %s: have %s, expected %s",
					v.Hostname, v.AppVersion, "1.1.3")
			}
		}
	}

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

func dumpRegistryInfo(ri []RegistryInfo) int {
	count := 0
	fmt.Printf("Registry contents: %v\n", ri)
	for _, v := range ri {
		fmt.Printf("   Host: %q Env: %q App: %q Version: %q\n", v.Hostname, v.Envname, v.Appname, v.AppVersion)
		count++
	}
	return count
}
