package main

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/api"
	"github.com/amalgam8/controller/checker"
	"github.com/amalgam8/controller/clients"
	"github.com/amalgam8/controller/config"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/controller/middleware"
	"github.com/amalgam8/controller/nginx"
	"github.com/amalgam8/controller/notification"
	"github.com/amalgam8/controller/proxyconfig"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/codegangsta/cli"
	"net/http"
	"os"
	"time"
)

const (
	statsdPrefix = "sp-XXXXXXX"
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

	statsdClient, err := statsd.NewBufferedClient(conf.StatsdHost, statsdPrefix, time.Duration(0), 0)
	if err != nil {
		logrus.WithError(err)
		setupHandler.SetError(err)
		return err
	}

	var catalogDB database.Catalog
	var rulesDB database.Rules
	if conf.Database.Type == "memory" {
		db := database.NewMemoryCloudantDB()
		catalogDB = database.NewCatalog(db)

		db = database.NewMemoryCloudantDB()
		rulesDB = database.NewRules(db)
	} else {
		setupHandler.SetError(err)
		return errors.New("")
	}

	registry := clients.NewRegistry()

	tpc := notification.NewTenantProducerCache()

	r := proxyconfig.NewManager(proxyconfig.Config{
		Database:      rulesDB,
		ProducerCache: tpc,
	})

	c := checker.New(checker.Config{
		Database:      catalogDB,
		ProxyConfig:   r,
		Registry:      registry,
		ProducerCache: tpc,
	})

	g, err := nginx.NewGenerator(nginx.Config{
		Path:         "./nginx/nginx.conf.tmpl",
		Catalog:      c,
		ProxyManager: r,
	})
	if err != nil {
		logrus.Error(err)
		setupHandler.SetError(err)
		return err
	}

	n := api.NewNGINX(api.NGINXConfig{
		Statsd:    statsdClient,
		Generator: g,
		Checker:   c,
	})

	t := api.NewTenant(api.TenantConfig{
		Statsd:      statsdClient,
		Checker:     c,
		ProxyConfig: r,
	})

	p := api.NewPoll(statsdClient, c)

	h := api.NewHealth(statsdClient)

	routes := n.Routes()
	routes = append(routes, t.Routes()...)
	routes = append(routes, h.Routes()...)
	routes = append(routes, p.Routes()...)

	api := rest.NewApi()

	api.Use(
		&rest.TimerMiddleware{},
		&rest.RecorderMiddleware{},
		&rest.RecoverMiddleware{
			EnableResponseStackTrace: false,
		},
		&rest.ContentTypeCheckerMiddleware{},
		&middleware.RequestIDMiddleware{},
		&middleware.LoggingMiddleware{},
	)

	router, err := rest.MakeRouter(
		routes...,
	)
	if err != nil {
		setupHandler.SetError(err)
		return err
	}
	api.SetApp(router)

	setupHandler.SetHandler(api.MakeHandler())

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
