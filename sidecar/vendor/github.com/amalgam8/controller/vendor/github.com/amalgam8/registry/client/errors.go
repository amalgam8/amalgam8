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

package client

import "bytes"

// ErrorCode represents an error condition which might occur when using the client.
type ErrorCode int

// Enumerate valid ErrorCode values.
const (
	ErrorCodeUndefined ErrorCode = iota

	ErrorCodeUnknownInstance

	ErrorCodeConnectionFailure

	ErrorCodeServiceUnavailable
	ErrorCodeInternalServerError

	ErrorCodeUnauthorized
	ErrorCodeInvalidConfiguration
	ErrorCodeInternalClientError
)

func (code ErrorCode) String() string {
	switch code {

	case ErrorCodeUnknownInstance:
		return "ErrorCodeUnknownInstance"
	case ErrorCodeConnectionFailure:
		return "ErrorCodeConnectionFailure"
	case ErrorCodeServiceUnavailable:
		return "ErrorCodeServiceUnavailable"
	case ErrorCodeInternalServerError:
		return "ErrorCodeInternalServerError"
	case ErrorCodeUnauthorized:
		return "ErrorCodeUnauthorized"
	case ErrorCodeInvalidConfiguration:
		return "ErrorCodeInvalidConfiguration"
	case ErrorCodeInternalClientError:
		return "ErrorCodeInternalClientError"

	default:
		return "ErrorCodeUndefined"
	}
}

// Error represents an actual error occurred which using the client.
type Error struct {
	Code      ErrorCode
	Message   string
	Cause     error
	RequestID string
}

func (err Error) Error() string {
	var buf bytes.Buffer
	buf.WriteString(err.Code.String())
	buf.WriteString(": ")
	buf.WriteString(err.Message)

	if err.Cause != nil {
		buf.WriteString(" (")
		buf.WriteString(err.Cause.Error())
		buf.WriteString(")")
	}

	if err.RequestID != "" {
		buf.WriteString(" (")
		buf.WriteString(err.RequestID)
		buf.WriteString(")")
	}

	return buf.String()
}

func newError(code ErrorCode, message string, cause error, requestID string) Error {
	return Error{
		Code:      code,
		Message:   message,
		Cause:     cause,
		RequestID: requestID,
	}
}
