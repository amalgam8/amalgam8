package network

import (
	"net"
	"time"
)

// Network-related utility functions

const (
	waitInterval = time.Duration(100) * time.Millisecond
	waitTimeout  = time.Duration(2) * time.Minute
)

// GetPrivateIP returns a private IP address, or panics if no IP is available.
func GetPrivateIP() net.IP {
	addr := getPrivateIPIfAvailable()
	if addr.IsUnspecified() {
		panic("No private IP address is available")
	}
	return addr
}

// WaitForPrivateNetwork blocks until a private IP address is available, or a timeout is reached.
// Returns 'true' if a private IP is available before timeout is reached, and 'false' otherwise.
func WaitForPrivateNetwork() bool {
	deadline := time.Now().Add(waitTimeout)
	for {
		addr := getPrivateIPIfAvailable()
		if !addr.IsUnspecified() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(waitInterval)
	}
}

// Returns a private IP address, or unspecified IP (0.0.0.0) if no IP is available
func getPrivateIPIfAvailable() net.IP {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			continue
		}
		if !ip.IsLoopback() {
			return ip
		}
	}
	return net.IPv4zero
}
