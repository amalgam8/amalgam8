package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/sidecar/config"
	"github.com/amalgam8/sidecar/register"
	"github.com/amalgam8/sidecar/router/checker"
	"github.com/amalgam8/sidecar/router/clients"
	"github.com/amalgam8/sidecar/router/nginx"
	"github.com/amalgam8/sidecar/supervisor"
	"github.com/codegangsta/cli"
)

func main() {
	// TODO: make configurable
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(os.Stderr)

	app := cli.NewApp()

	app.Name = "sidecar"
	app.Usage = "Amalgam8 Sidecar"
	app.Version = "0.1"
	app.Flags = config.TenantFlags
	app.Action = sidecarCommand

	args := os.Args
	index := -1
	for i, arg := range os.Args {
		if arg == "--supervise" {
			index = i
			break
		}
	}
	if index != -1 {
		args = args[0 : index+1]
	}

	err := app.Run(args)
	if err != nil {
		logrus.WithError(err).Error("Failure running main")
	}
}

func sidecarCommand(context *cli.Context) {
	conf := config.New(context)
	if err := sidecarMain(*conf); err != nil {
		logrus.WithError(err).Error("Setup failed")
	}
}

func sidecarMain(conf config.Config) error {
	var err error

	logrus.SetLevel(conf.LogLevel)

	if err = conf.Validate(); err != nil {
		logrus.WithError(err).Error("Validation of config failed")
		return err
	}

	if conf.Log {
		//Replace the LOGSTASH_REPLACEME string in filebeat.yml with
		//the value provided by the user

		//TODO: Make this configurable
		filebeatConf := "/etc/filebeat/filebeat.yml"
		filebeat, err := ioutil.ReadFile(filebeatConf)
		if err != nil {
			logrus.WithError(err).Error("Could not read filebeat conf")
			return err
		}

		fileContents := strings.Replace(string(filebeat), "LOGSTASH_REPLACEME", conf.LogstashServer, -1)

		err = ioutil.WriteFile("/tmp/filebeat.yml", []byte(fileContents), 0)
		if err != nil {
			logrus.WithError(err).Error("Could not write filebeat conf")
			return err
		}

		// TODO: Log failure?
		go supervisor.DoLogManagement("/tmp/filebeat.yml")
	}

	if conf.Proxy {
		go func() {
			if err = startProxy(&conf); err != nil {
				logrus.WithError(err).Error("Could not start proxy")
			}
		}()
	}

	if conf.Register {
		logrus.Info("Registering")
		register.DoServiceRegistrationAndHeartbeat(&conf, true)
	}

	if conf.Supervise {
		// TODO: do this more cleanly
		args := os.Args
		index := -1
		for i, arg := range os.Args {
			if arg == "--supervise" {
				index = i
				break
			}
		}
		if index != -1 {
			args = args[0 : index+1]
		}

		supervisor.DoAppSupervision(args)
	} else {
		select {}
	}

	return nil
}

func startProxy(conf *config.Config) error {
	var err error

	rc := clients.NewController(conf)

	// FIXME: hack until we add this as an argument (right now we are passing in "serviceName:version")
	serviceName := strings.Split(conf.ServiceName, ":")
	nginx := nginx.NewNginx(serviceName[0])

	err = checkIn(rc, conf)
	if err != nil {
		logrus.WithError(err).Error("Check in failed")
		return err
	}

	// for Kafka enabled tenants we should do both polling and listening
	if len(conf.Kafka.Brokers) != 0 {
		go func() {
			time.Sleep(time.Second * 10)
			logrus.Info("Attempting to connect to Kafka")
			var consumer checker.Consumer
			for {
				consumer, err = checker.NewConsumer(checker.ConsumerConfig{
					Brokers:     conf.Kafka.Brokers,
					Username:    conf.Kafka.Username,
					Password:    conf.Kafka.Password,
					ClientID:    conf.Kafka.APIKey,
					Topic:       "NewRules",
					SASLEnabled: conf.Kafka.SASL,
				})
				if err != nil {
					logrus.WithError(err).Error("Could not connect to Kafka, trying again . . .")
					time.Sleep(time.Second * 5) // TODO: exponential falloff?
				} else {
					break
				}
			}
			logrus.Info("Successfully connected to Kafka")

			listener := checker.NewListener(conf, consumer, rc, nginx)

			// listen to Kafka indefinitely
			if err := listener.Start(); err != nil {
				logrus.WithError(err).Error("Could not listen to Kafka")
			}
		}()
	}

	poller := checker.NewPoller(conf, rc, nginx)
	if err = poller.Start(); err != nil {
		logrus.WithError(err).Error("Could not poll Controller")
		return err
	}

	return nil
}

func checkIn(rc clients.Controller, conf *config.Config) error {
	creds, err := rc.GetCredentials()
	if err != nil {
		// TODO If controller returns 404, we need to register
		if strings.Contains(err.Error(), "ID not found") {
			logrus.Info("ID not found, registering with controller")
			// TODO validate creds before registering?
			if err = rc.Register(); err != nil {
				// TODO if register returns 409, possible race condition, try to get creds again
				if strings.Contains(err.Error(), "ID already present") {
					logrus.Warn("Possible race condition occurred during register")
					creds, err = rc.GetCredentials()
					if err != nil {
						// no idea what state is, return error
						logrus.WithError(err).Error("ID is in unknown state with control plane")
						return err
					}
					logrus.Info("Successfully retrieved credentials")
					conf.Kafka.APIKey = creds.Kafka.APIKey
					conf.Kafka.Brokers = creds.Kafka.Brokers
					conf.Kafka.Password = creds.Kafka.Password
					conf.Kafka.RestURL = creds.Kafka.RestURL
					conf.Kafka.SASL = creds.Kafka.SASL
					conf.Kafka.Username = creds.Kafka.User

					conf.Registry.Token = creds.Registry.Token
					conf.Registry.URL = creds.Registry.URL
					return nil

				}
				// something other than 409 was returned
				logrus.WithError(err).Error("Could not register with Controller")
				return err

			}
			// register succeeded
			return nil
		}
		// get creds returned something other than 404
		logrus.WithError(err).Error("Could not retrieve credentials")
		return err
	}

	conf.Kafka.APIKey = creds.Kafka.APIKey
	conf.Kafka.Brokers = creds.Kafka.Brokers
	conf.Kafka.Password = creds.Kafka.Password
	conf.Kafka.RestURL = creds.Kafka.RestURL
	conf.Kafka.SASL = creds.Kafka.SASL
	conf.Kafka.Username = creds.Kafka.User

	conf.Registry.Token = creds.Registry.Token
	conf.Registry.URL = creds.Registry.URL
	return nil
}
