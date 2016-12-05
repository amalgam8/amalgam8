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

package registry

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/amalgam8/amalgam8/pkg/adapters/registry/filesystem"
	"github.com/amalgam8/amalgam8/pkg/adapters/registry/kubernetes"
	"github.com/amalgam8/amalgam8/pkg/api"
	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/pkg/version"
	"github.com/amalgam8/amalgam8/registry/cluster"
	"github.com/amalgam8/amalgam8/registry/config"
	"github.com/amalgam8/amalgam8/registry/replication"
	"github.com/amalgam8/amalgam8/registry/server"
	"github.com/amalgam8/amalgam8/registry/store"
	"github.com/amalgam8/amalgam8/registry/store/eureka"
	"github.com/amalgam8/amalgam8/registry/utils/i18n"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
	"github.com/amalgam8/amalgam8/registry/utils/metrics"
	"github.com/amalgam8/amalgam8/registry/utils/network"
)

// Main is the entrypoint for the registry when running as an executable
func Main() {
	app := cli.NewApp()

	app.Name = "registry"
	app.Usage = "Service Registry Server"
	app.Version = version.Build.Version
	app.Flags = config.Flags
	app.Action = func(context *cli.Context) error {
		return Run(config.NewValuesFromContext(context))
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("failure running main: %s", err.Error())
	}
}

// Run the registry with the given configuration
func Run(conf *config.Values) error {

	// Configure logging
	parsedLogLevel, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		return err
	}
	logrus.SetLevel(parsedLogLevel)

	formatter, err := logging.GetLogFormatter(conf.LogFormat)
	if err != nil {
		return err
	}
	logrus.SetFormatter(formatter)

	// Configure locales and translations
	err = i18n.LoadLocales("./locales")
	if err != nil {
		return err
	}

	// Redis store requires an address and password
	if conf.Store == "redis" {
		if conf.StoreAddr == "" {
			return fmt.Errorf("Address required for Redis store")
		}
	}

	var rep replication.Replication

	// Don't need replication if using a store that's not in memory
	if conf.Store == "inmem" {
		if conf.Replication {

			// Wait for private network to become available
			// In some cloud environments, that may take several seconds
			networkAvailable := network.WaitForPrivateNetwork()
			if !networkAvailable {
				return fmt.Errorf("No private network is available within defined timeout")
			}

			// Configure and create the cluster module
			clConfig := &cluster.Config{
				BackendType: cluster.FilesystemBackend,
				Directory:   conf.ClusterDirectory,
				Size:        conf.ClusterSize,
			}
			cl, err := cluster.New(clConfig)
			if err != nil {
				return fmt.Errorf("Failed to create the cluster module: %s", err)
			}

			// Configure and create the replication module
			self := cluster.NewMember(network.GetPrivateIP(), conf.ReplicationPort)
			repConfig := &replication.Config{
				Membership:  cl.Membership(),
				Registrator: cl.Registrator(self),
			}
			rep, err = replication.New(repConfig)
			if err != nil {
				return fmt.Errorf("Failed to create the replication module: %s", err)
			}
		}
	}

	var authenticator auth.Authenticator
	if len(conf.AuthModes) > 0 {
		auths := make([]auth.Authenticator, len(conf.AuthModes))
		for i, mode := range conf.AuthModes {
			switch mode {
			case "trusted":
				auths[i] = auth.NewTrustedAuthenticator()
			case "jwt":
				jwtAuth, err := auth.NewJWTAuthenticator([]byte(conf.JWTSecret))
				if err != nil {
					return fmt.Errorf("Failed to create the authentication module: %s", err)
				}
				auths[i] = jwtAuth
			default:
				return fmt.Errorf("Failed to create the authentication module: unrecognized authentication mode '%s'", err)
			}
		}
		authenticator, err = auth.NewChainAuthenticator(auths)
		if err != nil {
			return err
		}
	} else {
		authenticator = auth.DefaultAuthenticator()
	}

	catalogsExt := []store.CatalogFactory{}
	// See whether kubernetes adapter is enabled
	if conf.K8sURL != "" {
		k8sFactory := store.NewDiscoveryAdapter(func(namespace auth.Namespace) (api.ServiceDiscovery, error) {
			return kubernetes.New(kubernetes.Config{
				URL:       conf.K8sURL,
				Token:     conf.K8sToken,
				Namespace: namespace,
			})
		})
		catalogsExt = append(catalogsExt, k8sFactory)
	}

	// See whether eureka adapter is enabled
	if len(conf.EurekaURLs) > 0 {
		eurekaFactory, err := eureka.New(&eureka.Config{EurekaURLs: conf.EurekaURLs})
		if err != nil {
			return err
		}
		catalogsExt = append(catalogsExt, eurekaFactory)
	}

	// See whether Filesystem adapter is enabled
	if conf.FSCatalog != "" {
		fsFactory := store.NewDiscoveryAdapter(func(namespace auth.Namespace) (api.ServiceDiscovery, error) {
			return filesystem.New(filesystem.Config{
				Dir:       conf.FSCatalog,
				Namespace: namespace,
			})
		})
		catalogsExt = append(catalogsExt, fsFactory)
	}

	cmConfig := &store.Config{
		DefaultTTL:        conf.DefaultTTL,
		MinimumTTL:        conf.MinTTL,
		MaximumTTL:        conf.MaxTTL,
		SyncWaitTime:      conf.SyncTimeout,
		NamespaceCapacity: conf.NamespaceCapacity,
		Replication:       rep,
		Extensions:        catalogsExt,
		Store:             conf.Store,
		StoreAddr:         conf.StoreAddr,
		StorePassword:     conf.StorePassword,
	}
	cm := store.New(cmConfig)

	serverConfig := &server.Config{
		HTTPAddressSpec: fmt.Sprintf(":%d", conf.APIPort),
		CatalogMap:      cm,
		Authenticator:   authenticator,
		RequireHTTPS:    conf.RequireHTTPS,
	}
	server, err := server.New(serverConfig)
	if err != nil {
		return err
	}

	go metrics.DumpPeriodically()

	return server.Start()
}
