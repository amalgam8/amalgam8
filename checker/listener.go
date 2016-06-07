package checker

import (
	"bytes"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/sidecar/clients"
	"github.com/amalgam8/sidecar/config"
	"github.com/amalgam8/sidecar/nginx"
)

// Listener listens for events from Message Hub and updates NGINX
type Listener interface {
	Start() error
	Stop() error
}

type listener struct {
	config     *config.Config
	consumer   Consumer
	nginx      nginx.Nginx
	controller clients.Controller
}

// NewListener new Listener implementation
func NewListener(config *config.Config, consumer Consumer, c clients.Controller, nginx nginx.Nginx) Listener {
	return &listener{
		config:     config,
		consumer:   consumer,
		nginx:      nginx,
		controller: c,
	}
}

// Start listens messages to arrive
func (l *listener) Start() error {
	logrus.Info("Listening for messages")
	for {
		err := l.listenForUpdate()
		if err != nil {
			logrus.WithError(err).Error("Update failed")
		}
	}
}

// ListenForUpdate sleeps until an event indicating that the rules for this tenant have
// changed. Once the event occurs we attempt to update our configuration.
func (l *listener) listenForUpdate() error {

	// Sleep until we receive an event indicating that the our rules have changed
	for {
		key, value, err := l.consumer.ReceiveEvent()
		if err != nil {
			logrus.WithError(err).Error("Couldn't read from Kafka bus")
			return err
		}

		if key == l.config.Tenant.ID {
			logrus.WithFields(logrus.Fields{
				"key":   key,
				"value": value,
			}).Info("Tenant event received")
			break
		}
	}

	// Get latest config from Controller
	conf, err := l.controller.GetNGINXConfig(nil)
	if err != nil {
		logrus.WithError(err).Error("Call to Controller failed")
		return err
	}
	//since version="", Controller should never return empty string

	reader := bytes.NewBufferString(conf)

	// Update our existing NGINX config
	if err := l.nginx.Update(reader); err != nil {
		logrus.WithError(err).Error("Could not update NGINX config")
		return err
	}

	return nil
}

// Stop do any necessary cleanup
func (l *listener) Stop() error {
	return nil
}
