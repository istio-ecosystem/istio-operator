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
	"fmt"
	"reflect"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/util"
)

const (
	validationMethodName = "Validate"
)

// ValidateConfig  calls validation func for every defined element in Values
func ValidateConfig(failOnMissingValidation bool, values *v1alpha2.Values, icpls *v1alpha2.IstioControlPlaneSpec) util.Errors {
	var validationErrors util.Errors

	validationErrors = util.AppendErrs(validationErrors, validateSubTypes(reflect.ValueOf(values).Elem(), failOnMissingValidation, values, icpls))

	return validationErrors
}

func validateSubTypes(e reflect.Value, failOnMissingValidation bool, values *v1alpha2.Values, icpls *v1alpha2.IstioControlPlaneSpec) util.Errors {
	var validationErrors util.Errors
	var ptr reflect.Value
	var value reflect.Value
	var object reflect.Value
	var finalMethod reflect.Value

	// Dealing with receiver pointer and receiver value
	if e.Type().Kind() == reflect.Ptr {
		ptr = e
		value = ptr.Elem()
		object = reflect.Indirect(e)
	} else {
		ptr = reflect.New(reflect.TypeOf(e.Interface()))
		temp := ptr.Elem()
		temp.Set(e)
		object = e
		value = e
	}

	// check for method on value
	method := value.MethodByName(validationMethodName)
	if method.IsValid() {
		finalMethod = method
	}
	// check for method on pointer
	method = ptr.MethodByName(validationMethodName)
	if method.IsValid() {
		finalMethod = method
	}

	if util.IsNilOrInvalidValue(finalMethod) {
		if failOnMissingValidation {
			validationErrors = append(validationErrors, fmt.Errorf("type %s is missing Validation method", e.Type().String()))
		}
	} else {
		r := finalMethod.Call([]reflect.Value{reflect.ValueOf(failOnMissingValidation), reflect.ValueOf(values), reflect.ValueOf(icpls)})[0].Interface().(util.Errors)
		if len(r) != 0 {
			validationErrors = append(validationErrors, r...)
		}
	}
	// If it is not a struct nothing to do, returning previously collected validation errors
	if object.Kind() != reflect.Struct {
		return validationErrors
	}
	for i := 0; i < object.NumField(); i++ {
		// Corner case of a slice or map of something, if something is a defined type, then process it recursiveley.
		if object.Field(i).Kind() == reflect.Slice || object.Field(i).Kind() == reflect.Map {
			validationErrors = append(validationErrors, processMapOrSlice(object.Field(i), failOnMissingValidation, values, icpls)...)
			continue
		}
		// Validation is not required if it is not a defined type
		if object.Field(i).Kind() != reflect.Interface && object.Field(i).Kind() != reflect.Ptr {
			continue
		}
		val := object.Field(i).Elem()
		if util.IsNilOrInvalidValue(val) {
			continue
		}
		validationErrors = append(validationErrors, validateSubTypes(object.Field(i), failOnMissingValidation, values, icpls)...)
	}

	return validationErrors
}

func processMapOrSlice(e reflect.Value, failOnMissingValidation bool, values *v1alpha2.Values, icpls *v1alpha2.IstioControlPlaneSpec) util.Errors {
	var validationErrors util.Errors
	for i := 0; i < e.Len(); i++ {
		if e.Index(i).Kind() == reflect.Interface || e.Index(i).Kind() == reflect.Ptr {
			validationErrors = append(validationErrors, validateSubTypes(e.Index(i), failOnMissingValidation, values, icpls)...)
		}
	}

	return validationErrors
}
