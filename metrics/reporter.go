package metrics

import (
	"time"

	"github.com/Sirupsen/logrus"
)

// Reporter interface for metric recording
// can use this interface to implement statsd metric reporting
type Reporter interface {
	Failure(id string, endTime time.Duration, err error) error
	Success(id string, endTime time.Duration) error
}

type logger struct{}

// NewReporter basic Reporter implementation that logs succes and failure
func NewReporter() Reporter {
	return &logger{}
}

func (l *logger) Failure(id string, time time.Duration, err error) error {
	//statsdClient.Inc(name+"CountFailure", 1, 1.0)
	//statsdclient.TimingDuration(id+"ResponseTimeFailure", endTime, 1.0)
	logrus.WithFields(logrus.Fields{
		"err":  err,
		"id":   id,
		"time": time.String(),
	}).Error("Metric recorded failure")
	return nil
}

func (l *logger) Success(id string, time time.Duration) error {
	//statsdClient.Inc(id+"CountSuccess", 1, 1.0)
	//statsdClient.TimingDuration(name+"ResponseTimeSuccess", endTime, 1.0)
	logrus.WithFields(logrus.Fields{
		"id":   id,
		"time": time.String(),
	}).Debug("Metric recorded success")
	return nil
}
