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

	"istio.io/operator/pkg/util"
)

// Validation  calls a validation func for every defined element of Values
func Validation(failOnMissingValidation bool, values *Values, icpls *IstioControlPlaneSpec) []string {
	var validationErrors []string

	validationErrors = append(validationErrors, validateSubTypes(reflect.ValueOf(values).Elem(), failOnMissingValidation, values, icpls)...)

	return validationErrors
}

// Validation checks PilotConfig  and all subc types
func (t *PilotConfig) Validation(failOnMissingValidation bool, values *Values, icpls *IstioControlPlaneSpec) []string {
	var validationErrors []string

	return validationErrors
}

// Validation checks CNIConfig  and all subc types
func (t *CNIConfig) Validation(failOnMissingValidation bool, values *Values, icpls *IstioControlPlaneSpec) []string {
	var validationErrors []string

	return validationErrors
}

func validateSubTypes(e reflect.Value, failOnMissingValidation bool, values *Values, icpls *IstioControlPlaneSpec) []string {
	var validationErrors []string

	for i := 0; i < e.NumField(); i++ {
		// Validation is not required if it is not a defined type
		if e.Field(i).Kind() != reflect.Interface && e.Field(i).Kind() != reflect.Ptr {
			continue
		}
		val := e.Field(i).Elem()
		if util.IsNilOrInvalidValue(val) {
			continue
		}
		validation := e.Field(i).MethodByName("Validation")
		if util.IsNilOrInvalidValue(validation) {
			if failOnMissingValidation {
				validationErrors = append(validationErrors, fmt.Sprintf("type %s is missing Validation method", e.Type().Field(i).Type))
			}
		} else {
			r := validation.Call([]reflect.Value{reflect.ValueOf(failOnMissingValidation), reflect.ValueOf(values), reflect.ValueOf(icpls)})[0].Interface().([]string)
			if len(r) != 0 {
				validationErrors = append(validationErrors, r...)
			}
		}
		validationErrors = append(validationErrors, validateSubTypes(e.Field(i).Elem(), failOnMissingValidation, values, icpls)...)
	}

	return validationErrors
}
