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
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
	"github.com/rcrowley/go-metrics"
)

const notificationQueueSize = 25

const (
	membershipSizeMetricName  = "cluster.membership.size"
	membershipChurnMetricName = "cluster.membership.churn"
)

// Membership provides access to the current set of members of the cluster.
type Membership interface {

	// Returns the current set of member nodes of the cluster.
	Members() map[Member]struct{}

	// Registers a listener to receive continuous callbacks upon cluster membership changes.
	RegisterListener(l Listener)

	// Unregisters a listener from receiving further callbacks upon cluster membership changes.
	DeregisterListener(l Listener)
}

// Listener receives callbacks upon cluster membership changes.
type Listener interface {

	// Invoked when a member joins a cluster.
	OnJoin(m Member)

	// Invoked when a member leaves a cluster.
	OnLeave(m Member)
}

func newMembership(backend backend, ttl, interval time.Duration) *membership {
	m := &membership{
		backend:     backend,
		cache:       make(map[MemberID]*member),
		ttl:         ttl,
		interval:    interval,
		sizeMetric:  metrics.NewRegisteredGauge(membershipSizeMetricName, metrics.DefaultRegistry),
		churnMetric: metrics.NewRegisteredMeter(membershipChurnMetricName, metrics.DefaultRegistry),
		logger:      logging.GetLogger(module),
	}

	m.sizeMetric.Update(0)

	return m
}

// membership is a Membership implementation that actively
// expires members which fail to renew their registration
type membership struct {

	// Members
	backend backend
	cache   map[MemberID]*member

	// Scheduling
	ttl      time.Duration
	interval time.Duration
	ticker   *time.Ticker

	// Callbacks
	listeners     []Listener
	notifications chan *notification

	// Synchronization
	done       chan struct{}
	stateMutex sync.RWMutex
	startMutex sync.Mutex

	// Metrics
	sizeMetric  metrics.Gauge
	churnMetric metrics.Meter

	logger *logrus.Entry
}

type notification struct {
	callback func(Listener, Member)
	member   Member
}

func (m *membership) Members() map[Member]struct{} {

	m.stateMutex.RLock()
	defer m.stateMutex.RUnlock()

	members := make(map[Member]struct{}, len(m.cache))
	for _, member := range m.cache {
		members[member] = struct{}{}
	}

	return members

}

func (m *membership) RegisterListener(l Listener) {

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	// Copy-On-Write
	newListeners := make([]Listener, len(m.listeners), len(m.listeners)+1)
	copy(newListeners, m.listeners)
	m.listeners = append(newListeners, l)

}

func (m *membership) DeregisterListener(l Listener) {

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	for i, ml := range m.listeners {
		if ml != l {
			continue
		}
		newListeners := make([]Listener, len(m.listeners)-1)
		copy(newListeners, m.listeners[:i])
		copy(newListeners[i:], m.listeners[i+1:])
		m.listeners = newListeners
		break
	}

}

func (m *membership) enqueueNotification(callback func(Listener, Member), member Member) {
	ntf := &notification{
		callback: callback,
		member:   member,
	}
	select {
	case m.notifications <- ntf:
	default:
		m.logger.Warning("Notification queue is full, falling back to blocking send")
		m.notifications <- ntf
	}
}

func (m *membership) deliverNotification(ntf *notification) {
	var listeners []Listener
	func() {
		m.stateMutex.RLock()
		defer m.stateMutex.RUnlock()
		listeners = m.listeners
	}()

	for _, l := range listeners {
		ntf.callback(l, ntf.member)
	}
}

func (m *membership) StartMonitoring() {

	m.startMutex.Lock()
	defer m.startMutex.Unlock()

	// if already monitoring, do nothing
	if m.ticker != nil {
		m.logger.Warning("StartMonitoring() called when already monitoring")
		return
	}

	m.logger.Info("Started monitoring cluster membership")

	m.ticker = time.NewTicker(m.interval)
	m.done = make(chan struct{})
	m.notifications = make(chan *notification, notificationQueueSize)

	go m.monitor()
	go m.notify()
}

func (m *membership) StopMonitoring() {

	m.startMutex.Lock()
	defer m.startMutex.Unlock()

	// if not monitoring, do nothing
	if m.ticker == nil {
		m.logger.Warning("StopMonitoring() called when not monitoring")
		return
	}

	m.logger.Info("Stopped monitoring cluster membership")

	m.ticker.Stop()

	m.done <- struct{}{}
	m.clearCache()
	close(m.notifications)
}

func (m *membership) monitor() {
	for {
		select {
		case <-m.ticker.C:
			joined, left, expired, phantom := m.syncWithBackend()
			m.processChanges(joined, left, expired, phantom)
		case <-m.done:
			return
		}
	}
}

func (m *membership) notify() {
	for ntf := range m.notifications {
		m.deliverNotification(ntf)
	}
}

func (m *membership) syncWithBackend() (joined, left, expired, phantom []Member) {

	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	now := time.Now()
	backendMembers, err := m.backend.ReadMembers()
	if err != nil {
		// Log it, but play on: all members will be regarded as "left"
		m.logger.WithField("error", err).Warning("Error syncing cluster membership - assuming no members exist")
	}

	nMembersPrev := len(m.cache)
	nBackendMembers := len(backendMembers)

	// Preallocate capacity according to max possible len
	joined = make([]Member, 0, nBackendMembers)
	left = make([]Member, 0, nMembersPrev)
	expired = make([]Member, 0, nMembersPrev)
	phantom = make([]Member, 0, nBackendMembers)

	for id, member := range backendMembers {
		if _, exists := m.cache[id]; !exists {
			// Phantom members are expired members who aren't in cache
			if now.Sub(member.Timestamp) > m.ttl {
				phantom = append(phantom, member)
				continue
			}
			joined = append(joined, member)
		}
		m.cache[id] = member
	}

	for id, member := range m.cache {
		if _, exists := backendMembers[id]; !exists {
			left = append(left, member)
			delete(m.cache, id)
		} else if now.Sub(member.Timestamp) > m.ttl {
			expired = append(expired, member)
			delete(m.cache, id)
		}
	}

	nMembersCurr := len(m.cache)
	m.sizeMetric.Update(int64(nMembersCurr))
	if nMembersPrev != nMembersCurr {
		// Do not modify: this log message is searched by the redeploy script
		m.logger.Infof("Number of members in the cluster has been changed. prevSize: %d, currSize: %d", nMembersPrev, nMembersCurr)
	}

	nMembersChanged := len(joined) + len(left) + len(expired)
	if nMembersChanged > 0 {
		m.churnMetric.Mark(int64(nMembersChanged))
	}

	return
}

func (m *membership) processChanges(joined, left, expired, phantom []Member) {
	m.processJoined(joined)
	m.processLeft(left)
	m.processExpired(expired)
	m.processPhantom(phantom)
}

func (m *membership) processJoined(members []Member) {
	for _, member := range members {
		m.logger.Infof("Member joined: %v", member)
		m.enqueueNotification(Listener.OnJoin, member)
	}
}

func (m *membership) processLeft(members []Member) {
	for _, member := range members {
		m.logger.Infof("Member left: %v", member)
		m.enqueueNotification(Listener.OnLeave, member)
	}
}

func (m *membership) processExpired(members []Member) {
	for _, member := range members {
		m.logger.Infof("Member expired: %v", member)
		err := m.backend.DeleteMember(member.ID())
		if err != nil {
			// In case of an error, we'll re-attempt at next timer tick
			m.logger.WithField("error", err).Warningf("Error deleting expired member: %v", member)
		}
		m.enqueueNotification(Listener.OnLeave, member)
	}
}

func (m *membership) processPhantom(members []Member) {
	for _, member := range members {
		m.logger.Debugf("Phantom member found (ignoring): %v", member)
		err := m.backend.DeleteMember(member.ID())
		if err != nil {
			// In case of an error, we'll re-attempt at next timer tick
			m.logger.WithField("error", err).Warningf("Error deleting phantom member: %v", member)
		}
	}
}

func (m *membership) clearCache() {
	for id, member := range m.cache {
		delete(m.cache, id)
		m.enqueueNotification(Listener.OnLeave, member)
	}
}
