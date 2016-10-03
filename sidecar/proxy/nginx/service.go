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
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Service provides management operations for the NGINX service
type Service interface {
	Start() error
	Reload() error
	Running() (bool, error)
}

// NewService creates new instance
func NewService(name string) Service {
	return &service{
		name: name,
	}
}

type service struct {
	name string
}

// Start the NGINX service
func (s *service) Start() error {

	cmd := exec.Command("nginx", "-g", "daemon on;")
	cmdEnv := os.Environ()
	cmdEnv = append(cmdEnv, "A8_SERVICE="+s.name)
	cmd.Env = cmdEnv

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(err.Error() + ": " + string(out))
	}

	return nil
}

// Reload the NGINX service
func (s *service) Reload() error {
	out, err := exec.Command("nginx", "-s", "reload").CombinedOutput()
	if err != nil {
		return errors.New(err.Error() + ": " + string(out))
	}

	return nil
}

// Running indicates whether or not the NGINX service is currently running
func (s *service) Running() (bool, error) {
	pidBytes, err := ioutil.ReadFile("/var/run/nginx.pid")
	if err != nil {
		// Assume that service is not running
		return false, nil
	}

	pidString := strings.TrimSpace(string(pidBytes))

	// The command "kill -s 0 <pid>" has an exit code of 0 when the service is running and 1 when the service is
	// not running.
	out, err := exec.Command("kill", "-s", "0", pidString).CombinedOutput()
	if err != nil {
		// An error is returned from exec.Command when there is an issue executing the command OR if a non-zero
		// exit code is returned.

		// Check if the error is a non-zero exit code error
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				// Check if exit code is 1 (process is not running)
				if status.ExitStatus() == 1 {
					return false, nil
				}
			}
		}

		// Unknown error
		return false, err
	}

	// Output also indicates an error, even if the return code is 0
	if len(out) != 0 {
		return false, errors.New(string(out))
	}

	return true, nil
}
