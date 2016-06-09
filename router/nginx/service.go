package nginx

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"strings"
	"syscall"
)

// Service provides management operations for NGINX service
type Service interface {
	Start() error
	Reload() error
	Running() (bool, error)
}

// NewService creates new instance
func NewService() Service {
	return &service{}
}

type service struct {
}

// Start the NGINX service
func (s *service) Start() error {
	out, err := exec.Command("nginx", "-g", "daemon on;").CombinedOutput()
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

// Running returns whether the NGINX service is currently running
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

	// On error there is output
	if len(out) != 0 {
		return false, errors.New(string(out))
	}

	return true, nil
}
