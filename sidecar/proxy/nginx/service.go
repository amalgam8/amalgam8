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

package nginx

import (
	"os"
	"os/exec"
	"sync"
	"time"

	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/sidecar/identity"
)

// MinimumRestartWait is the minimum amount of time to allow between restarts of the NGINX service.
const MinimumRestartWait = 15 * time.Second

// ServiceEnvironmentVariable is the NGINX environment variable
// specifying the service and tags used for source identification.
const ServiceEnvironmentVariable = "A8_SERVICE"

// Service maintains a NGINX service.
type Service interface {
	Start() error
	Stop() error
}

// NewService creates new instance.
func NewService(identity identity.Provider) Service {
	return &service{
		identity: identity,
		stop:     make(chan struct{}),
	}
}

type service struct {
	identity identity.Provider
	running  bool
	stop     chan struct{}
	mutex    sync.Mutex
}

// Start maintaining the NGINX service.
func (s *service) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return nil
	}

	cmd, err := s.build()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	s.running = true
	go s.maintain(cmd)

	return nil
}

// build the NGINX service command.
func (s *service) build() (*exec.Cmd, error) {
	identity, err := s.identity.GetIdentity()
	if err != nil {
		return nil, err
	}

	serviceEnvVar := fmt.Sprintf("%s=%s:%s", ServiceEnvironmentVariable,
		identity.ServiceName, strings.Join(identity.Tags, ","))

	cmd := exec.Command("nginx", "-g", "daemon off;")
	cmd.Env = append(os.Environ(), serviceEnvVar)

	return cmd, nil
}

// maintain the NGINX service. Automatically restart the service if it exits.
func (s *service) maintain(cmd *exec.Cmd) {
	start := time.Now()
	status := make(chan error)
	go func() {
		status <- cmd.Wait()
	}()

	select {
	case err := <-status:
		if err != nil {
			logrus.WithError(err).Error("NGINX exited with error")
		} else {
			logrus.Error("NGINX exited")
		}

		// Ensure that we always wait at least the minimum amount of time between restarts.
		delta := time.Now().Sub(start)
		if delta < MinimumRestartWait {
			time.Sleep(MinimumRestartWait - delta)
		}

		// Restart NGINX.
		go s.restart()
	case <-s.stop:
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				logrus.WithError(err).Error("NGINX did not terminate cleanly")
			}
		}
	}
}

// restart the NGINX service. Retry indefinitely on restart failure. On restart success maintain the service.
func (s *service) restart() {
	var cmd *exec.Cmd
	var err error
	for {
		cmd, err = s.build()
		if err == nil {
			err = cmd.Start()
			if err == nil {
				break
			}
		}

		logrus.WithError(err).Error("NGINX failed to start")
		select {
		case <-time.After(MinimumRestartWait):
			continue
		case <-s.stop:
			return
		}
	}
	go s.maintain(cmd)
}

// Stop maintaining the NGINX service and terminate the running NGINX service.
func (s *service) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return nil
	}

	s.stop <- struct{}{}
	s.running = false

	return nil
}
