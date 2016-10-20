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
	case dns.TypeANY:
	default:
		response.SetRcode(request, dns.RcodeServerFailure)
		return fmt.Errorf("unsupported DNS question type: %v", dns.Type(question.Qtype).String())
	}

	serviceInstances, err := s.retrieveServices(question, request, response)

	if err != nil {
		return err
	}

	err = s.findMatchingServices(question, request, response, serviceInstances)
	return err

}

func (s *Server) retrieveServices(question dns.Question, request, response *dns.Msg) ([]*client.ServiceInstance, error) {
	// parse query :
	// Query format:
	// [tag]*.<service>.<domain>.
	numberOfLabels, isValidDomain := dns.IsDomainName(question.Name)
	if !isValidDomain {
		response.SetRcode(request, dns.RcodeBadName)
		return nil, fmt.Errorf("Invalid Domain name %s", question.Name)
	}

	fullDomainRequestArray := dns.SplitDomainName(question.Name)
	if len(fullDomainRequestArray) == 1 {
		response.SetRcode(request, dns.RcodeNameError)
		return nil, fmt.Errorf("service name wasn't included in domain %s", question.Name)

	}
	serviceName := fullDomainRequestArray[numberOfLabels-2]
	tags := fullDomainRequestArray[:len(fullDomainRequestArray)-2]
	var ServiceInstances []*client.ServiceInstance
	var err error
	if len(tags) == 0 {
		ServiceInstances, err = s.config.DiscoveryClient.ListServiceInstances(serviceName)
	} else {
		filters := client.InstanceFilter{ServiceName: serviceName, Tags: tags}
		ServiceInstances, err = s.config.DiscoveryClient.ListInstances(filters)
	}
	if err != nil {
		// TODO: what Error should we return ?
		response.SetRcode(request, dns.RcodeServerFailure)
		return nil, fmt.Errorf("Error while reading from registry: %s", err.Error())
	}
	return ServiceInstances, nil
}

func (s *Server) findMatchingServices(question dns.Question, request, response *dns.Msg,
	serviceInstances []*client.ServiceInstance) error {
	numOfMatchingRecords := 0
	for _, serviceInstance := range serviceInstances {
		endPointType := serviceInstance.Endpoint.Type
		var ip net.IP
		var err error

		switch endPointType {
		case "tcp", "udp":
			ip, err = validateEndPointTypeTCPAndUDP(serviceInstance.Endpoint.Value)

		case "http", "https":
			ip, err = validateEndPointTypeHTTP(serviceInstance.Endpoint.Value)

		default:
			continue
		}
		if err != nil {
			continue
		}
		numOfMatchingRecords++
		if ip.To4() != nil {
			record := createARecord(question.Name, ip)
			response.Answer = append(response.Answer, record)
		} else if ip.To16() != nil {
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

func validateEndPointTypeTCPAndUDP(value string) (net.IP, error) {
	ip, _, err := net.SplitHostPort(value)
	if err != nil {
		return nil, err
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("tcp/udp ip value %s is not a valid ip format", ip)
	}
	return parsedIP, nil
}

func validateEndPointTypeHTTP(value string) (net.IP, error) {
	startsWithHttp := strings.HasPrefix(value, "http://")
	startsWithHttps := strings.HasPrefix(value, "https://")
	if !startsWithHttps && !startsWithHttp {
		value += "http://"
	}
	parsedURL, err := url.Parse(value)
	if err != nil {
		return nil, err
	}
	host := parsedURL.Host
	ip, _, err := net.SplitHostPort(host)
	if err != nil {
		return nil, err
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("url ip value %s is not a valid ip format", ip)
	}
	return parsedIP, nil

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
