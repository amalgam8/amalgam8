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

package discovery

import (
	"errors"
	"net"
	"net/http"

	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
	"github.com/ant0ine/go-json-rest/rest"
)

const (
	module = "SIDECARREGISTER"
)

// Server defines an interface for controlling server lifecycle
type Server interface {
	Start() error
	Stop()
}

type server struct {
	config   *Config
	listener net.Listener
	logger   *log.Entry
	mutex    sync.Mutex
}

// NewDiscoveryServer creates a new server based on the provided configuration options.
// Returns a valid Server interface on success or an error on failure
func NewDiscoveryServer(conf *Config) (Server, error) {
	if conf == nil {
		return nil, errors.New("null Discovery configuration provided")
	}

	s := &server{
		config: conf,
		logger: logging.GetLogger(module),
	}

	s.logger.Infof("Creating Discovery REST API on %s", s.config.HTTPAddressSpec)

	return s, nil
}

func (s *server) Start() error {
	handler, err := s.setup()
	if err != nil {
		return err
	}

	go func() {
		err = s.serve(handler)
		if err != nil {
			s.logger.WithFields(log.Fields{
				"error": err,
			}).Warn("Failed to start the server")
		}
	}()

	return nil
}

func (s *server) Stop() {
	s.logger.Info("Stopping rest server")
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.WithFields(log.Fields{
				"error": err,
			}).Warn("Failed to close listener")
		}
	}
}

func (s *server) setup() (http.Handler, error) {
	restAPI := rest.NewApi()

	restAPI.Use(
		&rest.RecoverMiddleware{},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.ContentTypeCheckerMiddleware{},
	)

	discoveryAPI := NewDiscovery(s.config.Discovery, s.config.Rules)

	router, err := rest.MakeRouter(discoveryAPI.Routes()...)
	if err != nil {
		return nil, err
	}

	restAPI.SetApp(router)
	return restAPI.MakeHandler(), nil
}

func (s *server) serve(h http.Handler) error {
	s.logger.Infof("Starting service discovery REST API on %s", s.config.HTTPAddressSpec)

	listener, err := net.Listen("tcp", s.config.HTTPAddressSpec)
	if err != nil {
		s.logger.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to start the server")
		return err
	}

	s.mutex.Lock()
	s.listener = listener
	s.mutex.Unlock()
	if err := http.Serve(listener, h); err != nil {
		s.logger.WithFields(log.Fields{
			"error": err,
		}).Warn("Server has aborted")
		return err
	}

	return nil
}
