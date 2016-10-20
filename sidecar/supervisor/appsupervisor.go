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

	log "github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/sidecar/register"
)

// AppSupervisor TODO
type AppSupervisor struct {
	agent *register.RegistrationAgent
}

// DoAppSupervision TODO
func (a *AppSupervisor) DoAppSupervision(cmdArgs []string, agent *register.RegistrationAgent) {

	a.agent = agent

	// Launch the user app
	var appProcess *os.Process
	appChan := make(chan error, 1)
	if len(cmdArgs) > 0 {
		log.Infof("Launching app '%s' with args '%s'", cmdArgs[0], strings.Join(cmdArgs[1:], " "))

		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		if err != nil {
			appChan <- err
		} else {
			appProcess = cmd.Process
			go func() {
				appChan <- cmd.Wait()
			}()
		}
	}

	// Intercept SIGTERM/SIGINT and stop
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case sig := <-sigChan:
			log.Infof("Intercepted signal '%s'", sig)
			if appProcess != nil {
				appProcess.Signal(sig)
			} else {
				a.Shutdown(0)
			}
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

			a.Shutdown(exitCode)
		}
	}
}

// Shutdown TODO
func (a *AppSupervisor) Shutdown(exitCode int) {
	// TODO: Gracefully shutdown Ngnix, and filebeat
	//ugly temporary hack to kill all processes in container
	exec.Command("pkill", "-9", "-f", "nginx")
	exec.Command("pkill", "-9", "-f", "filebeat")

	if a.agent != nil {
		a.agent.Stop()
	}

	log.Infof("Shutting down with exit code %d", exitCode)
	os.Exit(exitCode)
}

// DoLogManagement TODO
func (a *AppSupervisor) DoLogManagement(filebeatConf string) {
	// starting filebeat
	logcmd := exec.Command("filebeat", "-c", filebeatConf)
	env := os.Environ()
	env = append(env, "GODEBUG=netdns=go")
	logcmd.Env = env

	logcmd.Stdin = os.Stdin
	logcmd.Stdout = os.Stdout
	err := logcmd.Run()
	if err != nil {
		log.WithError(err).Warn("Filebeat process exited with error")
	}
}
