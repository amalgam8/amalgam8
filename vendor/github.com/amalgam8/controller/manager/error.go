package manager

import (
	"fmt"
)

// InvalidRuleError error
type InvalidRuleError struct {
	Reason       string
	ErrorMessage string
}

// Error interface
func (e *InvalidRuleError) Error() string {
	return e.Reason
}

// ServiceUnavailableError error
type ServiceUnavailableError struct {
	Reason       string
	ErrorMessage string
	Err          error
}

// Error interface
func (e *ServiceUnavailableError) Error() string {
	return fmt.Sprintf("%v: %v", e.Reason, e.Err.Error())
}

// DBError err
type DBError struct {
	Err error
}

// Error interface
func (e *DBError) Error() string {
	return e.Err.Error()
}

// RuleNotFoundError error
type RuleNotFoundError struct {
	Reason       string
	ErrorMessage string
}

// Error interface
func (e *RuleNotFoundError) Error() string {
	return e.Reason
}
