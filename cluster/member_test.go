package cluster

import (
	"net"
	"time"
)

// Predefined test-time member constants

var (
	member1    = &member{MemberIP: net.ParseIP("192.168.1.201"), MemberPort: 6100, Timestamp: time.Now()}
	member2    = &member{MemberIP: net.ParseIP("192.168.1.202"), MemberPort: 6100, Timestamp: time.Now()}
	member3    = &member{MemberIP: net.ParseIP("192.168.1.203"), MemberPort: 6100, Timestamp: time.Now()}
	allMembers = []*member{member1, member2, member3}
)
