// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package cluster

import (
	"time"

	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/utils/health"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
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
