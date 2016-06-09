package supervisor

import (
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

// DoAppSupervision TODO
func DoAppSupervision(cmdArgs []string) {

	// Launch the user app
	var appProcess *os.Process
	appChan := make(chan error, 1)
	if len(cmdArgs) > 0 {
		log.Infof("Launching app '%s' with args '%s'", cmdArgs[0], strings.Join(cmdArgs[1:], " "))

		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout

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
				Shutdown(0)
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

			Shutdown(exitCode)
		}
	}
}

// Shutdown TODO
func Shutdown(exitCode int) {
	// TODO: Gracefully shutdown Ngnix, and filebeat
	//ugly temporary hack to kill all processes in container
	exec.Command("pkill", "-9", "-f", "nginx")
	exec.Command("pkill", "-9", "-f", "filebeat")

	log.Infof("Shutting down with exit code %d", exitCode)
	os.Exit(exitCode)
}

// DoLogManagement TODO
func DoLogManagement(filebeatConf string) {
	// starting filebeat
	logcmd := exec.Command("filebeat", "-c", filebeatConf)
	logcmd.Stdin = os.Stdin
	logcmd.Stdout = os.Stdout
	err := logcmd.Run()
	if err != nil {
		log.WithError(err).Error("Failed to launch filebeat")
		Shutdown(1)
	}
}
