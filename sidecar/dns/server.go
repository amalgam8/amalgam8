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

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/client"
	"github.com/miekg/dns"
)

// Server represent a DNS server. has config field for port,domain,and client discovery, and the DNS server itself
type Server struct {
	config    Config
	dnsServer *dns.Server
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
		config: config,
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
			// TODO: what should the dns response return ?
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

	serviceInstances, err := s.retrieveServices(question, request, response)

	if err != nil {
		return err
	}
	err = s.createRecordsForInstances(question, request, response, serviceInstances)
	return err

}

func (s *Server) retrieveServices(question dns.Question, request, response *dns.Msg) ([]*client.ServiceInstance, error) {
	var serviceInstances []*client.ServiceInstance
	var err error
	// parse query :
	// Query format:
	// [tag or endpoint type]*.<service>.service.<domain>.
	// <instance_id>.instance.<domain>.
	// For SRV types we also support :
	// _<service>._<tag or endpoint type>.<domain>.

	/// IsDomainName checks if s is a valid domain name
	//  When false is returned the number of labels is not
	// defined.  Also note that this function is extremely liberal; almost any
	// string is a valid domain name as the DNS is 8 bit protocol. It checks if each
	// label fits in 63 characters, but there is no length check for the entire
	// string s. I.e.  a domain name longer than 255 characters is considered valid.
	numberOfLabels, isValidDomain := dns.IsDomainName(question.Name)
	if !isValidDomain {
		response.SetRcode(request, dns.RcodeFormatError)
		return nil, fmt.Errorf("Invalid Domain name %s", question.Name)
	}
	fullDomainRequestArray := dns.SplitDomainName(question.Name)
	if len(fullDomainRequestArray) == 1 || len(fullDomainRequestArray) == 2 {
		response.SetRcode(request, dns.RcodeNameError)
		return nil, fmt.Errorf("service name wasn't included in domain %s", question.Name)
	}
	if fullDomainRequestArray[numberOfLabels-2] == "service" {
		if question.Qtype == dns.TypeSRV && numberOfLabels == 4 &&
			strings.HasPrefix(fullDomainRequestArray[0], "_") &&
			strings.HasPrefix(fullDomainRequestArray[1], "_") {
			// SRV Query :
			tagOrProtocol := fullDomainRequestArray[1][1:]
			serviceName := fullDomainRequestArray[0][1:]
			serviceInstances, err = s.retrieveInstancesForServiceQuery(serviceName, request, response, tagOrProtocol)
		} else {
			serviceName := fullDomainRequestArray[numberOfLabels-3]
			tagsOrProtocol := fullDomainRequestArray[:numberOfLabels-3]
			serviceInstances, err = s.retrieveInstancesForServiceQuery(serviceName, request, response, tagsOrProtocol...)

		}

	} else if fullDomainRequestArray[numberOfLabels-2] == "instance" && (question.Qtype == dns.TypeA ||
		question.Qtype == dns.TypeAAAA) && numberOfLabels == 3 {

		instanceID := fullDomainRequestArray[0]
		serviceInstances, err = s.retrieveInstancesForInstanceQuery(instanceID, request, response)
	}
	return serviceInstances, err
}

func (s *Server) retrieveInstancesForServiceQuery(serviceName string, request, response *dns.Msg, tagOrProtocol ...string) ([]*client.ServiceInstance, error) {
	protocol := ""
	tags := make([]string, 0, len(tagOrProtocol))

	// Split tags and protocol filters
	for _, tag := range tagOrProtocol {
		switch tag {
		case "tcp", "udp", "http", "https":
			if protocol != "" {
				response.SetRcode(request, dns.RcodeFormatError)
				return nil, fmt.Errorf("invalid DNS query: more than one protocol specified")
			}
			protocol = tag
		default:
			tags = append(tags, tag)
		}
	}
	filters := client.InstanceFilter{ServiceName: serviceName, Tags: tags}

	// Dispatch query to registry
	serviceInstances, err := s.config.DiscoveryClient.ListInstances(filters)
	if err != nil {
		response.SetRcode(request, dns.RcodeServerFailure)
		return nil, err
	}

	// Apply protocol filter
	if protocol != "" {
		k := 0
		for _, serviceInstance := range serviceInstances {
			if serviceInstance.Endpoint.Type == protocol {
				serviceInstances[k] = serviceInstance
				k++
			}
		}
		serviceInstances = serviceInstances[:k]
	}

	return serviceInstances, nil
}

func (s *Server) retrieveInstancesForInstanceQuery(instanceID string, request, response *dns.Msg) ([]*client.ServiceInstance, error) {
	serviceInstances, err := s.config.DiscoveryClient.ListInstances(client.InstanceFilter{})
	if err != nil {
		response.SetRcode(request, dns.RcodeServerFailure)
		return serviceInstances, err
	}
	for _, serviceInstance := range serviceInstances {
		if serviceInstance.ID == instanceID {
			return []*client.ServiceInstance{serviceInstance}, nil
		}
	}
	response.SetRcode(request, dns.RcodeNameError)
	return nil, fmt.Errorf("Error : didn't find a service with the id given %s", instanceID)
}

func (s *Server) createRecordsForInstances(question dns.Question, request, response *dns.Msg,
	serviceInstances []*client.ServiceInstance) error {
	numOfMatchingRecords := 0
	for _, serviceInstance := range serviceInstances {
		endPointType := serviceInstance.Endpoint.Type
		var ip net.IP
		var err error
		var port string

		switch endPointType {
		case "tcp", "udp":
			ip, port, err = validateEndPointTypeTCPAndUDP(serviceInstance.Endpoint.Value)
		case "http", "https":
			ip, port, err = validateEndPointTypeHTTP(serviceInstance.Endpoint.Value)

		default:
			continue
		}
		if err != nil {
			continue
		}
		numOfMatchingRecords++
		if question.Qtype == dns.TypeSRV {

			domainName := s.config.Domain
			instanceID := serviceInstance.ID
			targetName := fmt.Sprintf("%s.instance.%s", instanceID, domainName)
			portNumber, _ := strconv.Atoi(port)
			recordSRV := &dns.SRV{Hdr: dns.RR_Header{
				Name:   question.Name,
				Rrtype: dns.TypeSRV,
				Class:  dns.ClassINET,
				Ttl:    0,
			},
				Port:     uint16(portNumber),
				Priority: 1,
				Target:   targetName,
				Weight:   1,
			}
			response.Answer = append(response.Answer, recordSRV)
			if ip.To4() != nil {
				recordA := createARecord(targetName, ip)
				response.Extra = append(response.Extra, recordA)
			} else if ip.To16() != nil {
				recordAAAA := createAAARecord(targetName, ip)
				response.Extra = append(response.Extra, recordAAAA)
			}

		} else if ip.To4() != nil && question.Qtype == dns.TypeA {
			record := createARecord(question.Name, ip)
			response.Answer = append(response.Answer, record)
		} else if ip.To16() != nil && question.Qtype == dns.TypeAAAA {
			record := createAAARecord(question.Name, ip)
			response.Answer = append(response.Answer, record)
		}
	}
	if numOfMatchingRecords == 0 {
		//Non-Existent Domain
		response.SetRcode(request, dns.RcodeNameError)
		return fmt.Errorf("Non-Existent Domain	 %s", question.Name)

	}
	response.SetRcode(request, dns.RcodeSuccess)
	return nil

}

func validateEndPointTypeTCPAndUDP(value string) (net.IP, string, error) {
	ip, port, err := net.SplitHostPort(value)
	var parsedIP net.IP
	if err != nil {
		parsedIP = net.ParseIP(value)
		if parsedIP == nil {
			return nil, port, err
		}
		port = "0"
	} else {
		parsedIP = net.ParseIP(ip)
		if parsedIP == nil {
			return nil, port, fmt.Errorf("tcp/udp ip value %s is not a valid ip format", ip)
		}
	}
	return parsedIP, port, nil
}

func validateEndPointTypeHTTP(value string) (net.IP, string, error) {
	startsWithHTTP := strings.HasPrefix(value, "http://")
	startsWithHTTPS := strings.HasPrefix(value, "https://")
	if !startsWithHTTPS && !startsWithHTTP {
		value = "http://" + value
	}
	var port string
	parsedURL, err := url.Parse(value)

	if err != nil {
		return nil, port, err
	}

	host := parsedURL.Host
	ip, port, err := validateEndPointTypeTCPAndUDP(host)
	if err != nil {
		return nil, port, err
	}
	if port == "0" && startsWithHTTPS {
		port = "430"
	} else if port == "0" {
		port = "80"
	}
	return ip, port, nil

}

func createARecord(questionName string, ip net.IP) *dns.A {
	record := &dns.A{
		Hdr: dns.RR_Header{
			Name:   questionName,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    0,
		},

		A: ip,
	}
	return record
}

func createAAARecord(questionName string, ip net.IP) *dns.AAAA {
	record := &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   questionName,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    0,
		},

		AAAA: ip,
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
