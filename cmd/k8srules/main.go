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

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)
}

// Package global logger
var logger = logrus.WithField("module", "ROUTINGRULES")

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	_, err := new(ctx, os.Getenv("K8S_NAMESPACE"))
	if err != nil {
		logger.WithError(err).Error("Failed creating Routing-Rules Controller")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cancel()
}
