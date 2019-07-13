// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha2

import (
	fmt "fmt"
	"reflect"
)

// Validation  calls a validation func for every defined element of Values
// Validation checks
func (v *Values) Validation(failOnMissingValidation bool) []string {
	var validationErrors []string

	e := reflect.ValueOf(v).Elem()
	for i := 0; i < e.NumField(); i++ {
		// Validation is not requred if it is not a defined type
		if e.Field(i).Kind() != reflect.Interface && e.Field(i).Kind() != reflect.Ptr {
			continue
		}
		val := e.Field(i).Elem()
		if val == reflect.ValueOf(nil) {
			fmt.Printf("element: %s is not defined\n", e.Type().Field(i).Name)
			continue
		}
		validation := e.Field(i).MethodByName("Validation")
		if validation == reflect.ValueOf(nil) {
			if failOnMissingValidation {
				validationErrors = append(validationErrors, fmt.Sprintf("type %s is missing Validation method", e.Type().Field(i).Type))
			}
		}
		r := validation.Call([]reflect.Value{reflect.ValueOf(failOnMissingValidation)})[0].Interface().([]string)
		if len(r) != 0 {
			validationErrors = append(validationErrors, r...)
		}
	}

	return validationErrors
}

// Validation checks
func (c *CNIConfig) Validation(failOnMissingValidation bool) []string {
	fmt.Printf("CNIConfig validation was called\n")
	return nil
}
