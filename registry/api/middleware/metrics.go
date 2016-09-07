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

package middleware

import (
	"fmt"
	"time"

	"github.com/amalgam8/amalgam8/registry/api/env"
	"github.com/amalgam8/amalgam8/registry/api/protocol"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/rcrowley/go-metrics"
)

// MetricsMiddleware is an HTTP middleware that collects API usage metrics.
// It depends on the "ELAPSED_TIME" and "STATUS_CODE" being in r.Env (injected by rest.TimerMiddleware / rest.RecorderMiddleware),
// as well as the protocol.ProtocolKey and protocol.OperationKey values.
type MetricsMiddleware struct{}

// MiddlewareFunc implements the Middleware interface
func (mw *MetricsMiddleware) MiddlewareFunc(h rest.HandlerFunc) rest.HandlerFunc {
	return func(w rest.ResponseWriter, r *rest.Request) {
		h(w, r)
		mw.collectMetrics(w, r)
	}
}

func (mw *MetricsMiddleware) collectMetrics(w rest.ResponseWriter, r *rest.Request) {
	proto, ok := r.Env[env.APIProtocol].(protocol.Type)
	if !ok {
		return
	}

	operation, ok := r.Env[env.APIOperation].(protocol.Operation)
	if !ok {
		return
	}

	// Injected by TimerMiddleware
	latency, ok := r.Env[env.ElapsedTime].(*time.Duration)
	if !ok {
		logging.GetLogger(module).Error("could not find 'ELAPSED_TIME' parameter in HTTP request context")
		return
	}

	// Injected by RecorderMiddleware
	status, ok := r.Env[env.StatusCode].(int)
	if !ok {
		logging.GetLogger(module).Error("could not find 'STATUS_CODE' parameter in HTTP request context")
		return
	}

	histogramFactory := func() metrics.Histogram { return metrics.NewHistogram(metrics.NewExpDecaySample(256, 0.015)) }
	meterFactory := func() metrics.Meter { return metrics.NewMeter() }

	protocolName := protocol.NameOf(proto)
	operationName := operation.String()

	statusMeterName := fmt.Sprintf("api.%s.%s.status.%d", protocolName, operationName, status)
	statusMeter := metrics.DefaultRegistry.GetOrRegister(statusMeterName, meterFactory).(metrics.Meter)
	statusMeter.Mark(1)

	rateMeterName := fmt.Sprintf("api.%s.%s.rate", protocolName, operationName)
	rateMeter := metrics.DefaultRegistry.GetOrRegister(rateMeterName, meterFactory).(metrics.Meter)
	rateMeter.Mark(1)

	latencyHistogramName := fmt.Sprintf("api.%s.%s.latency", protocolName, operationName)
	latencyHistogram := metrics.DefaultRegistry.GetOrRegister(latencyHistogramName, histogramFactory).(metrics.Histogram)
	latencyHistogram.Update(int64(*latency))

	globalLatencyHistogramName := "api.global.latency"
	globalLatencyHistogram := metrics.DefaultRegistry.GetOrRegister(globalLatencyHistogramName, histogramFactory).(metrics.Histogram)
	globalLatencyHistogram.Update(int64(*latency))
}
