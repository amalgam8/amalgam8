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

// Entry generic Cloudant entry
// To use this interface with the Cloudant DB, the following fields need to
// be present in the struct:
//    ID string `json:"_id"`
//    Rev string `json:"_rev,omitempty"`
type Entry interface {
	IDRev() (string, string)
	GetIV() string
	SetIV(iv string)
	SetRev()
}

// AllDocs generic interface representing bulk read objects
type AllDocs interface {
	GetEntries() []Entry
}
