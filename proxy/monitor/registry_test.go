package monitor

import (
	"testing"
	"time"

	"github.com/amalgam8/registry/client"
)

func TestCatalogComparison(t *testing.T) {
	r := registry{}

	cases := []struct {
		A, B  []client.ServiceInstance
		Equal bool
	}{
		{
			A:     []client.ServiceInstance{},
			B:     []client.ServiceInstance{},
			Equal: true,
		},
		{ // TTL and heartbeat should be ignored when comparing
			A: []client.ServiceInstance{
				{
					LastHeartbeat: time.Unix(0, 0),
					TTL:           1,
				},
			},
			B: []client.ServiceInstance{
				{
					LastHeartbeat: time.Unix(1, 0),
					TTL:           2,
				},
			},
			Equal: true,
		},
		{
			A: []client.ServiceInstance{
				{
					ServiceName: "ServiceA",
				},
			},
			B:     []client.ServiceInstance{},
			Equal: false,
		},
		{
			A: []client.ServiceInstance{
				{
					ServiceName: "ServiceA",
				},
			},
			B: []client.ServiceInstance{
				{
					ServiceName: "ServiceB",
				},
			},
			Equal: false,
		},
	}
	for _, c := range cases {
		actual := r.catalogsEqual(c.A, c.B)
		if actual != c.Equal {
			t.Errorf("catalogsEqual(%v, %v): expected %v, got %v", c.A, c.B, c.Equal, actual)
		}
	}
}
