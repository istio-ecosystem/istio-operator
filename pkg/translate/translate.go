// Copyright 2017 Istio Authors
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

// Package translate defines translations from installer proto to values.yaml.
package translate

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
	"istio.io/operator/pkg/apis/istio/v1alpha1"
	"istio.io/operator/pkg/util"
)

// TranslationFunc maps a yamlStr API path into a YAML values tree.
type TranslationFunc func(t *Translation, root util.Tree, valuesPath string, value interface{}) error

// Translation is a mapping between a yamlStr data path and a YAML values path.
type Translation struct {
	yAMLPath        string
	k8sPath         string
	translationFunc TranslationFunc
}

var (
	// defaultMappings is a mapping between an API path and the corresponding values.yaml path using longest prefix
	// match. If the path is a non-leaf node, the output path is the matching portion of the path, plus any remaining
	// output path.
	defaultMappings = map[string]*Translation{
		"Hub":                                {"global.hub", "", nil},
		"Tag":                                {"global.tag", "", nil},
		"K8SDefaults.Resources.Requests.cpu": {"global.defaultResources.requests.cpu", "", nil},

		"TrafficManagement.ClusterDomain":                                   {"global.clusterDomain", "", nil},
		"TrafficManagement.SidecarInjector.EnableNamespacesByDefault.Value": {"sidecarInjectorWebhook.enableNamespacesByDefault", "", nil},
		"TrafficManagement.Proxy.Common.Resources.Requests.cpu":             {"global.proxy.resources.requests.cpu", "", nil},
		"TrafficManagement.Proxy.Common.Resources.Requests.memory":          {"global.proxy.resources.requests.memory", "", nil},
		"TrafficManagement.Proxy.Common.Resources.Limits.cpu":               {"global.proxy.resources.limits.cpu", "", nil},
		"TrafficManagement.Proxy.Common.Resources.Limits.memory":            {"global.proxy.resources.limits.memory", "", nil},

		"PolicyTelemetry.PolicyCheckFailOpen":       {"global.policyCheckFailOpen", "", nil},
		"PolicyTelemetry.OutboundTrafficPolicyMode": {"global.outboundTrafficPolicy.mode", "", nil},

		"Security.ControlPlaneMtls.Value":    {"global.controlPlaneSecurityEnabled", "", nil},
		"Security.DataPlaneMtlsStrict.Value": {"global.mtls.enabled", "", nil},
		"Security.TrustDomain":               {"global.trustDomain", "", nil},
		"Security.SelfSigned.Value":          {"security.selfSigned", "", nil},
		"Security.CreateMeshPolicy.Value":    {"security.createMeshPolicy", "", nil},
	}
)

// defaultTranslationFunc is the default translation to values. It maps a Go data path into a YAML path.
func defaultTranslationFunc(m *Translation, root util.Tree, valuesPath string, value interface{}) error {
	var path []string

	if util.IsEmptyString(value) {
		dbgPrint("Skip empty string value for path %s", m.k8sPath)
		return nil
	}
	if valuesPath == "" {
		dbgPrint("Not mapping to values, resources path is %s", m.k8sPath)
		return nil
	}

	for _, p := range util.PathFromString(valuesPath) {
		path = append(path, firstCharToLower(p))
	}

	return setYAML(root, path, value)
}

// ProtoToValues traverses the supplied InstallerSpec and returns a values.yaml translation from it. Mappings defines
// a mapping set of translations.
func ProtoToValues(mappings map[string]*Translation, ii *v1alpha1.InstallerSpec) (string, error) {
	root := make(util.Tree)

	errs := protoToValues(mappings, ii, root, nil)

	if len(root) == 0 {
		return "", nil
	}

	y, err := yaml.Marshal(root)
	if err != nil {
		return "", util.AppendErr(errs, err).ToError()
	}

	return string(y), errs.ToError()
}

// protoToValues takes an interface which must be a struct ptr and recursively iterates through all its fields.
// For each leaf, if looks for a mapping from the struct data path to the corresponding YAML path and if one is
// found, it calls the associated mapping function if one is defined to populate the values YAML path.
// If no mapping function is defined, it uses the default mapping function.
func protoToValues(mappings map[string]*Translation, structPtr interface{}, root util.Tree, path util.Path) (errs util.Errors) {
	dbgPrint("protoToValues with path %s, %v (%T)", path, structPtr, structPtr)
	if structPtr == nil {
		return nil
	}
	if reflect.TypeOf(structPtr).Kind() != reflect.Ptr {
		return util.NewErrs(fmt.Errorf("protoToValues path %s, value: %v, expected ptr, got %T", path, structPtr, structPtr))
	}
	structElems := reflect.ValueOf(structPtr).Elem()
	if reflect.TypeOf(structElems).Kind() != reflect.Struct {
		return util.NewErrs(fmt.Errorf("protoToValues path %s, value: %v, expected struct, got %T", path, structElems, structElems))
	}

	if util.IsNilOrInvalidValue(structElems) {
		return
	}

	for i := 0; i < structElems.NumField(); i++ {
		fieldName := structElems.Type().Field(i).Name
		fieldValue := structElems.Field(i)
		kind := structElems.Type().Field(i).Type.Kind()
		if a, ok := structElems.Type().Field(i).Tag.Lookup("json"); ok && a == "-" {
			continue
		}

		dbgPrint("Checking field %s", fieldName)
		switch kind {
		case reflect.Struct:
			dbgPrint("Struct")
			errs = util.AppendErrs(errs, protoToValues(mappings, fieldValue.Addr().Interface(), root, append(path, fieldName)))
		case reflect.Map:
			dbgPrint("Map")
			newPath := append(path, fieldName)
			for _, key := range fieldValue.MapKeys() {
				nnp := append(newPath, key.String())
				errs = util.AppendErrs(errs, insertLeaf(mappings, root, nnp, fieldValue.MapIndex(key)))
			}
		case reflect.Slice:
			dbgPrint("Slice")
			for i := 0; i < fieldValue.Len(); i++ {
				errs = util.AppendErrs(errs, protoToValues(mappings, fieldValue.Index(i).Elem().Interface(), root, path))
			}
		case reflect.Ptr:
			if util.IsNilOrInvalidValue(fieldValue.Elem()) {
				continue
			}
			newPath := append(path, fieldName)
			if fieldValue.Elem().Kind() == reflect.Struct {
				dbgPrint("Struct Ptr")
				errs = util.AppendErrs(errs, protoToValues(mappings, fieldValue.Interface(), root, newPath))
			} else {
				fmt.Println("Leaf Ptr")
				errs = util.AppendErrs(errs, insertLeaf(mappings, root, newPath, fieldValue))
			}
		default:
			dbgPrint("field has kind %s", kind)
			if structElems.Field(i).CanInterface() {
				errs = util.AppendErrs(errs, insertLeaf(mappings, root, append(path, fieldName), fieldValue))
			}
		}
	}
	return errs
}

func insertLeaf(mappings map[string]*Translation, root util.Tree, newPath util.Path, fieldValue reflect.Value) (errs util.Errors) {
	// Must be a scalar leaf. See if we have a mapping.
	valuesPath, m := getValuesPathMapping(mappings, newPath)
	var v interface{}
	if fieldValue.Kind() == reflect.Ptr {
		v = fieldValue.Elem().Interface()
	} else {
		v = fieldValue.Interface()
	}
	switch {
	case m == nil:
		break
	case m.translationFunc == nil:
		// Use default translation which just maps to a different part of the tree.
		errs = util.AppendErr(errs, defaultTranslationFunc(m, root, valuesPath, v))
	default:
		// Use a custom translation function.
		errs = util.AppendErr(errs, m.translationFunc(m, root, valuesPath, v))
	}
	return errs
}

// getValuesPathMapping tries to map path against the passed in mappings with a longest prefix match. If a matching prefix
// is found, it returns the translated YAML path and the corresponding translation.
// e.g. for mapping "a/b"  -> "1/2", the input path "a/b/c/d" would yield "1/2/c/d".
func getValuesPathMapping(mappings map[string]*Translation, path util.Path) (string, *Translation) {
	p := path
	var m *Translation
	for ; len(p) > 0; p = p[0 : len(p)-1] {
		m = mappings[p.String()]
		if m != nil {
			break
		}
	}
	if m == nil {
		return "", nil
	}

	if m.yAMLPath == "" {
		return "", m
	}

	d := len(path) - len(p)
	out := filepath.Join(m.yAMLPath, path[len(path)-d:].String())
	dbgPrint("translating %s to %s", path, out)
	return out, m
}

// setYAML sets the YAML path in the given Tree to the given value, creating any required intermediate nodes.
func setYAML(root util.Tree, path util.Path, value interface{}) error {
	dbgPrint("setYAML %s:%v", path, value)
	if len(path) == 0 {
		return fmt.Errorf("path cannot be empty")
	}
	if len(path) == 1 {
		root[path[0]] = value
		return nil
	}
	if root[path[0]] == nil {
		root[path[0]] = make(util.Tree)
	}
	return setYAML(root[path[0]].(util.Tree), path[1:], value)
}

func dbgPrint(v ...interface{}) {
	return
	fmt.Println(fmt.Sprintf(v[0].(string), v[1:]...))
}

func firstCharToLower(s string) string {
	return strings.ToLower(s[0:1]) + s[1:]
}
