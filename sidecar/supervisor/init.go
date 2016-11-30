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
	"syscall"

	"github.com/Sirupsen/logrus"
)

// Init launches new sidecar process with same args and env vars as current process
// and runs zombie process cleanup for any inherited orphans
func Init() {
	cmd := exec.Command("a8sidecar", os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	// listen for SIGCHLDs from any of the supervised processes and any orphans spun off from those
	go reapZombies()

	go launchChild(cmd)

	// Intercept SIGTERM/SIGINT and stop
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-sigChan:
		logrus.Infof("Intercepted signal '%v'", sig)

		// forwarding signal to supervised applications to exit gracefully
		cmd.Process.Signal(sig)
		shutdown(0)
	}
}

func launchChild(cmd *exec.Cmd) {
	err := cmd.Run()
	exitCode := 0
	if err == nil {
		logrus.Info("App terminated with exit code 0")
	} else {
		exitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			if waitStatus, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = waitStatus.ExitStatus()
			}
			logrus.Errorf("App terminated with exit code %v", exitCode)
		} else {
			logrus.Errorf("App failed to start: %v", err)
		}
	}
	shutdown(exitCode)
}

// reapZombies cleans up any zombies sidecar may have inherited from terminated children
// - on SIGCHLD send wait4() (ref http://linux.die.net/man/2/waitpid)
func reapZombies() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGCHLD)

	for range sigChan {
		for {
			_, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)

			// Spurious wakeup
			if err == syscall.EINTR {
				continue
			}
			logrus.Debug("Zombie reaped")
			// Done
			break
		}
	}
}

func shutdown(sig int) {
	logrus.Infof("Shutting down with exit code %v", sig)
	os.Exit(sig)
}
