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

package uptime

import (
	"math"
	"net/http"
	"syscall"
	"time"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/amalgam8/registry/utils/health"
	"github.com/amalgam8/amalgam8/registry/utils/version"
)

func uptimeHandler(w rest.ResponseWriter, r *rest.Request) {
	uHandler.ServeHTTP(w.(http.ResponseWriter), r.Request)
}

// healthyHandler implements an ever-healthy HTTP handler which always returns HTTP 200.
// This is used for several cloud environments in which a URL endpoint (normally, "/") is periodically polled
// to indicate healthness for an auto-recovery mechanism.
func healthyHandler(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusOK)
}

// uptimeHealthCheck implements an ever-healthy health.CheckerFunc that records the current process uptime and load.
func uptimeHealthCheck() health.Status {
	info := currentSysInfo()
	return health.StatusHealthyWithProperties(
		map[string]interface{}{
			"uptime":          info.Uptime.String(),
			"load_1m":         info.LastMinute,
			"load_5m":         info.LastFive,
			"load_15m":        info.LastFifteen,
			"build_version":   version.Build.Version,
			"build_revision":  version.Build.GitRevision,
			"build_date":      version.Build.BuildDate,
			"build_goversion": version.Build.GoVersion,
		})
}

type uptimeInfo struct {
	Uptime      time.Duration `json:"uptime"`
	LastMinute  float64       `json:"load_1m"`
	LastFive    float64       `json:"load_5m"`
	LastFifteen float64       `json:"load_15m"`
}

func currentSysInfo() *uptimeInfo {
	// see http://www.linuxquestions.org/questions/programming-9/%27load-average%27-return-values-from-sysinfo-309720/
	const loadScale = 65536.0 // magic conversion factor 2^16

	info := &uptimeInfo{}
	sysinfo := syscall.Sysinfo_t{}

	if err := syscall.Sysinfo(&sysinfo); err != nil { // just stop processing, values will be all zero's
		return info
	}

	info.Uptime = time.Now().UTC().Sub(startTime)
	info.LastMinute = fixedPrecision(float64(sysinfo.Loads[0])/loadScale, 3)
	info.LastFive = fixedPrecision(float64(sysinfo.Loads[1])/loadScale, 3)
	info.LastFifteen = fixedPrecision(float64(sysinfo.Loads[2])/loadScale, 3)
	return info
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func fixedPrecision(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

var startTime = time.Now().UTC()
var uHandler = health.Handler()

func init() {
	health.RegisterFunc("UPTIME", uptimeHealthCheck)
}
