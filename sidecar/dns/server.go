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

	"errors"
	"strconv"
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

	err = s.findMatchingServices(question, request, response, serviceInstances)
	return err

}

func (s *Server) retrieveServices(question dns.Question, request, response *dns.Msg) ([]*client.ServiceInstance, error) {
	var ServiceInstances []*client.ServiceInstance
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
		response.SetRcode(request, dns.RcodeBadName)
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
			ServiceInstances, err = s.retrieveServicesFromSRVQuery(serviceName, tagOrProtocol)
		} else {
			serviceName := fullDomainRequestArray[numberOfLabels-3]
			tagsOrProtocol := fullDomainRequestArray[:numberOfLabels-3]
			ServiceInstances, err = s.retrieveServicesFromRegularQuery(serviceName, tagsOrProtocol)

		}

	} else if fullDomainRequestArray[numberOfLabels-2] == "instance" && (question.Qtype == dns.TypeA ||
		question.Qtype == dns.TypeAAAA) && numberOfLabels == 3 {

		instanceID := fullDomainRequestArray[0]
		ServiceInstances, err = s.retrieveServicesFromInstanceQuery(instanceID)

	}

	/*
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
	*/

	return ServiceInstances, err
}

func (s *Server) retrieveServicesFromSRVQuery(serviceName string, tagOrProtocol string) ([]*client.ServiceInstance, error) {
	if tagOrProtocol == "tcp" || tagOrProtocol == "udp" || tagOrProtocol == "http" || tagOrProtocol == "https" {
		ServiceInstances, err := s.config.DiscoveryClient.ListServiceInstances(serviceName)
		if err != nil {
			return ServiceInstances, err
		}
		k := 0
		for i, serviceInstance := range ServiceInstances {
			if serviceInstance.Endpoint.Type == tagOrProtocol {
				ServiceInstances[k] = ServiceInstances[i]
				k++
			}
		}
		ServiceInstances = ServiceInstances[:k]
		return ServiceInstances, nil
	}
	filters := client.InstanceFilter{ServiceName: serviceName, Tags: []string{tagOrProtocol}}
	ServiceInstances, err := s.config.DiscoveryClient.ListInstances(filters)
	return ServiceInstances, err

}

func (s *Server) retrieveServicesFromRegularQuery(serviceName string, tagsOrProtocol []string) ([]*client.ServiceInstance, error) {
	numberOfProtocols := 0
	var protocol string
	for i, tag := range tagsOrProtocol {
		if tag == "tcp" || tag == "udp" || tag == "http" || tag == "https" {
			protocol = tag
			numberOfProtocols++
			tagsOrProtocol = append(tagsOrProtocol[:i], tagsOrProtocol[i+1:]...)

		}
	}
	if numberOfProtocols > 1 {
		return nil, errors.New("Error while parsing request - more then one protocol found. " +
			"request needs to contain at most one protocol type")
	}
	filters := client.InstanceFilter{ServiceName: serviceName, Tags: tagsOrProtocol}
	ServiceInstances, err := s.config.DiscoveryClient.ListInstances(filters)
	if err != nil {
		return ServiceInstances, err
	}
	if numberOfProtocols == 1 {
		k := 0
		for i, serviceInstance := range ServiceInstances {
			if serviceInstance.Endpoint.Type == protocol {
				ServiceInstances[k] = ServiceInstances[i]
				k++

			}
		}
		ServiceInstances = ServiceInstances[:k]
		return ServiceInstances, nil
	}
	return ServiceInstances, nil

}

func (s *Server) retrieveServicesFromInstanceQuery(instanceID string) ([]*client.ServiceInstance, error) {
	ServiceInstances, err := s.config.DiscoveryClient.ListInstances(client.InstanceFilter{})
	if err != nil {
		return ServiceInstances, err
	}
	for _, serviceInstance := range ServiceInstances {
		if serviceInstance.ID == instanceID {
			return []*client.ServiceInstance{serviceInstance}, nil
		}
	}
	return nil, fmt.Errorf("Error : didn't find a service with the id given %s", instanceID)
}

func (s *Server) findMatchingServices(question dns.Question, request, response *dns.Msg,
	serviceInstances []*client.ServiceInstance) error {
	numOfMatchingRecords := 0
	for _, serviceInstance := range serviceInstances {
		endPointType := serviceInstance.Endpoint.Type
		var ip net.IP
		var err error
		var port string

		switch endPointType {
		case "tcp", "udp":
			ip,port,  err = validateEndPointTypeTCPAndUDP(serviceInstance.Endpoint.Value)

		case "http", "https":
			ip,port,  err = validateEndPointTypeHTTP(serviceInstance.Endpoint.Value)

		default:
			continue
		}
		if err != nil {
			continue
		}
		numOfMatchingRecords++
		if question.Qtype == dns.TypeSRV {
			fullDomainRequestArray := dns.SplitDomainName(question.Name)
			domainName := fullDomainRequestArray[len(fullDomainRequestArray)-1]
			instanceID := serviceInstance.ID
			targetName := instanceID + ".instance." + domainName + "."
			portNumber,_ := strconv.Atoi(port)
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

func validateEndPointTypeTCPAndUDP(value string) (net.IP,string, error) {
	ip, port, err := net.SplitHostPort(value)
	if err != nil {
		return nil,port, err
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil,port, fmt.Errorf("tcp/udp ip value %s is not a valid ip format", ip)
	}
	return parsedIP,port, nil
}

func validateEndPointTypeHTTP(value string) (net.IP,string, error) {
	startsWithHTTP := strings.HasPrefix(value, "http://")
	startsWithHTTPS := strings.HasPrefix(value, "https://")
	if !startsWithHTTPS && !startsWithHTTP {
		value += "http://"
	}
	var port string
	parsedURL, err := url.Parse(value)
	if err != nil {
		return nil,port, err
	}
	host := parsedURL.Host
	ip, port , err := net.SplitHostPort(host)
	if err != nil {
		return nil,port, err
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil,port, fmt.Errorf("url ip value %s is not a valid ip format", ip)
	}
	return parsedIP,port, nil

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
