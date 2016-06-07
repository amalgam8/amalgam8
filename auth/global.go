package auth

type globalAuthenticator struct{}

var globalAuth = &globalAuthenticator{}

// NewGlobalAuthenticator creates a global authenticator instance
func NewGlobalAuthenticator() Authenticator {
	return globalAuth
}

var globalNamespace = Namespace("global")

func (aut *globalAuthenticator) Authenticate(token string) (*Namespace, error) {
	if token == "" {
		return &globalNamespace, nil
	}

	return nil, ErrUnrecognizedToken
}
