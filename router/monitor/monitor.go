package monitor

type Monitor interface {
	Start() error
	Stop() error
}
