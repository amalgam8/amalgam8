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

package healthcheck

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/sidecar/config"
)

const (
	defaultCommandTimeout  = 5 * time.Second
	defaultCommandInterval = 30 * time.Second
)

// Command performs a health check by running a command and inspecting the results.
type Command struct {
	cmd      string
	args     []string
	timeout  time.Duration
	exitCode int
}

// NewCommand creates a new executable health check.
func NewCommand(conf config.HealthCheck) (Check, error) {
	if err := validateCommandConfig(&conf); err != nil {
		return nil, err
	}

	return &Command{
		cmd:      conf.Value,
		args:     conf.Args,
		exitCode: conf.Code,
		timeout:  conf.Timeout,
	}, nil
}

// validateCommandConfig validates, sanitizes, and sets defaults for a command health check configuration.
func validateCommandConfig(conf *config.HealthCheck) error {
	if conf.Type != config.CommandHealthCheck {
		return fmt.Errorf("invalid type for a command healthcheck: '%s'", conf.Type)
	}

	if conf.Timeout == 0 {
		conf.Timeout = defaultCommandTimeout
	}

	if conf.Interval == 0 {
		conf.Interval = defaultCommandInterval
	}

	if conf.Value == "" {
		return fmt.Errorf("command path required")
	}

	return nil
}

// Execute the command and check the results.
// TODO: return any of stdout, stderr?
func (c *Command) Execute() error {
	cmd := exec.Command(c.cmd, c.args...)
	cmd.Stdout = ioutil.Discard // Consume any output
	cmd.Stderr = ioutil.Discard
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					exitCode := waitStatus.ExitStatus()
					if exitCode != c.exitCode {
						return fmt.Errorf("%s returned %v, expected %v", c.cmd, exitCode, c.exitCode)
					}
				}
			}

			return fmt.Errorf("%s failed: %v", c.cmd, err.Error())
		}
	case <-time.After(c.timeout):
		if err := cmd.Process.Kill(); err != nil {
			logrus.WithError(err).Warn("Error killing process")
		}
		return fmt.Errorf("%s timed out", c.cmd)
	}

	return nil
}
