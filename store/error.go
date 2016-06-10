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

package store

import (
	"fmt"
)

// ErrorCode represents an error condition that might occur during a registry operation
type ErrorCode int

// ErrorCode predefined values
const (
	ErrorBadRequest ErrorCode = iota
	ErrorNoSuchServiceName
	ErrorNoSuchServiceInstance
	ErrorNamespaceQuotaExceeded
	ErrorInternalServerError
)

// Error is an error implementation that is associated with an ErrorCode
type Error struct {
	Code    ErrorCode
	Message string
	Cause   interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d - %s (%v)", e.Code, e.Message, e.Cause)
}

// NewError creates a new registry.Error with the specified code, message and cause.
func NewError(code ErrorCode, message string, cause interface{}) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}
