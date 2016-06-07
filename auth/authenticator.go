package auth

// Module name to be used in logging
const module = "AUTH"

// Authenticator is an interface for token authentication
type Authenticator interface {
	Authenticate(token string) (*Namespace, error)
}

// DefaultAuthenticator returns the default authenticator provided by this package
func DefaultAuthenticator() Authenticator {
	return NewGlobalAuthenticator()
}
