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

package dns

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"math/rand"

	"sort"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/client"
	"github.com/miekg/dns"
)

// Server represent a DNS server. has config field for port,domain,and client discovery, and the DNS server itself
type Server struct {
	dnsServer       *dns.Server
	discoveryClient client.Discovery

	domain       string
	domainLabels int
}

// Config represents the DNS server configurations.
type Config struct {
	DiscoveryClient client.Discovery
	Port            uint16
	Domain          string
}

// NewServer creates a new instance of a DNS server with the given configurations
func NewServer(config Config) (*Server, error) {
	err := validate(&config)
	if err != nil {
		return nil, err
	}
	s := &Server{
		discoveryClient: config.DiscoveryClient,
		domain:          config.Domain,
		domainLabels:    len(dns.Split(config.Domain)),
	}

	// Setup DNS muxing
	mux := dns.NewServeMux()
	mux.HandleFunc(config.Domain, s.handleRequest)

	// Setup a DNS server
	s.dnsServer = &dns.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Net:     "udp",
		Handler: mux,
	}

	return s, nil
}

// ListenAndServe starts the DNS server
func (s *Server) ListenAndServe() error {
	logrus.Info("Starting DNS server")
	err := s.dnsServer.ListenAndServe()

	if err != nil {
		logrus.WithError(err).Errorf("Error starting DNS server")
	}

	return nil
}

// Shutdown stops the DNS server
func (s *Server) Shutdown() error {
	logrus.Info("Shutting down DNS server")
	err := s.dnsServer.Shutdown()

	if err != nil {
		logrus.WithError(err).Errorf("Error shutting down DNS server")
	} else {
		logrus.Info("DNS server has shutdown")
	}

	return err
}

func (s *Server) handleRequest(w dns.ResponseWriter, request *dns.Msg) {
	response := new(dns.Msg)
	response.SetReply(request)
	response.Extra = request.Extra
	response.Authoritative = true
	response.RecursionAvailable = false

	for i, question := range request.Question {
		err := s.handleQuestion(question, request, response)
		if err != nil {
			logrus.WithError(err).Errorf("Error handling DNS question %d: %s", i, question.String())
			break
		}
	}
	err := w.WriteMsg(response)
	if err != nil {
		logrus.WithError(err).Errorf("Error writing DNS response")
	}
}

func (s *Server) handleQuestion(question dns.Question, request, response *dns.Msg) error {
	switch question.Qclass {
	case dns.ClassINET:
	default:
		response.SetRcode(request, dns.RcodeServerFailure)
		return fmt.Errorf("unsupported DNS question class: %v", dns.Class(question.Qclass).String())
	}

	switch question.Qtype {
	case dns.TypeA:
	case dns.TypeAAAA:
	case dns.TypeSRV:
	default:
		response.SetRcode(request, dns.RcodeServerFailure)
		return fmt.Errorf("unsupported DNS question type: %v", dns.Type(question.Qtype).String())
	}

	instances, err := s.retrieveInstances(question, request, response)
	if err != nil {
		return err
	}
	return s.createRecords(question, request, response, instances)
}

func (s *Server) retrieveInstances(question dns.Question, request, response *dns.Msg) ([]*client.ServiceInstance, error) {
	// Validate the domain name in the question.
	_, isValidDomain := dns.IsDomainName(question.Name)
	if !isValidDomain {
		response.SetRcode(request, dns.RcodeNameError)
		return nil, fmt.Errorf("invalid domain name")
	}
	labels := dns.SplitDomainName(question.Name)

	// Query format can be either of the following:
	// 1. [tag|protocol|instanceID]*.<service>.<domain> (A/AAAA query)
	// 2. _<service>._<tag|protocol|instanceID>.domain> (SRV query per RFC 2782)

	if len(labels) < 1+s.domainLabels {
		response.SetRcode(request, dns.RcodeNameError)
		return nil, fmt.Errorf("no service specified")
	}

	// Extract service name and filtering labels (tags / protocol / instance ID)
	var service string
	var filters []string
	switch question.Qtype {
	case dns.TypeA, dns.TypeAAAA:
		servicePos := len(labels) - s.domainLabels - 1
		service = labels[servicePos]
		filters = labels[0:servicePos]
	case dns.TypeSRV:
		// Make sure the query syntax complies to RFC 2782
		if len(labels) != 2+s.domainLabels ||
			!strings.HasPrefix(labels[0], "_") ||
			!strings.HasPrefix(labels[1], "_") {
			response.SetRcode(request, dns.RcodeNameError)
			return nil, fmt.Errorf("invalid SRV query syntax")
		}
		service = strings.TrimPrefix(labels[0], "_")
		filters = []string{strings.TrimPrefix(labels[1], "_")}
	}

	// Dispatch query to registry
	instances, err := s.discoveryClient.ListServiceInstances(service)
	if err != nil {
		response.SetRcode(request, dns.RcodeServerFailure)
		return nil, err
	}

	filteredInstances, err := s.filterInstances(instances, filters)
	if err != nil {
		response.SetRcode(request, dns.RcodeNameError)
	}
	return filteredInstances, err
}

func (s *Server) filterInstances(instances []*client.ServiceInstance, filters []string) ([]*client.ServiceInstance, error) {
	// If no filters are specified, all instances match vacuously
	if len(filters) == 0 {
		return instances, nil
	}

	// If only a single filter is specified, first attempt to match it as an instance ID
	if len(filters) == 1 {
		id := filters[0]
		for _, instance := range instances {
			if instance.ID == id {
				return []*client.ServiceInstance{instance}, nil
			}
		}
	}

	// Split tags and protocol filters
	protocol := ""
	tags := make([]string, 0, len(filters))
	for _, filter := range filters {
		switch filter {
		case "tcp", "udp", "http", "https":
			if protocol != "" {
				return nil, fmt.Errorf("invalid DNS query: more than one protocol specified")
			}
			protocol = filter
		default:
			tags = append(tags, filter)
		}
	}

	// Sort the tags
	sort.Strings(tags)

	// Apply filters
	// Note: filtering is done in-place, without allocating another array
	k := 0
	for _, instance := range instances {

		// Apply protocol filter
		if protocol != "" && instance.Endpoint.Type != protocol {
			continue
		}

		// Apply tags filter
		// Note: We presort the instance tags so that we can filter with a single pass
		sort.Strings(instance.Tags)
		j := 0
		match := true
		for _, tag := range tags {
			found := false
			for i := j; i < len(instance.Tags); i++ {
				if tag == instance.Tags[i] {
					found = true
					j = i + 1
					break
				}
			}
			if !found {
				match = false
				break
			}
		}
		if !match {
			continue
		}

		instances[k] = instance
		k++
	}

	return instances[0:k], nil
}

func (s *Server) createRecords(question dns.Question, request, response *dns.Msg, instances []*client.ServiceInstance) error {
	answer := make([]dns.RR, 0, 3)
	extra := make([]dns.RR, 0, 3)

	for _, instance := range instances {
		ip, port, err := splitHostPort(instance.Endpoint)
		if err != nil {
			logrus.WithError(err).Warnf("unable to resolve ip address for instance '%s' in DNS query '%s'",
				instance.ID, question.Name)
			continue
		}

		switch question.Qtype {
		case dns.TypeA:
			ipV4 := ip.To4()
			if ipV4 != nil {
				answer = append(answer, createARecord(question.Name, ipV4))
			}
		case dns.TypeAAAA:
			ipV4 := ip.To4()
			if ipV4 == nil {
				answer = append(answer, createARecord(question.Name, ip.To16()))
			}
		case dns.TypeSRV:
			target := fmt.Sprintf("%s.%s.%s", instance.ID, instance.ServiceName, s.domain)
			answer = append(answer, createSRVRecord(question.Name, port, target))

			ipV4 := ip.To4()
			if ipV4 != nil {
				extra = append(extra, createARecord(question.Name, ipV4))
			} else {
				extra = append(extra, createAAAARecord(question.Name, ip.To16()))
			}

		}
	}

	if len(answer) == 0 {
		response.SetRcode(request, dns.RcodeNameError)
		return nil
	}

	// Poor-man's load balancing: randomize returned records order
	shuffleRecords(answer)
	shuffleRecords(extra)

	response.Answer = append(response.Answer, answer...)
	response.Extra = append(response.Extra, extra...)
	response.SetRcode(request, dns.RcodeSuccess)
	return nil

}

func splitHostPort(endpoint client.ServiceEndpoint) (net.IP, uint16, error) {
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

func createARecord(name string, ip net.IP) *dns.A {
	record := &dns.A{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    0,
		},
		A: ip,
	}
	return record
}

func createAAAARecord(name string, ip net.IP) *dns.AAAA {
	record := &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    0,
		},
		AAAA: ip,
	}
	return record
}

func createSRVRecord(name string, port uint16, target string) *dns.SRV {
	record := &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET,
			Ttl:    0,
		},
		Port:     port,
		Priority: 0,
		Weight:   0,
		Target:   target,
	}
	return record
}

func validate(config *Config) error {
	if config.DiscoveryClient == nil {
		return fmt.Errorf("Discovery client is nil")
	}

	config.Domain = dns.Fqdn(config.Domain)

	return nil
}

func shuffleRecords(records []dns.RR) {
	for i := range records {
		j := rand.Intn(i + 1)
		records[i], records[j] = records[j], records[i]
	}
}
