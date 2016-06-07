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
	}
}
