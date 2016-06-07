package auth

// Namespace represents the service registration and discovery scope
type Namespace string

// NamespaceFrom returns the namespace identified by the given string
func NamespaceFrom(s string) Namespace {
	return Namespace(s)
}

// String returns the string representation of the given Namespace
func (ns Namespace) String() string {
	return string(ns)
}
