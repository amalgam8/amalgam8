// Package cluster defines and implements types related to service discovery clustering.
package cluster

import "github.com/amalgam8/registry/utils/health"

// Module name to be used in logging
const module = "CLUSTER"

// Cluster represents a collection of service discovery servers ("members").
type Cluster interface {
	Registrator(m Member) Registrator
	Membership() Membership
}

// New creates and initializes a new cluster with the given configuration.
// Nil argument results with default values for the configuration.
func New(conf *Config) (Cluster, error) {
	conf = defaultize(conf)

	b, err := newBackend(conf)
	if err != nil {
		return nil, err
	}

	m := newMembership(b, conf.TTL, conf.ScanInterval)
	c := &cluster{backend: b, membership: m, conf: conf}

	m.StartMonitoring()

	hc := newHealthChecker(m, conf.Size)
	health.Register(module, hc)

	return c, nil
}

// cluster is an implementation of the Cluster interface
type cluster struct {
	backend    backend
	membership *membership
	conf       *Config
}

func (c *cluster) Registrator(m Member) Registrator {
	return newRegistrator(c.backend, m, c.membership, c.conf.RenewInterval)
}

func (c *cluster) Membership() Membership {
	return c.membership
}
