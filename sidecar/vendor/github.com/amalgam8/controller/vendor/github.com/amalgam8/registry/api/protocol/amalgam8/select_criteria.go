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

package amalgam8

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"

	"github.com/amalgam8/amalgam8/registry/store"
	"github.com/amalgam8/amalgam8/registry/utils/reflection"
)

type selectCriteria struct {
	criteria     map[string]interface{}
	filterStatus string
}

// Parses the request's query params and returns the wanted fields and their expected values
// This method is used to apply a selection filter on instances
func newSelectCriteria(r *rest.Request) (*selectCriteria, error) {
	if len(r.URL.Query()) == 0 {
		return &selectCriteria{}, nil
	}

	c := make(map[string]interface{})

	var filter string

	for param := range r.URL.Query() {
		// fields is used for projection-type filtering
		if param == "fields" {
			continue
		}

		// check whether the param name is a valid field
		if _, ok := instanceQueryValuesToFieldNames[param]; !ok {
			return nil, fmt.Errorf("Field %s is not a valid field", param)
		}

		// put the expected value of param to the selection criteria
		requestedValue := r.URL.Query().Get(param)
		// convert param field's name to its actual value in the struct definition
		fieldName := instanceQueryValuesToFieldNames[param]

		// The status is made upper case on registration so ignore the query param case
		if fieldName == "Status" {
			// If the status is user defined, leave the case alone
			if strings.EqualFold(requestedValue, store.Up) ||
				strings.EqualFold(requestedValue, store.Starting) ||
				strings.EqualFold(requestedValue, store.OutOfService) ||
				strings.EqualFold(requestedValue, store.All) {
				requestedValue = strings.ToUpper(requestedValue)
			}
			filter = requestedValue
		}

		// if it's a string array, split by commas
		tmpServiceInstance := ServiceInstance{}
		fieldType := reflect.Indirect(reflect.ValueOf(tmpServiceInstance)).FieldByName(fieldName).Type().String()

		// ensure field type is string or []string
		if fieldType != "[]string" && fieldType != "string" {
			return nil, fmt.Errorf("Field %s is not a string or a []string but is of type %s", fieldName, fieldType)
		}

		if fieldType == "[]string" {
			c[fieldName] = strings.Split(requestedValue, ",")
		} else {
			c[fieldName] = requestedValue
		}
	}

	return &selectCriteria{criteria: c, filterStatus: filter}, nil
}

func (sc *selectCriteria) instanceFilter(si *store.ServiceInstance) bool {
	var otherkey bool

	if sc.criteria != nil {
		// iterate over selection criteria
		for key, val := range sc.criteria {
			// If the query string param is Status need to support ALL or filter appropriately
			if key == "Status" {
				if sc.filterStatus == store.All {
					continue
				}
			} else {
				// If we're filtering on another key, include any statuses
				otherkey = true
			}
			fits, err := reflection.StructFieldMatchesValue(si, key, val)
			if err != nil || !fits {
				return false
			}
		}
	}

	// Filter out those instances that are not UP if the status query string param was not set
	// unless there is another param specified
	if sc.filterStatus == "" && !otherkey {
		// For now, handle the case where there are user defined statuses as we want to treat them as UP
		switch si.Status {
		case store.Starting:
			return false
		case store.OutOfService:
			return false
		case store.Up:
		default:
		}
	}

	return true
}
