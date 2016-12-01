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
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/sidecar/config"
	"github.com/amalgam8/amalgam8/sidecar/register"
)

// AppSupervisor manages process in sidecar
type AppSupervisor struct {
	registration register.Lifecycle
	processes    []*process
}

type process struct {
	Cmd       *exec.Cmd
	Action    string
	KillGroup bool
}

// NewAppSupervisor builds new AppSupervisor using Commands in Config object
func NewAppSupervisor(conf *config.Config, registration register.Lifecycle) *AppSupervisor {
	a := AppSupervisor{
		registration: registration,
		processes:    []*process{},
	}

	for _, cmd := range conf.Commands {
		osCmd := exec.Command(cmd.Cmd[0], cmd.Cmd[1:]...)

		osCmd.Stdout = os.Stdout
		osCmd.Stderr = os.Stderr

		// Enable setting process's group ID so we can kill this process
		// and all of its children (if any) if `kill_group` flag is set
		osCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		osCmd.Env = append(cmd.Env, os.Environ()...)
		proc := &process{
			Cmd:       osCmd,
			Action:    cmd.OnExit,
			KillGroup: cmd.KillGroup,
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
func (a *AppSupervisor) DoAppSupervision() {
	appChan := make(chan processError, len(a.processes))
	for _, proc := range a.processes {
		log.Infof("Launching app '%v' with args '%v'", proc.Cmd.Args[0], strings.Join(proc.Cmd.Args[1:], " "))
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
			log.Infof("Intercepted signal '%v'", sig)

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
					log.Errorf("App terminated with exit code %v", exitCode)
				} else {
					log.Errorf("App failed to start: %v", err)
				}
			}

			switch err.Proc.Action {
			case config.IgnoreProcess:
				//Ignore this dead process
				log.WithError(err.Err).Warnf("App '%v' with args '%v' exited.  Ignoring", err.Proc.Cmd.Args[0], strings.Join(err.Proc.Cmd.Args[1:], " "))

			case config.TerminateProcess:
				log.WithError(err.Err).Errorf("App '%v' with args '%v' exited.  Exiting", err.Proc.Cmd.Args[0], strings.Join(err.Proc.Cmd.Args[1:], " "))
				terminateSubprocesses(a.processes, syscall.SIGTERM)
				a.Shutdown(exitCode)
			}

		}
	}
}

// Shutdown deregister the app with registry and exit sidecar
func (a *AppSupervisor) Shutdown(sig int) {
	if a.registration != nil {
		a.registration.Stop()
	}

	log.Infof("Shutting down with exit code %v", sig)
	os.Exit(sig)
}

func terminateSubprocesses(procs []*process, sig os.Signal) {
	for _, proc := range procs {
		if proc.Cmd.Process != nil {

			// If enabled, send the signal to the child and all of its
			// own children (using process group ID), else send just to
			// the immediate child
			if proc.KillGroup {
				syscall.Kill(-proc.Cmd.Process.Pid, sig.(syscall.Signal))
			} else {
				proc.Cmd.Process.Signal(sig)
			}

		}
	}

	timeout := time.Now().Add(3 * time.Second)
	alive := len(procs)
	for alive > 0 {
		alive = len(procs)
		for _, proc := range procs {
			if proc.Cmd.Process != nil {
				if e := syscall.Kill(proc.Cmd.Process.Pid, syscall.Signal(0)); e != nil {
					if e == syscall.ESRCH {
						alive--
					}
				}
			}
		}

		// If services have not exited gracefully within the alotted time, force kill them all
		if time.Now().After(timeout) {
			for _, proc := range procs {
				if proc.Cmd.Process != nil {
					if proc.KillGroup {
						syscall.Kill(-proc.Cmd.Process.Pid, syscall.SIGKILL)
					} else {
						proc.Cmd.Process.Kill()
					}
				}
			}
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// reapZombies cleans up any zombies sidecar may have inherited from terminated children
// - on SIGCHLD send wait4() (ref http://linux.die.net/man/2/waitpid)
func ReapZombies() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGCHLD)

	for range sigChan {
		for {
			_, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)

			// Spurious wakeup
			if err == syscall.EINTR {
				continue
			}
			log.Debug("Zombie reaped")
			// Done
			break
		}
	}
}
