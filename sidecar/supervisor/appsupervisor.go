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

// AppSupervisor manages process in sidecar
type AppSupervisor struct {
	agent     *register.RegistrationAgent
	processes []*process
}

type process struct {
	Cmd    *exec.Cmd
	Action string
}

// NewAppSupervisor builds new AppSupervisor using Commands in Config object
func NewAppSupervisor(conf *config.Config) *AppSupervisor {
	a := AppSupervisor{
		processes: []*process{},
	}

	for _, cmd := range conf.Commands {
		osCmd := exec.Command(cmd.Cmd[0], cmd.Cmd[1:]...)

		osCmd.Stdout = os.Stdout
		osCmd.Stderr = os.Stderr

		osCmd.Env = append(cmd.Env, os.Environ()...)
		proc := &process{
			Cmd:    osCmd,
			Action: cmd.OnExit,
		}
		if proc.Action == "" {
			proc.Action = config.IgnoreProcess
		}

		a.processes = append(a.processes, proc)
	}

	return &a
}

type processError struct {
	Err  error
	Proc *process
}

// DoAppSupervision starts subprocesses and manages their lifecycle - exiting if necessary
func (a *AppSupervisor) DoAppSupervision(agent *register.RegistrationAgent) {

	a.agent = agent

	appChan := make(chan processError, len(a.processes))
	for _, proc := range a.processes {
		log.Infof("Launching app '%s' with args '%s'", proc.Cmd.Args[0], strings.Join(proc.Cmd.Args[1:], " "))
		err := proc.Cmd.Start()
		if err != nil {
			appChan <- processError{
				Err:  err,
				Proc: proc,
			}
		} else {
			go func(proc *process) {
				err := proc.Cmd.Wait()
				appChan <- processError{
					Err:  err,
					Proc: proc,
				}
			}(proc)
		}
	}

	// Intercept SIGTERM/SIGINT and stop
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case sig := <-sigChan:
			log.Infof("Intercepted signal '%s'", sig)

			// forwarding signal to supervised applications to exit gracefully
			terminateSubprocesses(a.processes, sig)

			a.Shutdown(0)
		case err := <-appChan:
			exitCode := 0
			if err.Err == nil {
				log.Info("App terminated with exit code 0")
			} else {
				exitCode = 1
				if exitErr, ok := err.Err.(*exec.ExitError); ok {
					if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); ok {
						exitCode = waitStatus.ExitStatus()
					}
					log.Errorf("App terminated with exit code %d", exitCode)
				} else {
					log.Errorf("App failed to start: %v", err)
				}
			}

			switch err.Proc.Action {
			case config.IgnoreProcess:
				//Ignore this dead process
				log.WithError(err.Err).Warn("App '%s' with args '%s' exited with error.  Ignoring", err.Proc.Cmd.Args[0], strings.Join(err.Proc.Cmd.Args[1:], " "))
				err.Proc.Cmd.Wait()

			case config.TerminateProcess:
				log.WithError(err.Err).Error("App '%s' with args '%s' exited with error.  Exiting", err.Proc.Cmd.Args[0], strings.Join(err.Proc.Cmd.Args[1:], " "))
				terminateSubprocesses(a.processes, syscall.SIGTERM)
				a.Shutdown(exitCode)
			}

		}
	}
}

// Shutdown deregister the app with registry and exit sidecar
func (a *AppSupervisor) Shutdown(sig int) {

	if a.agent != nil {
		a.agent.Stop()
	}

	log.Infof("Shutting down with exit code %v", sig)
	os.Exit(sig)
}

func terminateSubprocesses(procs []*process, sig os.Signal) {

	var wg sync.WaitGroup

	wg.Add(len(procs))
	for _, proc := range procs {
		go func(cmd *exec.Cmd) {
			defer wg.Done()
			timer := time.AfterFunc(3*time.Second, func() {
				cmd.Process.Kill()
			})
			cmd.Process.Signal(sig)

			cmd.Wait()
			timer.Stop()
		}(proc.Cmd)
	}

	wg.Wait()
}
