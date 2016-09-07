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
	"errors"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

// Registrator enables a member node to join and leave the cluster
type Registrator interface {

	// Self returns the associated member node.
	Self() Member

	// Join adds the associated member node to the cluster.
	// Returns a non-nil error if and only if the member node could not be added to the cluster.
	Join() error

	// Leave removes the associated member node from the cluster.
	// Returns a non-nil error if and only if the member node could not be removed from the cluster.
	Leave() error
}

func newRegistrator(backend backend, member Member, membership Membership, interval time.Duration) *registrator {
	reg := &registrator{
		backend:    backend,
		member:     member,
		membership: membership,
		interval:   interval,
		rejoin:     make(chan struct{}),
		done:       make(chan struct{}),
		logger:     logging.GetLogger(module),
	}
	reg.listener = &ongoingListener{reg: reg}
	return reg
}

// registration is a Registration implementation that actively rejoins
// the cluster as long as not explicitly stopped.
type registrator struct {
	backend    backend
	member     Member
	membership Membership
	interval   time.Duration
	rejoin     chan struct{}
	done       chan struct{}
	renewer    *time.Ticker
	listener   *ongoingListener
	mutex      sync.Mutex
	logger     *logrus.Entry
}

func (r *registrator) Self() Member {
	return r.member
}

func (r *registrator) Join() error {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.renewer != nil {
		err := errors.New("Already joined")
		r.logger.WithField("error", err).Warning()
		return err
	}

	r.membership.RegisterListener(r.listener)
	r.renewer = time.NewTicker(r.interval)
	go r.maintainRegistration(r.renewer.C, r.rejoin, r.done)
	r.writeSelf()

	return nil

}

func (r *registrator) Leave() error {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.renewer == nil {
		err := errors.New("Already left or never joined")
		r.logger.WithField("error", err).Warning()
		return err
	}

	r.membership.DeregisterListener(r.listener)
	r.renewer.Stop()
	r.renewer = nil
	r.done <- struct{}{}
	return r.deleteSelf()

}

func (r *registrator) maintainRegistration(tick <-chan time.Time, left <-chan struct{}, done <-chan struct{}) {
	for {
		select {
		// race condition: Leave() and then deleteSelf()
		// might be called just prior to calling writeSelf().
		// this is both rare and self-healing, so not guarded
		case <-tick:
			r.writeSelf()
		case <-left:
			r.writeSelf()
		case <-done:
			return
		}
	}
}

func (r *registrator) writeSelf() {
	self := &member{
		MemberIP:   r.member.IP(),
		MemberPort: r.member.Port(),
		Timestamp:  time.Now(),
	}
	err := r.backend.WriteMember(self)
	if err != nil {
		_, joined := r.membership.Members()[r.member]
		if joined {
			r.logger.WithField("error", err).Warningf("Error renewing member %v", self.ID())
		} else {
			r.logger.WithField("error", err).Warningf("Error joining member %v", self.ID())
		}
	}
	return
}

func (r *registrator) deleteSelf() error {
	err := r.backend.DeleteMember(r.member.ID())
	if err != nil {
		r.logger.WithField("error", err).Warningf("Error leaving member %v", r.member.ID())
	}
	return err
}

type ongoingListener struct {
	reg *registrator
}

func (l *ongoingListener) OnJoin(m Member) {
	// Do nothing
}

func (l *ongoingListener) OnLeave(m Member) {
	if m.ID() == l.reg.member.ID() {
		l.reg.logger.Warningf("Attempting to rejoin the cluster: %v", m.ID())
		l.reg.rejoin <- struct{}{}
	}
}
