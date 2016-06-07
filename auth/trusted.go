package auth

type trustedAuthenticator struct{}

var trustedAuth = &trustedAuthenticator{}

// NewTrustedAuthenticator creates a trusted authenticator instance
func NewTrustedAuthenticator() Authenticator {
	return trustedAuth
}

func (aut *trustedAuthenticator) Authenticate(token string) (*Namespace, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	namespace := Namespace(token)
	return &namespace, nil
}
