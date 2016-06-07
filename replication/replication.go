package replication

import (
	"time"

	"github.com/amalgam8/registry/auth"
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
