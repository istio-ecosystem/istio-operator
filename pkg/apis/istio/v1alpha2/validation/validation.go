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

package validation

import (
	fmt "fmt"
	"reflect"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/util"
)

const (
	validationMethodName = "Validation"
)

// ValidateConfig  calls validation func for every defined element in Values
func ValidateConfig(failOnMissingValidation bool, values *v1alpha2.Values, icpls *v1alpha2.IstioControlPlaneSpec) util.Errors {
	var validationErrors util.Errors

	validationErrors = util.AppendErrs(validationErrors, validateSubTypes(reflect.ValueOf(values).Elem(), failOnMissingValidation, values, icpls))

	return validationErrors
}

func validateSubTypes(e reflect.Value, failOnMissingValidation bool, values *v1alpha2.Values, icpls *v1alpha2.IstioControlPlaneSpec) util.Errors {
	var validationErrors util.Errors

	for i := 0; i < e.NumField(); i++ {
		// Validation is not required if it is not a defined type
		if e.Field(i).Kind() != reflect.Interface && e.Field(i).Kind() != reflect.Ptr {
			continue
		}
		val := e.Field(i).Elem()
		if util.IsNilOrInvalidValue(val) {
			continue
		}
		validation := e.Field(i).MethodByName(validationMethodName)
		if util.IsNilOrInvalidValue(validation) {
			if failOnMissingValidation {
				validationErrors = util.AppendErr(validationErrors, fmt.Errorf("type %s is missing Validation method", e.Type().Field(i).Type))
			}
		} else {
			r := validation.Call([]reflect.Value{reflect.ValueOf(failOnMissingValidation), reflect.ValueOf(values), reflect.ValueOf(icpls)})[0].Interface().(util.Errors)
			if len(r) != 0 {
				validationErrors = util.AppendErrs(validationErrors, r)
			}
		}
		validationErrors = util.AppendErrs(validationErrors, validateSubTypes(e.Field(i).Elem(), failOnMissingValidation, values, icpls))
	}

	return validationErrors
}
