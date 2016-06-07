package replication

import (
	"github.com/amalgam8/registry/cluster"
)

// Config - configuration structure of the replication module
type Config struct {
	Membership  cluster.Membership
	Registrator cluster.Registrator
}
