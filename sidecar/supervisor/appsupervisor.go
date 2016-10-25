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

package supervisor

import (
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/register"
)

// AppSupervisor TODO
type AppSupervisor struct {
	agent   *register.RegistrationAgent
	app     *exec.Cmd
	helpers []*exec.Cmd
}

// NewAppSupervisor TODO
func NewAppSupervisor(conf *config.Config) *AppSupervisor {
	a := AppSupervisor{
		helpers: []*exec.Cmd{},
	}

	for _, cmd := range conf.Commands {
		osCmd := exec.Command(cmd.Cmd[0], cmd.Cmd[1:]...)

		osCmd.Stdin = os.Stdin
		osCmd.Stdout = os.Stdout
		osCmd.Stderr = os.Stderr

		cmdEnv := os.Environ()
		if cmd.Env != nil {
			cmdEnv = append(cmdEnv, cmd.Env...)
		}
		osCmd.Env = cmdEnv

		if cmd.Primary {
			a.app = osCmd
		} else {
			a.helpers = append(a.helpers, osCmd)
		}
	}

	return &a
}

// DoAppSupervision TODO
func (a *AppSupervisor) DoAppSupervision(agent *register.RegistrationAgent) {

	a.agent = agent

	appChan := make(chan error, 1)
	if a.app != nil {
		log.Infof("Launching app '%s' with args '%s'", a.app.Args[0], strings.Join(a.app.Args[1:], " "))
		err := a.app.Start()
		if err != nil {
			appChan <- err
		} else {
			go func() {
				appChan <- a.app.Wait()
			}()
		}
	}

	for _, cmd := range a.helpers {
		log.Infof("Launching app '%s' with args '%s'", cmd.Args[0], strings.Join(cmd.Args[1:], " "))
		go func(cmd *exec.Cmd) {
			err := cmd.Run()
			log.WithError(err).Warn("Failed to launch helper command '%s'", cmd.Args[0])
		}(cmd)
	}

	// Intercept SIGTERM/SIGINT and stop
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case sig := <-sigChan:
			log.Infof("Intercepted signal '%s'", sig)

			// forwarding signal to application parent process
			exit(append(a.helpers, a.app), sig)

			a.Shutdown(0)
		case err := <-appChan:
			exitCode := 0
			if err == nil {
				log.Info("App terminated with exit code 0")
			} else {
				exitCode = 1
				if exitErr, ok := err.(*exec.ExitError); ok {
					if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); ok {
						exitCode = waitStatus.ExitStatus()
					}
					log.Errorf("App terminated with exit code %d", exitCode)
				} else {
					log.Errorf("App failed to start: %v", err)
				}
			}
			exit(a.helpers, syscall.SIGKILL)
			a.Shutdown(exitCode)
		}
	}
}

// Shutdown TODO
func (a *AppSupervisor) Shutdown(sig int) {

	if a.agent != nil {
		a.agent.Stop()
	}

	log.Infof("Shutting down with exit code %v", sig)
	os.Exit(sig)
}

func exit(cmds []*exec.Cmd, sig os.Signal) {

	var wg sync.WaitGroup

	wg.Add(len(cmds))
	for _, cmd := range cmds {
		go func(cmd *exec.Cmd) {
			defer wg.Done()
			timer := time.AfterFunc(3*time.Second, func() {
				cmd.Process.Kill()
			})
			cmd.Process.Signal(sig)

			cmd.Wait()
			timer.Stop()
		}(cmd)
	}

	wg.Wait()
}
