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

package replication

import (
	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/cluster"
	"github.com/amalgam8/amalgam8/registry/utils/channels"
)

// Replicator - the per-namespace interface for sending messages
type Replicator interface {
	Broadcast(data []byte) error
	Send(memberID cluster.MemberID, data []byte) error
}

type replicator struct {
	auth.Namespace
	broadcast channels.ChannelTimeout
	repair    channels.ChannelTimeout
}

func newReplicator(namespace auth.Namespace, broadcast channels.ChannelTimeout, repair channels.ChannelTimeout) *replicator {
	return &replicator{
		Namespace: namespace,
		broadcast: broadcast,
		repair:    repair,
	}
}

func (r *replicator) Broadcast(d []byte) error {
	return r.broadcast.Send(&outMessage{Namespace: r.Namespace, Data: d}, repTimeout)
}

func (r *replicator) Send(memberID cluster.MemberID, d []byte) error {
	return r.repair.Send(&outMessage{memberID: memberID, Namespace: r.Namespace, Data: d}, repTimeout)
}
