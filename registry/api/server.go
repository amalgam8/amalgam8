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

package api

import (
	"errors"
	"net"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/amalgam8/registry/api/middleware"
	"github.com/amalgam8/amalgam8/registry/api/protocol/amalgam8"
	"github.com/amalgam8/amalgam8/registry/api/protocol/eureka"
	"github.com/amalgam8/amalgam8/registry/api/uptime"
	"github.com/amalgam8/amalgam8/registry/store"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

const (
	module = "API"
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
}

// NewServer creates a new server based on the provided configuration options.
// Returns a valid Server interface on success or an error on failure
func NewServer(conf *Config) (Server, error) {
	if conf == nil {
		return nil, errors.New("null Service Discovery configuration provided")
	}

	s := &server{
		config: &*conf,
		logger: logging.GetLogger(module),
	}

	if s.config.HTTPAddressSpec == "" {
		s.config.HTTPAddressSpec = ":8080"
	}

	if s.config.CatalogMap == nil {
		s.config.CatalogMap = store.New(nil)
	}

	s.logger.Infof("Creating Service Discovery REST API on %s", s.config.HTTPAddressSpec)

	return s, nil
}

func (s *server) Start() error {
	handler, err := s.setup()
	if err != nil {
		return err
	}
	return s.serve(handler)
}

func (s *server) Stop() {
	s.logger.Info("Stopping rest server")

	if err := s.listener.Close(); err != nil {
		s.logger.WithFields(log.Fields{
			"error": err,
		}).Warn("Failed to close listener")
	}
}

func (s *server) setup() (http.Handler, error) {
	restAPI := rest.NewApi()

	restAPI.Use(
		&rest.RecoverMiddleware{},
		&middleware.AccessLog{},
		middleware.NewTrace())

	// Add the extension middlewares here
	for _, mw := range s.config.Middlewares {
		if mw != nil {
			restAPI.Use(mw)
		}
	}

	restAPI.Use(
		&middleware.MetricsMiddleware{},
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&middleware.GzipMiddleware{},
		&rest.ContentTypeCheckerMiddleware{})

	log.SetOutput(s.logger.Logger.Out)

	secureMw := middleware.NewRequireHTTPS(middleware.CheckRequest{
		IsSecure: middleware.IsUsingSecureConnection,
		Disabled: !s.config.RequireHTTPS,
	})

	var routes []*rest.Route
	routes = append(routes, uptime.RouteHandlers()...)

	amalgam8Routes := amalgam8.New(s.config.CatalogMap)
	eurekaRoutes := eureka.New(s.config.CatalogMap)
	authMw := &middleware.AuthMiddleware{TokenRouteParam: eureka.RouteParamToken, Authenticator: s.config.Authenticator}

	routes = append(routes, amalgam8Routes.RouteHandlers(secureMw, authMw)...)
	routes = append(routes, eurekaRoutes.RouteHandlers(secureMw, authMw)...)
	router, err := rest.MakeRouter(routes...)

	if err != nil {
		return nil, err
	}

	restAPI.SetApp(router)
	return restAPI.MakeHandler(), nil
}

func (s *server) serve(h http.Handler) error {
	s.logger.Infof("Starting Service Discovery REST API on %s", s.config.HTTPAddressSpec)

	listener, err := net.Listen("tcp", s.config.HTTPAddressSpec)
	if err != nil {
		s.logger.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to start the server")
		return err
	}

	s.listener = listener
	if err := http.Serve(listener, h); err != nil {
		s.logger.WithFields(log.Fields{
			"error": err,
		}).Warn("Server has aborted")
		return err
	}

	return nil
}
