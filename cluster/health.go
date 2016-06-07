package cluster

import (
	"time"

	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/registry/utils/health"
	"github.com/amalgam8/registry/utils/logging"
)

const (
	defaultSubsizeThreshold = 10 * time.Minute
)

// healthChecker is an health.Checker implementation that checks that a cluster size doesn't fall below a threshold.
type healthChecker struct {

	// membership references the healthchecked cluster.Membership.
	membership Membership

	// threshold specifies the minimum cluster size which is regarded as healthy.
	threshold int

	// clusterSize specifies the current recorded cluster size.
	clusterSize int

	// subsizeTimestamp records the timestamp at which the cluster size has fallen below the threshold.
	subsizeTimestamp time.Time

	// subsizeGracePeriod specifies the duration after which a subsized cluster is regarded as unhealthy.
	subsizeGracePeriod time.Duration

	logger *logrus.Entry
	mutex  sync.Mutex
}

func newHealthChecker(membership Membership, threshold int) *healthChecker {
	hc := &healthChecker{
		membership:         membership,
		threshold:          threshold,
		clusterSize:        threshold, // initialize "healthy"
		subsizeTimestamp:   time.Now(),
		subsizeGracePeriod: defaultSubsizeThreshold,
		logger:             logging.GetLogger(module),
	}
	hc.RecordSize()
	membership.RegisterListener(hc)
	return hc
}

func (hc *healthChecker) Check() health.Status {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	props := map[string]interface{}{"size": hc.clusterSize, "threshold": hc.threshold}

	if hc.clusterSize < hc.threshold {
		subsizeDuration := time.Now().Sub(hc.subsizeTimestamp)
		message := fmt.Sprintf("Cluster size (%d) is beneath threshold (%d) for %v", hc.clusterSize, hc.threshold, subsizeDuration)
		if subsizeDuration > hc.subsizeGracePeriod {
			hc.logger.Error(message)
			props["message"] = message
			return health.StatusHealthyWithProperties(props)
		}
		hc.logger.Warning(message)
	}

	return health.StatusHealthyWithProperties(props)
}

func (hc *healthChecker) RecordSize() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	prevSize := hc.clusterSize
	hc.clusterSize = len(hc.membership.Members())

	// Record timestamp for first observed subsize
	if (hc.clusterSize < hc.threshold) && (prevSize >= hc.threshold) {
		hc.subsizeTimestamp = time.Now()
	}
}

// Implement cluster.Listener interface

func (hc *healthChecker) OnJoin(m Member) {
	hc.RecordSize()
}

func (hc *healthChecker) OnLeave(m Member) {
	hc.RecordSize()
}
