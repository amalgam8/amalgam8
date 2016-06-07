package database

import (
	"fmt"
)

// DBError provides a struct of status/error codes
// this can be  used for the different impls,
// in memory or cloudant
type DBError struct {
	Status     string `json:"-"`
	StatusCode int    `json:"-"`
	ErrorType  string `json:"error"`
	Reason     string `json:"reason"`
}

// Error TODO
func (e *DBError) Error() string {
	return fmt.Sprintf(
		"DatabaseError: status_code=%v error=%v reason=%v",
		e.StatusCode, e.ErrorType, e.Reason,
	)
}

// NewDatabaseError generates a new Cloudant error
// we need this so we can check for the same error in the
// code no matter if in memory or db
func NewDatabaseError(reason, status, error string, respcode int) error {
	dbErr := new(DBError)
	dbErr.Reason = reason
	dbErr.Status = status
	dbErr.ErrorType = error
	dbErr.StatusCode = respcode

	return dbErr
}
