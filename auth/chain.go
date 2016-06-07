package auth

import "fmt"

// NewChainAuthenticator creates and initializes a new chain authenticator, wrapping the given authenticators.
func NewChainAuthenticator(auths []Authenticator) (Authenticator, error) {
	if len(auths) == 0 {
		return nil, fmt.Errorf("Authenticators list is empty")
	}

	ca := &chainAuthenticator{
		authenticators: make([]Authenticator, 0, len(auths)),
	}
	for _, a := range auths {
		if a == nil {
			return nil, fmt.Errorf("Authenticators list contains a nil authenticator")
		}
		ca.authenticators = append(ca.authenticators, a)
	}
	return ca, nil
}

type chainAuthenticator struct {
	authenticators []Authenticator
}

// Authenticate verifies the specified token with the registered authenticators.
// The function returns the Namespace of this token or an error if the token is not valid
func (r *chainAuthenticator) Authenticate(token string) (*Namespace, error) {
	// Scan the list of authenticators in order
	for _, a := range r.authenticators {
		namespace, err := a.Authenticate(token)
		if err == ErrUnrecognizedToken {
			continue
		}
		return namespace, err
	}

	return nil, ErrUnauthorized
}
