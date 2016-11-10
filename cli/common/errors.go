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

package common

import (
	"errors"
)

var (
	// ErrRegistryURLInvalid .
	ErrRegistryURLInvalid = errors.New("Invalid Registry URL")
	// ErrRegistryURLNotFound .
	ErrRegistryURLNotFound = errors.New("Registry URL has not been set")
	// ErrControllerURLInvalid .
	ErrControllerURLInvalid = errors.New("Invalid Controller URL")
	// ErrControllerURLNotFound .
	ErrControllerURLNotFound = errors.New("Controller URL has not been set")
	// ErrUnsoportedFormat .
	ErrUnsoportedFormat = errors.New("Unsupported format")
	// ErrNotFound .
	ErrNotFound = errors.New("Not found")
	// ErrUnknowFlag .
	ErrUnknowFlag = errors.New("flag provided but not defined")
	// ErrInvalidFlagArg .
	ErrInvalidFlagArg = errors.New("flag needs an argument")
)
