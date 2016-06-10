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

// Error for database
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
