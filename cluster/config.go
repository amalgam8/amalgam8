// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

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
