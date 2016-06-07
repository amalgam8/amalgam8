package replication

import (
	"fmt"

	"github.com/amalgam8/registry/auth"
	"github.com/amalgam8/registry/cluster"
)

type outMessage struct {
	memberID  cluster.MemberID // Zero value - indicates a broadcast
	Namespace auth.Namespace
	Data      []byte
}

func (msg *outMessage) String() string {
	return fmt.Sprintf("member: %s, namespace: %s, data: %s", msg.memberID, msg.Namespace, msg.Data)
}

// InMessage - incoming replication message which is forwarded to the user
type InMessage struct {
	MemberID  cluster.MemberID
	Namespace auth.Namespace
	Data      []byte
}

func (msg *InMessage) String() string {
	return fmt.Sprintf("member: %s, namespace: %s, data: %s", msg.MemberID, msg.Namespace, msg.Data)
}
