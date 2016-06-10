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

package reflection

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// ExistsInArray Returns true if value exists in arr, false otherwise
func ExistsInArray(value string, arr []string) bool {
	for _, val := range arr {
		if val == value {
			return true
		}
	}
	return false
}

func existsInArrayGeneric(value interface{}, array interface{}) bool {
	for j := 0; j < reflect.ValueOf(array).Len(); j++ {
		if value == reflect.ValueOf(array).Index(j).Interface() {
			return true
		}
	}
	return false
}

// Cache of (struct name) --> (struct fields) for optimization
var fieldsCache = make(map[string][]string)

func getFieldsOfStruct(obj interface{}) []string {
	structType := reflect.TypeOf(obj).String()
	var fields, exists = fieldsCache[structType]
	if exists {
		return fields
	}

	structMetadata := reflect.Indirect(reflect.ValueOf(obj))
	numFields := structMetadata.NumField()
	fieldsInStruct := make([]string, numFields, numFields)
	for i := 0; i < numFields; i++ {
		fieldsInStruct[i] = structMetadata.Type().Field(i).Name
	}
	fieldsCache[structType] = fieldsInStruct
	return fieldsInStruct
}

// FilterStructByFields Returns a copy of the same struct with only the includedFields,
// or the a copy if the fields passed are nil
func FilterStructByFields(in interface{}, out interface{}, includedFields []string) (err error) {
	defer func() {
		if cause := recover(); cause != nil {
			err = fmt.Errorf("Filtering failed in reflection: %v", cause)
		}
	}()

	if b, err := json.Marshal(in); err == nil {
		err := json.Unmarshal(b, out)
		if includedFields == nil || err != nil {
			return err
		}
	} else {
		return err
	}

	outStructMetadata := reflect.Indirect(reflect.ValueOf(out))
	fieldsOfStruct := getFieldsOfStruct(in)

	for _, field := range includedFields {
		if !ExistsInArray(field, fieldsOfStruct) {
			return fmt.Errorf("Field '%s' doesn't exist in struct", field)
		}
	}

	for _, field := range fieldsOfStruct {
		if !ExistsInArray(field, includedFields) {
			fldObj := outStructMetadata.FieldByName(field)
			fldObj.Set(reflect.Zero(fldObj.Type()))
		}
	}
	return nil
}

// matchValues matches value1 against value2
// Assumes both values are same type
// If they're arrays, returns whether value1 is subset of value2
// else, compares between the two values
func matchValues(value1 interface{}, value2 interface{}) bool {
	valueType := reflect.TypeOf(value1).String()
	// if type of value1 is less than 2 chars, [0:2] access will panic
	if len(valueType) <= 2 || valueType[0:2] != "[]" {
		return value1 == value2
	}

	for i := 0; i < reflect.ValueOf(value1).Len(); i++ {
		if !existsInArrayGeneric(reflect.ValueOf(value1).Index(i).Interface(), value2) {
			return false
		}
	}
	return true
}

// StructFieldMatchesValue Returns whether a struct entity's given field matches a certain value
// If it's an array, is returns whether the array's members are a subset of the field's array members
// else, they are compared by a == operator
// Returns true if value is nil
// Returns an error in case of errors in reflection or if values are of different types
func StructFieldMatchesValue(structEntity interface{}, field string, value interface{}) (matches bool, err error) {
	if value == nil {
		return true, nil
	}

	defer func() {
		if cause := recover(); cause != nil {
			err = fmt.Errorf("Filtering failed in reflection: %v", cause)
		}
	}()

	fieldsOfStruct := getFieldsOfStruct(structEntity)
	if !ExistsInArray(field, fieldsOfStruct) {
		return false, fmt.Errorf("Field '%s' doesn't exist in struct", field)
	}

	valueType := reflect.TypeOf(value).String()
	fieldType := reflect.Indirect(reflect.ValueOf(structEntity)).FieldByName(field).Type().String()
	if valueType != fieldType {
		return false, fmt.Errorf("Field '%s' is of type '%s', but given value is of type '%s'", field, fieldType, valueType)
	}
	return matchValues(value, reflect.Indirect(reflect.ValueOf(structEntity)).FieldByName(field).Interface()), nil
}

// GetJSONToFieldsMap returns a map from JSON field to struct field name
func GetJSONToFieldsMap(structEntity interface{}) map[string]string {
	t := reflect.TypeOf(structEntity)

	// If structEntity is a pointer then we need to de-reference it
	if t.Kind() == reflect.Ptr {
		t = reflect.Indirect(reflect.ValueOf(structEntity)).Type()
	}

	numFields := t.NumField()
	fieldsMap := map[string]string{}

	for i := 0; i < numFields; i++ {
		// Find the json field name
		jsonName := t.Field(i).Name
		tag := t.Field(i).Tag.Get("json")
		if tag == "-" {
			continue
		}

		if idx := strings.Index(tag, ","); idx != -1 {
			jsonName = tag[:idx]
		} else if tag != "" {
			jsonName = tag
		}

		fieldsMap[jsonName] = t.Field(i).Name
	}

	return fieldsMap
}
