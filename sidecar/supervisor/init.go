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
	cmd := exec.Command(os.Args[0], os.Args[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		logrus.Errorf("Sidecar failed to start: %v", err)
		os.Exit(1)
	}

	handleSignals(cmd.Process)

}

func handleSignals(proc *os.Process) {
	// Trap all signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)

	for sig := range sigChan {
		switch sig {
		case syscall.SIGCHLD:
			reapZombies(proc)
		default:
			forwardSignal(proc, sig)
		}
	}
}

// reapZombies cleans up any zombies sidecar may have inherited from terminated children
// - on SIGCHLD send wait4() (ref http://linux.die.net/man/2/waitpid)
func reapZombies(proc *os.Process) {
	for {
		var waitStatus syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &waitStatus, syscall.WNOHANG, nil)
		if proc != nil && proc.Pid == pid {
			logrus.Debugf("Sidecar exited with exit code %v", waitStatus.ExitStatus())
			os.Exit(waitStatus.ExitStatus())
		}

		// Spurious wakeup
		if err == syscall.EINTR {
			continue
		}
		logrus.Debug("Zombie reaped")
		// Done
		break
	}
}

func forwardSignal(proc *os.Process, sig os.Signal) {
	if proc != nil {
		proc.Signal(sig)
	}
}
