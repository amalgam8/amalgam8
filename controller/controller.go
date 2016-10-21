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

package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/urfave/cli"

	"github.com/amalgam8/amalgam8/controller/api"
	"github.com/amalgam8/amalgam8/controller/config"
	"github.com/amalgam8/amalgam8/controller/metrics"
	"github.com/amalgam8/amalgam8/controller/middleware"
	"github.com/amalgam8/amalgam8/controller/rules"
	"github.com/amalgam8/amalgam8/controller/util/i18n"
	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/pkg/version"
)

// Main is the entrypoint for the controller when running as an executable
func Main() {
	app := cli.NewApp()

	app.Name = "controller"
	app.Usage = "Amalgam8 Controller"
	app.Version = version.Build.Version
	app.Flags = config.Flags
	app.Action = func(context *cli.Context) error {
		return Run(config.New(context))
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.WithError(err).Error("Command setup failed")
	}
}

// Run the controller with the given configuration
func Run(conf *config.Config) error {
	var err error

	logrus.ErrorKey = "error"
	logrus.SetLevel(conf.LogLevel)

	i18n.LoadLocales("./locales")

	setupHandler := middleware.NewSetupHandler()

	var validationErr error
	if validationErr = conf.Validate(); validationErr != nil {
		logrus.WithError(validationErr).Error("Validation of config failed")
		setupHandler.SetError(validationErr)
	}

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%v", conf.APIPort), setupHandler); err != nil {
			logrus.WithError(err).Error("Server init failed")
		}
	}()

	if validationErr != nil {
		setupHandler.SetError(validationErr)
		return validationErr
	}

	reporter := metrics.NewReporter()

	healthAPI := api.NewHealth(reporter)

	validator, err := rules.NewValidator()
	if err != nil {
		logrus.WithError(err).Error("Validator creation failed")
		setupHandler.SetError(err)
		return err
	}

	var ruleManager rules.Manager
	if conf.Database.Type == "redis" {
		ruleManager = rules.NewRedisManager(
			conf.Database.Host,
			conf.Database.Password,
			validator,
		)
	} else {
		ruleManager = rules.NewMemoryManager(validator)
	}
	rulesAPI := api.NewRule(ruleManager, reporter)

	a := rest.NewApi()
	a.Use(
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.RecoverMiddleware{
			EnableResponseStackTrace: false,
		},
		&rest.ContentTypeCheckerMiddleware{},
		&middleware.RequestIDMiddleware{},
		&middleware.LoggingMiddleware{},
		middleware.NewRequireHTTPS(middleware.CheckRequest{
			IsSecure: middleware.IsUsingSecureConnection,
			Disabled: !conf.RequireHTTPS,
		}),
	)

	authenticator, err := setupAuthenticator(conf)
	if err != nil {
		setupHandler.SetError(err)
		return err
	}

	authMw := &middleware.AuthMiddleware{Authenticator: authenticator}

	routes := rulesAPI.Routes(authMw)
	routes = append(routes, healthAPI.Routes()...)
	router, err := rest.MakeRouter(
		routes...,
	)
	if err != nil {
		setupHandler.SetError(err)
		return err
	}
	a.SetApp(router)

	setupHandler.SetHandler(a.MakeHandler())

	// Server is already started
	logrus.WithFields(logrus.Fields{
		"port": conf.APIPort,
	}).Info("Server started")

	select {}
}

func setupAuthenticator(conf *config.Config) (authenticator auth.Authenticator, err error) {
	if len(conf.AuthModes) > 0 {
		auths := make([]auth.Authenticator, len(conf.AuthModes))
		for i, mode := range conf.AuthModes {
			switch mode {
			case "trusted":
				auths[i] = auth.NewTrustedAuthenticator()
			case "jwt":
				jwtAuth, err := auth.NewJWTAuthenticator([]byte(conf.JWTSecret))
				if err != nil {
					return authenticator, fmt.Errorf("Failed to create the authentication module: %s", err)
				}
				auths[i] = jwtAuth
			default:
				return authenticator, fmt.Errorf("Failed to create the authentication module: unrecognized authentication mode '%s'", err)
			}
		}
		authenticator, err = auth.NewChainAuthenticator(auths)
		if err != nil {
			return authenticator, err
		}
	} else {
		authenticator = auth.DefaultAuthenticator()
	}

	return authenticator, nil
}
