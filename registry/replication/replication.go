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
	"time"

	"github.com/amalgam8/amalgam8/pkg/auth"
)

// Replication - interface for replication between registry cluster peers
type Replication interface {
	// Returns a replicator for a specific namespace responsible handling
	// the broadcast and send operations of outgoing events
	GetReplicator(namespace auth.Namespace) (Replicator, error)
	// Returns a channel on which the replicated incoming events from remote peers are received
	Notification() <-chan *InMessage
	// Starts a synchronization procedure with the cluster, returns a channel on which incoming sync events
	// are received and need to be stored locally
	Sync(waitTime time.Duration) <-chan *InMessage
	// Use this method to listen for synchronization request of remote peers, receive a channel on which
	// to send the outgoing sync events
	SyncRequest() <-chan chan []byte
	// Stops the replication and free resources
	Stop()
}
