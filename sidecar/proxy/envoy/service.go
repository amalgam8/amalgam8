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

package envoy

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"time"

	"github.com/Sirupsen/logrus"
)

// Service manages a proxy service.
type Service interface {
	Reload() error
}

type service struct {
	config             string            // Envoy config filename.
	cmdMap             map[*exec.Cmd]int // Map of known running Envoy processes and their restart epochs.
	drainTime          int
	parentShutdownTime int
	mutex              sync.Mutex
}

// ServiceConfig config for Service
type ServiceConfig struct {
	EnvoyBinary               string
	EnvoyConfig               string
	DrainTimeSeconds          int
	ParentShutdownTimeSeconds int
}

// NewService creates new instance.
func NewService(config ServiceConfig) Service {
	return &service{
		binary:             config.EnvoyBinary,
		config:             config.EnvoyConfig,
		drainTime:          config.DrainTimeSeconds,
		parentShutdownTime: config.ParentShutdownTimeSeconds,
		cmdMap:             make(map[*exec.Cmd]int),
	}
}

// Reload Envoy with a hot restart. Envoy hot restarts are performed by launching a new Envoy process with an
// incremented restart epoch. To successfully launch a new Envoy process that will replace the running Envoy processes,
// the restart epoch of the new process must be exactly 1 greater than the highest restart epoch of the currently
// running Envoy processes. To ensure that we launch the new Envoy process with the correct restart epoch, we keep track
// of all running Envoy processes and their restart epochs.
//
// Envoy hot restart documentation: https://lyft.github.io/envoy/docs/intro/arch_overview/hot_restart.html
func (s *service) Reload() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find highest restart epoch of the known running Envoy processes.
	restartEpoch := -1
	for _, epoch := range s.cmdMap {
		if epoch > restartEpoch {
			restartEpoch = epoch
		}
	}
	restartEpoch++

	// Spin up a new Envoy process.
	cmd := exec.Command(s.binary,
		"-c", s.config,
		"--drain-time-s", fmt.Sprint(s.drainTime),
		"--parent-shutdown-time-s", fmt.Sprint(s.parentShutdownTime),
		"--restart-epoch", fmt.Sprint(restartEpoch))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	// Add the new Envoy process to the known set of running Envoy processes.
	s.cmdMap[cmd] = restartEpoch

	// Start tracking the process.
	go s.waitForExit(cmd)

	time.Sleep(256 * time.Millisecond)

	return nil
}

// waitForExit waits until the command exits and removes it from the set of known running Envoy processes.
func (s *service) waitForExit(cmd *exec.Cmd) {
	if err := cmd.Wait(); err != nil {
		logrus.Debugf("Envoy terminated: %v", err.Error())
	} else {
		logrus.Debug("Envoy process exited")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.cmdMap, cmd)
}
