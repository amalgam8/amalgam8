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
		"err":       err,
		"metric_id": id,
		"time":      time.String(),
	}).Error("Metric recorded failure")
	return nil
}

func (l *logger) Success(id string, time time.Duration) error {
	//statsdClient.Inc(id+"CountSuccess", 1, 1.0)
	//statsdClient.TimingDuration(name+"ResponseTimeSuccess", endTime, 1.0)
	logrus.WithFields(logrus.Fields{
		"metric_id": id,
		"time":      time.String(),
	}).Debug("Metric recorded success")
	return nil
}
