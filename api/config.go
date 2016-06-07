package api

import (
	"github.com/amalgam8/registry/auth"
	"github.com/amalgam8/registry/store"
)

// Config encapsulates REST server configuration parameters
type Config struct {
	HTTPAddressSpec string
	Registry        store.Registry
	Authenticator   auth.Authenticator
	RequireHTTPS    bool
}
