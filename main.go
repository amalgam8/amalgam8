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

package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/api"
	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/config"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/manager"
	"github.com/amalgam8/controller/metrics"
	"github.com/amalgam8/controller/middleware"
	"github.com/amalgam8/controller/nginx"
	"github.com/amalgam8/controller/notification"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "controller"
	app.Usage = "Amalgam8 Controller"
	app.Version = "0.1"
	app.Flags = config.Flags
	app.Action = controllerCommand

	err := app.Run(os.Args)
	if err != nil {
		logrus.WithError(err).Error("Command setup failed")
	}
}

func controllerCommand(context *cli.Context) {
	conf := config.New(context)
	if err := controllerMain(*conf); err != nil {
		logrus.WithError(err).Error("Controller setup failed")
		logrus.Warn("Simple error server running...")
		select {}
	}
}

func controllerMain(conf config.Config) error {
	var err error

	logrus.ErrorKey = "error"
	logrus.Info(conf.LogLevel)
	logrus.SetLevel(conf.LogLevel)

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

	var tenantDB database.Tenant
	if conf.Database.Type == "memory" {

		db := database.NewMemoryCloudantDB()
		tenantDB = database.NewTenant(db)
	} else {
		err = errors.New("unsupported database type")
		setupHandler.SetError(err)
		return err
	}

	tpc := notification.NewTenantProducerCache()

	g, err := nginx.NewGenerator(nginx.Config{
		Path:     "./nginx/nginx.conf.tmpl",
		Database: tenantDB,
	})
	if err != nil {
		logrus.Error(err)
		setupHandler.SetError(err)
		return err
	}

	r := manager.NewManager(manager.Config{
		Database:      tenantDB,
		ProducerCache: tpc,
		Generator:     g,
	})

	factory := checker.NewRegistryFactory()

	c := checker.New(checker.Config{
		Database:      tenantDB,
		ProducerCache: tpc,
		Generator:     g,
		Factory:       factory,
	})

	n := api.NewNGINX(api.NGINXConfig{
		Reporter:  reporter,
		Generator: g,
	})
	t := api.NewTenant(api.TenantConfig{
		Reporter: reporter,
		Manager:  r,
	})
	p := api.NewPoll(reporter, c)
	h := api.NewHealth(reporter)

	a := rest.NewApi()
	a.Use(
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.RecoverMiddleware{
			EnableResponseStackTrace: false,
		},
		&rest.ContentTypeCheckerMiddleware{},
		&middleware.RequestIDMiddleware{},
		&middleware.AuthMiddleware{
			Auth: &middleware.LocalAuth{},
			Key:  conf.ControlToken,
		},
		&middleware.LoggingMiddleware{},
	)

	routes := n.Routes()
	routes = append(routes, t.Routes()...)
	routes = append(routes, h.Routes()...)
	routes = append(routes, p.Routes()...)

	router, err := rest.MakeRouter(
		routes...,
	)
	if err != nil {
		setupHandler.SetError(err)
		return err
	}
	a.SetApp(router)

	setupHandler.SetHandler(a.MakeHandler())

	//start garbage collection on kafka producer cache
	tpc.StartGC()

	// Server is already started
	logrus.WithFields(logrus.Fields{
		"port": conf.APIPort,
	}).Info("Server started")
	if conf.PollInterval.Seconds() != 0.0 {
		logrus.Info("Beginning periodic poll...")
		ticker := time.NewTicker(conf.PollInterval)
		for {
			select {
			case <-ticker.C:
				logrus.Debug("Polling")
				if err = c.Check(nil); err != nil {
					logrus.WithError(err).Error("Periodic poll failed")
				}
			}
		}
	} else {
		select {}
	}
}
