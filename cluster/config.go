package cluster

import "time"

// Default values
const (
	DefaultBackendType   = FilesystemBackend
	DefaultDirectory     = "/tmp/sd-cluster"
	DefaultTTL           = time.Duration(30) * time.Second
	DefaultRenewInterval = time.Duration(7) * time.Second
	DefaultScanInterval  = time.Duration(5) * time.Second
	DefaultSize          = 0
)

// Config encapsulates cluster configuration parameters
type Config struct {
	BackendType   BackendType
	Directory     string
	TTL           time.Duration
	RenewInterval time.Duration
	ScanInterval  time.Duration
	Size          int
}

// defaultize creates an output configuration based on the input configuration,
// where missing configuration parameters (with zero values) are replaced with default values.
func defaultize(conf *Config) *Config {

	in := conf
	if in == nil {
		in = &Config{}
	}

	// Shallow copy
	out := &*in
	if out.BackendType == UnspecifiedBackend {
		out.BackendType = DefaultBackendType
	}
	if out.Directory == "" {
		out.Directory = DefaultDirectory
	}
	if out.TTL == time.Duration(0) {
		out.TTL = DefaultTTL
	}
	if out.RenewInterval == time.Duration(0) {
		out.RenewInterval = DefaultRenewInterval
	}
	if out.RenewInterval > out.TTL {
		out.RenewInterval = out.TTL / 3
	}
	if out.ScanInterval == time.Duration(0) {
		out.ScanInterval = DefaultScanInterval
	}
	if out.Size == 0 {
		out.Size = DefaultSize
	}

	return out

}
