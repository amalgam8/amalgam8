package client

import "bytes"

type ErrorCode int

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
