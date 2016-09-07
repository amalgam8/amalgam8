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
	"github.com/amalgam8/amalgam8/registry/utils/logging"
	gometrics "github.com/rcrowley/go-metrics"
)

// interval at which the metrics registry is dumped
const dumpInterval = 10 * time.Minute

// module name to be used in logging
const moduleName = "METRICS"

var logger = logging.GetLogger(moduleName)

// DumpPeriodically logs the values of the entire go-metrics registry, periodically.
// This function blocks, so should be called within a separate goroutine.
func DumpPeriodically() {
	dumpPeriodically(dumpInterval, gometrics.DefaultRegistry)
}

func dumpPeriodically(interval time.Duration, registry gometrics.Registry) {
	for range time.Tick(interval) {
		dumpRegistry(registry)
	}
}

func dumpRegistry(registry gometrics.Registry) {
	logger.Info("Dumping metrics registry")
	registry.Each(func(name string, metric interface{}) {
		dumpMetric(name, metric)
	})
}

func dumpMetric(name string, metric interface{}) {
	switch metric := metric.(type) {
	case gometrics.Counter:
		dumpCounter(name, metric)
	case gometrics.Gauge:
		dumpGauge(name, metric)
	case gometrics.GaugeFloat64:
		dumpGaugeFloat64(name, metric)
	case gometrics.Meter:
		dumpMeter(name, metric)
	case gometrics.Histogram:
		dumpHistogram(name, metric)
	case gometrics.Timer:
		dumpTimer(name, metric)
	}
}

func dumpCounter(name string, metric gometrics.Counter) {
	logger.WithFields(logrus.Fields{
		"name":  name,
		"count": metric.Count(),
	}).Info()
}

func dumpGauge(name string, metric gometrics.Gauge) {
	logger.WithFields(logrus.Fields{
		"name":  name,
		"value": metric.Value(),
	}).Info()
}

func dumpGaugeFloat64(name string, metric gometrics.GaugeFloat64) {
	logger.WithFields(logrus.Fields{
		"name":  name,
		"value": metric.Value(),
	}).Info()
}

func dumpMeter(name string, metric gometrics.Meter) {
	m := metric.Snapshot()
	logger.WithFields(logrus.Fields{
		"name":                name,
		"count":               m.Count(),
		"rate-one-minute":     m.Rate1(),
		"rate-five-minute":    m.Rate5(),
		"rate-fifteen-minute": m.Rate15(),
		"rate-mean":           m.RateMean(),
	}).Info()
}

func dumpHistogram(name string, metric gometrics.Histogram) {
	m := metric.Snapshot()
	logger.WithFields(logrus.Fields{
		"name":     name,
		"count":    m.Count(),
		"sum":      m.Sum(),
		"min":      m.Min(),
		"max":      m.Max(),
		"mean":     m.Mean(),
		"stddev":   m.StdDev(),
		"variance": m.Variance(),
	}).Info()
}

func dumpTimer(name string, metric gometrics.Timer) {
	m := metric.Snapshot()
	logger.WithFields(logrus.Fields{
		"name":                name,
		"count":               m.Count(),
		"sum":                 m.Sum(),
		"min":                 m.Min(),
		"max":                 m.Max(),
		"mean":                m.Mean(),
		"stddev":              m.StdDev(),
		"variance":            m.Variance(),
		"rate-one-minute":     m.Rate1(),
		"rate-five-minute":    m.Rate5(),
		"rate-fifteen-minute": m.Rate15(),
		"rate-mean":           m.RateMean(),
	}).Info()
}
