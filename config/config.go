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

package config

import (
	"time"

	"github.com/codegangsta/cli"
)

// Values holds the actual configuration values used for the registry server
type Values struct {
	LogLevel  string
	LogFormat string

	AuthModes    []string
	JWTSecret    string
	RequireHTTPS bool

	APIPort         uint16
	ReplicationPort uint16

	Replication bool
	SyncTimeout time.Duration

	ClusterDirectory string
	ClusterSize      int

	NamespaceCapacity int
	DefaultTTL        time.Duration
	MaxTTL            time.Duration
	MinTTL            time.Duration

	K8sURL   string
	K8sToken string

	FSCatalog string

	Store         string
	StoreAddr     string
	StorePassword string
}

// NewValuesFromContext creates a Config instance from the given CLI context
func NewValuesFromContext(context *cli.Context) *Values {
	return &Values{
		LogLevel:  context.String(LogLevelFlag),
		LogFormat: context.String(LogFormatFlag),

		AuthModes:    context.StringSlice(AuthModeFlag),
		JWTSecret:    context.String(JWTSecretFlag),
		RequireHTTPS: context.Bool(RequireHTTPSFlag),

		APIPort:         uint16(context.Int(RestAPIPortFlag)),
		ReplicationPort: uint16(context.Int(ReplicationPortFlag)),

		Replication: context.Bool(ReplicationFlag),
		SyncTimeout: context.Duration(SyncTimeoutFlag),

		ClusterDirectory: context.String(ClusterDirectoryFlag),
		ClusterSize:      context.Int(ClusterSizeFlag),

		NamespaceCapacity: context.Int(NamespaceCapacityFlag),
		DefaultTTL:        context.Duration(DefaultTTLFlag),
		MaxTTL:            context.Duration(MaxTTLFlag),
		MinTTL:            context.Duration(MinTTLFlag),

		K8sURL:   context.String(K8sURLFlag),
		K8sToken: context.String(K8sTokenFlag),

		FSCatalog: context.String(FSCatalogFlag),

		Store:         context.String(StoreFlag),
		StoreAddr:     context.String(StoreAddrFlag),
		StorePassword: context.String(StorePasswordFlag),
	}
}
