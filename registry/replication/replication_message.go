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
	"fmt"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/cluster"
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
