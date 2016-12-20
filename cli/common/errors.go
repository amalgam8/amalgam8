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
	// ErrGremlinURLInvalid .
	ErrGremlinURLInvalid = errors.New("Invalid Gremlin URL")
	// ErrGremlinURLNotFound .
	ErrGremlinURLNotFound = errors.New("Gremlin URL has not been set")
	// ErrUnsoportedFormat .
	ErrUnsoportedFormat = errors.New("Unsupported format")
	// ErrInvalidFormat .
	ErrInvalidFormat = errors.New("Invalid format")
	// ErrFileNotFound .
	ErrFileNotFound = errors.New("File does not exist")
	// ErrInvalidFile .
	ErrInvalidFile = errors.New("Invalid file")
	// ErrNotFound .
	ErrNotFound = errors.New("Not found")
	// ErrUnknowFlag .
	ErrUnknowFlag = errors.New("flag provided but not defined")
	// ErrInvalidFlagArg .
	ErrInvalidFlagArg = errors.New("flag needs an argument")
	// ErrIncorrectAmountRange .
	ErrIncorrectAmountRange = errors.New("--amount must be between 0 and 100")
	// ErrNotRulesFoundForService .
	ErrNotRulesFoundForService = errors.New("No rules found for service")
	// ErrInvalidStateForTrafficStart .
	ErrInvalidStateForTrafficStart = errors.New("Invalid state for start operation")
	// ErrInvalidStateForTrafficStep .
	ErrInvalidStateForTrafficStep = errors.New("Invalid state for step operation")
	// ErrTopologyOrScenariosNotFound .
	ErrTopologyOrScenariosNotFound = errors.New("You must specify -topology and -scenarios")
)
