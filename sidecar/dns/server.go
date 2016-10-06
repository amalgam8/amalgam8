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

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/client"
	"github.com/miekg/dns"
	"strings"
)

type Server struct {
	config    Config
	dnsServer *dns.Server
}

type Config struct {
	DiscoveryClient client.Discovery
	Port            uint16
	Domain          string
}

/***********************************************************/
func CreateNewClient() (client.Client, error) {
	conf := client.Config{URL: "http://172.17.0.02:8080"}
	return client.New(conf)
}

func CreateDNSServer() (*Server, error){
	myclient, _ := CreateNewClient()
	dnsConfig := Config{
		DiscoveryClient: myclient ,
		Port:            8053,
		Domain:          "amalgam8",
	}
	return NewServer(dnsConfig)

}
/***********************************************************/
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

func (s *Server) ListenAndServe() error {
	logrus.Info("Starting DNS server")
	err := s.dnsServer.ListenAndServe()

	if err != nil {
		logrus.WithError(err).Errorf("Error starting DNS server")
	}

	return nil
}

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
	err, ServiceInstances := s.retrieveServices(question, request, response)

	if err != nil {
		return err
	}
	numOfMatchingRecords := 0
	for _, serviceInstance := range ServiceInstances {
		endPointType := serviceInstance.Endpoint.Type

		if endPointType == "tcp" {
			numOfMatchingRecords++
			record := &dns.A{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				// TODO: what to do with the port.
				A: net.ParseIP(strings.Split(serviceInstance.Endpoint.Value, ":")[0]),
			}
			response.Answer = append(response.Answer, record)
		}
	}
	if numOfMatchingRecords == 0 {
		//Non-Existent Domain
		response.SetRcode(request, dns.RcodeNameError )
		return fmt.Errorf("Non-Existent Domain	 %s", question.Name)

	}
	response.SetRcode(request, dns.RcodeSuccess)
	return nil


}


func (s* Server)retrieveServices(question dns.Question, request, response *dns.Msg) (error, []*client.ServiceInstance) {
	// parse query :
	// Query format:
	// [tag]*.<service>.<domain>.
	numberOfLabels, isValidDomain := dns.IsDomainName(question.Name)
	if isValidDomain == false {
		response.SetRcode(request, dns.RcodeBadName)
		return fmt.Errorf("Invalid Domain name %s", question.Name) , nil
	}

	fullDomainRequestArray := dns.SplitDomainName(question.Name)
	if numberOfLabels == 1 {
		response.SetRcode(request, dns.RcodeNameError )
		return fmt.Errorf("service name wasn't included in domain %s", question.Name), nil

	}
	serviceName := fullDomainRequestArray[numberOfLabels - 2]
	tags := fullDomainRequestArray[:len(fullDomainRequestArray) -2]
	var ServiceInstances []*client.ServiceInstance
	var err error = nil
	if len(tags) == 0 {
		ServiceInstances, err = s.config.DiscoveryClient.ListServiceInstances(serviceName)
	} else {
		filters :=client.InstanceFilter{ServiceName:serviceName,Tags:tags}
		ServiceInstances, err = s.config.DiscoveryClient.ListInstances(filters)
	}
	if err != nil {
		// TODO: what Error should we return ?
		response.SetRcode(request, dns.RcodeServerFailure )
		return fmt.Errorf("Error while reading from registry: %s",err.Error() ) , nil
	}
	return nil, ServiceInstances
}

func validate(config *Config) error {
	// TODO: Validate port

	if config.DiscoveryClient == nil {
		return fmt.Errorf("Discovery client is nil")
	}

	config.Domain = dns.Fqdn(config.Domain)

	return nil
}
