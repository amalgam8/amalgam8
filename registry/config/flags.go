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

	"strings"

	"github.com/amalgam8/amalgam8/registry/cluster"
	"github.com/codegangsta/cli"
)

// Flag names
const (
	LogLevelFlag  = "log_level"
	LogFormatFlag = "log_format"

	AuthModeFlag     = "auth_mode"
	JWTSecretFlag    = "jwt_secret"
	RequireHTTPSFlag = "require_https"

	RestAPIPortFlag     = "api_port"
	ReplicationPortFlag = "replication_port"

	ReplicationFlag = "replication"
	SyncTimeoutFlag = "sync_timeout"

	ClusterDirectoryFlag = "cluster_dir"
	ClusterSizeFlag      = "cluster_size"

	NamespaceCapacityFlag = "namespace_capacity"
	DefaultTTLFlag        = "default_ttl"
	MaxTTLFlag            = "max_ttl"
	MinTTLFlag            = "min_ttl"

	K8sURLFlag   = "k8s_url"
	K8sTokenFlag = "k8s_token"

	FSCatalogFlag = "fs_catalog"

	StoreFlag         = "store"
	StoreAddrFlag     = "store_address"
	StorePasswordFlag = "store_password"
)

// Flags represents the set of supported flags
var Flags = []cli.Flag{

	cli.StringFlag{
		Name:   LogLevelFlag,
		EnvVar: envVarFromFlag(LogLevelFlag),
		Value:  "info",
		Usage:  "Logging level. Supported values are: 'debug', 'info', 'warn', 'error', 'fatal', 'panic'",
	},

	cli.StringFlag{
		Name:   LogFormatFlag,
		EnvVar: envVarFromFlag(LogFormatFlag),
		Value:  "text",
		Usage:  "Logging format. Supported values are: 'text', 'json', 'logstash'",
	},

	cli.StringSliceFlag{
		Name:   AuthModeFlag,
		EnvVar: envVarFromFlag(AuthModeFlag),
		Usage:  "Authentication modes. Supported values are: 'trusted', 'jwt'",
	},

	cli.StringFlag{
		Name:   JWTSecretFlag,
		EnvVar: envVarFromFlag(JWTSecretFlag),
		Usage:  "Secret key for JWT authentication",
	},

	cli.BoolFlag{
		Name:   RequireHTTPSFlag,
		EnvVar: envVarFromFlag(RequireHTTPSFlag),
		Usage:  "Require clients to use HTTPS for API calls",
	},

	cli.IntFlag{
		Name:   RestAPIPortFlag,
		EnvVar: envVarFromFlag(RestAPIPortFlag),
		Value:  8080,
		Usage:  "REST API port number",
	},

	cli.IntFlag{
		Name:   ReplicationPortFlag,
		EnvVar: envVarFromFlag(ReplicationPortFlag),
		Value:  6100,
		Usage:  "Replication port number",
	},

	cli.BoolFlag{
		Name:   ReplicationFlag,
		EnvVar: envVarFromFlag(ReplicationFlag),
		Usage:  "Enable replication",
	},

	cli.DurationFlag{
		Name:   SyncTimeoutFlag,
		EnvVar: envVarFromFlag(SyncTimeoutFlag),
		Value:  30 * time.Second,
		Usage:  "Registry timeout for establishing peer synchronization connection",
	},

	cli.StringFlag{
		Name:   ClusterDirectoryFlag,
		EnvVar: envVarFromFlag(ClusterDirectoryFlag),
		Value:  cluster.DefaultDirectory,
		Usage:  "Filesystem directory for cluster membership",
	},

	cli.IntFlag{
		Name:   ClusterSizeFlag,
		EnvVar: envVarFromFlag(ClusterSizeFlag),
		Value:  0,
		Usage:  "Cluster minimal healthy size",
	},

	cli.DurationFlag{
		Name:   DefaultTTLFlag,
		EnvVar: envVarFromFlag(DefaultTTLFlag),
		Value:  30 * time.Second,
		Usage:  "Registry default TTL",
	},

	cli.DurationFlag{
		Name:   MaxTTLFlag,
		EnvVar: envVarFromFlag(MaxTTLFlag),
		Value:  10 * time.Minute,
		Usage:  "Registry maximum TTL",
	},

	cli.DurationFlag{
		Name:   MinTTLFlag,
		EnvVar: envVarFromFlag(MinTTLFlag),
		Value:  10 * time.Second,
		Usage:  "Registry minimum TTL",
	},

	cli.IntFlag{
		Name:   NamespaceCapacityFlag,
		EnvVar: envVarFromFlag(NamespaceCapacityFlag),
		Value:  -1,
		Usage:  "Registry namespace capacity, value of -1 indicates no capacity limit",
	},

	cli.StringFlag{
		Name:   K8sURLFlag,
		EnvVar: envVarFromFlag(K8sURLFlag),
		Usage:  "Enable kubernetes catalog and specify the kubernetes API server URL",
	},

	cli.StringFlag{
		Name:   K8sTokenFlag,
		EnvVar: envVarFromFlag(K8sTokenFlag),
		Usage:  "Kubernetes token API",
	},

	cli.StringFlag{
		Name:   FSCatalogFlag,
		EnvVar: envVarFromFlag(FSCatalogFlag),
		Usage:  "Enable FileSystem catalog and specify the name of the data file",
	},

	cli.StringFlag{
		Name:   StoreFlag,
		EnvVar: envVarFromFlag(StoreFlag),
		Value:  "inmem",
		Usage:  "Backing store. Supported values are: 'inmem', 'redis'",
	},

	cli.StringFlag{
		Name:   StoreAddrFlag,
		EnvVar: envVarFromFlag(StoreAddrFlag),
		Value:  "",
		Usage:  "Store address",
	},
	cli.StringFlag{
		Name:   StorePasswordFlag,
		EnvVar: envVarFromFlag(StorePasswordFlag),
		Value:  "",
		Usage:  "Store password",
	},
}

// envVarFromFlag returns the environment variable bound to the given flag
func envVarFromFlag(name string) string {
	return "A8_" + strings.ToUpper(name)
}
