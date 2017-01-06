package util

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/amalgam8/amalgam8/pkg/api"
)

// SplitHostPort splits an endpoint into its constituent IP and port.
func SplitHostPort(endpoint api.ServiceEndpoint) (net.IP, uint16, error) {
	switch endpoint.Type {
	case "tcp", "udp":
		return splitHostPortTCPUDP(endpoint.Value)
	case "http", "https":
		return splitHostPortHTTP(endpoint.Value)
	default:
		return nil, 0, fmt.Errorf("unsupported endpoint type: %s", endpoint.Type)
	}
}

func splitHostPortTCPUDP(value string) (net.IP, uint16, error) {
	// Assume value is "host:port"
	host, port, err := net.SplitHostPort(value)

	// Assume value is "host" (no port)
	if err != nil {
		host = value
		port = "0"
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, 0, fmt.Errorf("could not parse '%s' as ip:port", value)
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, 0, err
	}

	return ip, uint16(portNum), nil
}

func splitHostPortHTTP(value string) (net.IP, uint16, error) {
	isHTTP := strings.HasPrefix(value, "http://")
	isHTTPS := strings.HasPrefix(value, "https://")
	if !isHTTPS && !isHTTP {
		value = "http://" + value
		isHTTP = true
	}

	parsedURL, err := url.Parse(value)
	if err != nil {
		return nil, 0, err
	}

	ip, port, err := splitHostPortTCPUDP(parsedURL.Host)
	if err != nil {
		return nil, 0, err
	}

	// Use default port, if not specified
	if port == 0 {
		if isHTTP {
			port = 80
		} else if isHTTPS {
			port = 443
		}
	}

	return ip, port, nil
}
