package replication

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"

	"fmt"

	"github.com/amalgam8/registry/cluster"
	"github.com/amalgam8/registry/utils/health"
	"github.com/amalgam8/registry/utils/logging"
)

const (
	defaultDisconnectedThreshold = 10 * time.Minute
)

// healthChecker is an health.Checker implementation that checks that replication connections are properly established.
type healthChecker struct {
	clients               map[cluster.MemberID]*clientHealth
	disconnectedThreshold time.Duration
	logger                *logrus.Entry
	mutex                 sync.Mutex
}

func newHealthChecker() *healthChecker {
	return &healthChecker{
		clients:               make(map[cluster.MemberID]*clientHealth),
		disconnectedThreshold: defaultDisconnectedThreshold,
		logger:                logging.GetLogger(module),
	}
}

// clientHealth records health info about a specific replication client
type clientHealth struct {
	cl               clientConnection
	connected        bool
	disconnectedTime time.Time
}

func (hc *healthChecker) AddClient(id cluster.MemberID, cl clientConnection) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	clh := &clientHealth{
		cl:               cl,
		connected:        cl.isConnected(),
		disconnectedTime: time.Now(),
	}

	hc.clients[id] = clh
}

func (hc *healthChecker) RemoveClient(id cluster.MemberID) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	delete(hc.clients, id)
}

func (hc *healthChecker) Check() health.Status {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	now := time.Now()
	numDisconnected := 0
	minDuration := time.Duration(0)

	for id, health := range hc.clients {

		wasConnected := health.connected
		health.connected = health.cl.isConnected()

		if !health.connected {
			if wasConnected {
				health.disconnectedTime = now
			}
			disconnectedDuration := now.Sub(health.disconnectedTime)
			if disconnectedDuration > hc.disconnectedThreshold {
				hc.logger.Warningf("Peer %v replication client disconnected for %v", id, disconnectedDuration)
				numDisconnected++
				if numDisconnected == 1 || (minDuration > disconnectedDuration) {
					minDuration = disconnectedDuration
				}
			}
		}
	}

	if numDisconnected > 0 {
		message := fmt.Sprintf(
			"%d/%d replication clients disconnected for at least %v",
			numDisconnected, len(hc.clients), minDuration)
		return health.StatusUnhealthy(message, nil)
	}
	return health.Healthy

}
