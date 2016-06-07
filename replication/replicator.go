package replication

import (
	"github.com/amalgam8/registry/auth"
	"github.com/amalgam8/registry/cluster"
	"github.com/amalgam8/registry/utils/channels"
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
