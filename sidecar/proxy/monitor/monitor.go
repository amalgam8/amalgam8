package monitor

// Monitor a source
type Monitor interface {
	Start() error
	Stop() error
}
