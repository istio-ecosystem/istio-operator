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

// Package translate defines translations from installer proto to values.yaml.
package translate

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/helm"
	"istio.io/operator/pkg/manifest"
	"istio.io/operator/pkg/name"
	"istio.io/operator/pkg/util"
	"istio.io/operator/pkg/version"
	"gopkg.in/yaml.v2"

	istio.io/pkg/log"
)

// Translator is a set of mappings to translate between API paths, charts, values.yaml and k8s paths.
type Translator struct {
	Version version.MinorVersion
	// APIMapping is a mapping between an API path and the corresponding values.yaml path using longest prefix
	// match. If the path is a non-leaf node, the output path is the matching portion of the path, plus any remaining
	// output path.
	APIMapping map[string]*Translation
	// KubernetesMapping defines mappings from an IstioControlPlane API paths to k8s resource paths.
	KubernetesMapping map[string]*Translation
	// ComponentToHelmValuesName is the root component name used in values YAML files in component charts.
	ComponentToHelmValuesName map[name.ComponentName]string
	// FeatureComponentToValues maps from feature and component names to value paths in values.yaml schema.
	FeatureComponentToValues map[name.FeatureName]map[name.ComponentName]componentValuePaths
	// ComponentDirLayout is a mapping between a component name and the subdirectory of the component Chart.
	ComponentDirLayout map[name.ComponentName]string
}

// TranslationFunc maps a yamlStr API path into a YAML values tree.
type TranslationFunc func(t *Translation, root util.Tree, valuesPath string, value interface{}) error

// Translation is a mapping between paths.
type Translation struct {
	inPath          string
	outPath         string
	translationFunc TranslationFunc
}

// componentValuePaths is a group of paths that exists in both IstioControlPlane and values.yaml.
type componentValuePaths struct {
	enabled   string
	namespace string
}

var (
	Mappings = map[version.MinorVersion]*Translator{
		version.MinorVersion{1, 2}: {
			APIMapping: map[string]*Translation{
				"Hub": {"global.hub", "", nil},
				"Tag": {"global.tag", "", nil},

				"TrafficManagement.Components.Proxy.Common.Values": {"global.proxy", "", nil},

				"PolicyTelemetry.PolicyCheckFailOpen":       {"global.policyCheckFailOpen", "", nil},
				"PolicyTelemetry.OutboundTrafficPolicyMode": {"global.outboundTrafficPolicy.mode", "", nil},

				"Security.ControlPlaneMtls.Value":    {"global.controlPlaneSecurityEnabled", "", nil},
				"Security.DataPlaneMtlsStrict.Value": {"global.mtls.enabled", "", nil},
			},
			KubernetesMapping: map[string]*Translation{
				"TrafficManagement.Components.Pilot.Common.K8S.Resources": {"[Deployment:istio-pilot].spec.template.spec.resources", "", nil},
			},
			ComponentToHelmValuesName: map[name.ComponentName]string{
				name.IstioBaseComponentName:       "global",
				name.PilotComponentName:           "pilot",
				name.GalleyComponentName:          "galley",
				name.SidecarInjectorComponentName: "sidecarInjectorWebhook",
				name.PolicyComponentName:          "mixer.policy",
				name.TelemetryComponentName:       "mixer.telemetry",
				name.CitadelComponentName:         "citadel",
				name.NodeAgentComponentName:       "nodeAgent",
				name.CertManagerComponentName:     "certManager",
				name.IngressComponentName:         "gateways.istio-ingressgateway",
				name.EgressComponentName:          "gateways.istio-ingressgateway",
			},
			FeatureComponentToValues: map[name.FeatureName]map[name.ComponentName]componentValuePaths{
				name.TrafficManagementFeatureName: {
					name.PilotComponentName: {
						enabled:   "pilot.enabled",
						namespace: "global.istioNamespace",
					},
				},
				name.PolicyFeatureName: {
					name.PolicyComponentName: {
						enabled:   "mixer.policy.enabled",
						namespace: "global.policyNamespace",
					},
				},
				name.TelemetryFeatureName: {
					name.TelemetryComponentName: {
						enabled:   "mixer.telemetry.enabled",
						namespace: "global.telemetryNamespace",
					},
				},
				// TODO: check if these really should be settable.
				name.SecurityFeatureName: {
					name.NodeAgentComponentName: {
						enabled:   "nodeagent.enabled",
						namespace: "global.istioNamespace",
					},
					name.CertManagerComponentName: {
						enabled:   "certmanager.enabled",
						namespace: "global.istioNamespace",
					},
					name.CitadelComponentName: {
						enabled:   "security.enabled",
						namespace: "global.istioNamespace",
					},
				},
			},
			ComponentDirLayout: map[name.ComponentName]string{
				name.PilotComponentName:           "istio-control/istio-discovery",
				name.GalleyComponentName:          "istio-control/istio-config",
				name.SidecarInjectorComponentName: "istio-control/istio-autoinject",
				name.PolicyComponentName:          "istio-policy",
				name.TelemetryComponentName:       "istio-telemetry",
				name.CitadelComponentName:         "security/citadel",
				name.NodeAgentComponentName:       "security/nodeagent",
				name.CertManagerComponentName:     "security/certmanager",
				name.IngressComponentName:         "gateways/istio-ingress",
				name.EgressComponentName:          "gateways/istio-egress",
			},
		},
	}
)

// OverlayK8sSettings overlays k8s settings from icp over the manifest objects, based on t's translation mappings.
func (t *Translator) OverlayK8sSettings(objects *manifest.Objects, icp *v1alpha2.IstioControlPlaneSpec) error {
	om := objects.ToNameKindMap()
	for _, v := range t.KubernetesMapping {
		m := make(map[string]interface{})
		found, err := name.SetFromPath(icp, v.inPath, &m)
		if err != nil {
			return err
		}
		if !found {
			log.Printf("inPath %s not found in IstioControlPlaneSpec, skip map to output", v.inPath)
			continue
		}
		path := util.PathFromString(v.outPath)
		pe := path[0]
		if !util.IsKVPathElement(pe) {
			return fmt.Errorf("path %s has an unexpected first element %s in OverlayK8sSettings", path, pe)
		}
		oo, ok := om[pe]
		if !ok {
			return fmt.Errorf("resource Kind:Name %s doesn't exist in output manifest", pe)
		}
		baseYAML, err := oo.YAML()
		if err != nil {
			return err
		}
		overlayYAML, err := yaml.Marshal(m)
		if err != nil {
			return err
		}
		mergedYAML, err := helm.OverlayYAML(string(baseYAML), string(overlayYAML))
		if err != nil {
			return err
		}
		mergedObj, err := manifest.ParseJSONToObject([]byte(mergedYAML))
		if err != nil {
			return err
		}
		om[pe] = mergedObj
	}
	return nil
}

// ProtoToValues traverses the supplied InstallerSpec and returns a values.yaml translation from it.
func (t *Translator) ProtoToValues(ii *v1alpha2.IstioControlPlaneSpec) (string, error) {
	root := make(util.Tree)

	errs := t.protoToValues(ii, root, nil)
	if len(errs) != 0 {
		return "", errs.ToError()
	}

	if err := t.setEnablementAndNamespaces(root, ii); err != nil {
		return "", err
	}

	if len(root) == 0 {
		return "", nil
	}

	y, err := yaml.Marshal(root)
	if err != nil {
		return "", util.AppendErr(errs, err).ToError()
	}

	return string(y), errs.ToError()
}

// ValuesOverlaysToHelmValues translates from component value overlays to helm value overlay paths.
func (t *Translator) ValuesOverlaysToHelmValues(in map[string]interface{}, cname name.ComponentName) map[string]interface{} {
	out := make(map[string]interface{})
	toPath, ok := t.ComponentToHelmValuesName[cname]
	if !ok {
		log2.Errorf("missing translation path for %s in ValuesOverlaysToHelmValues", cname)
		return nil
	}
	pv := strings.Split(toPath, ".")
	cur := out
	for len(pv) > 1 {
		cur[pv[0]] = make(map[string]interface{})
		cur = cur[pv[0]].(map[string]interface{})
		pv = pv[1:]
	}
	cur[pv[0]] = in
	return out
}

// setEnablementAndNamespaces translates the enablement and namespace value of each component in the root values tree,
// based on feature/component inheritance relationship.
func (t *Translator) setEnablementAndNamespaces(root util.Tree, ii *v1alpha2.IstioControlPlaneSpec) error {
	for fn, f := range t.FeatureComponentToValues {
		for cn, c := range f {
			if err := setTree(root, util.PathFromString(c.enabled), name.IsComponentEnabled(string(fn), cn, ii)); err != nil {
				return err
			}
			if err := setTree(root, util.PathFromString(c.namespace), name.Namespace(string(fn), cn, ii)); err != nil {
				return err
			}
		}
	}
	return nil
}

// protoToValues takes an interface which must be a struct ptr and recursively iterates through all its fields.
// For each leaf, if looks for a mapping from the struct data path to the corresponding YAML path and if one is
// found, it calls the associated mapping function if one is defined to populate the values YAML path.
// If no mapping function is defined, it uses the default mapping function.
func (t *Translator) protoToValues(structPtr interface{}, root util.Tree, path util.Path) (errs util.Errors) {
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
			errs = util.AppendErrs(errs, t.protoToValues(fieldValue.Addr().Interface(), root, append(path, fieldName)))
		case reflect.Map:
			dbgPrint("Map")
			newPath := append(path, fieldName)
			for _, key := range fieldValue.MapKeys() {
				nnp := append(newPath, key.String())
				errs = util.AppendErrs(errs, t.insertLeaf(root, nnp, fieldValue.MapIndex(key)))
			}
		case reflect.Slice:
			dbgPrint("Slice")
			for i := 0; i < fieldValue.Len(); i++ {
				errs = util.AppendErrs(errs, t.protoToValues(fieldValue.Index(i).Elem().Interface(), root, path))
			}
		case reflect.Ptr:
			if util.IsNilOrInvalidValue(fieldValue.Elem()) {
				continue
			}
			newPath := append(path, fieldName)
			if fieldValue.Elem().Kind() == reflect.Struct {
				dbgPrint("Struct Ptr")
				errs = util.AppendErrs(errs, t.protoToValues(fieldValue.Interface(), root, newPath))
			} else {
				fmt.Println("Leaf Ptr")
				errs = util.AppendErrs(errs, t.insertLeaf(root, newPath, fieldValue))
			}
		default:
			dbgPrint("field has kind %s", kind)
			if structElems.Field(i).CanInterface() {
				errs = util.AppendErrs(errs, t.insertLeaf(root, append(path, fieldName), fieldValue))
			}
		}
	}
	return errs
}

func (t *Translator) insertLeaf(root util.Tree, newPath util.Path, fieldValue reflect.Value) (errs util.Errors) {
	// Must be a scalar leaf. See if we have a mapping.
	valuesPath, m := getValuesPathMapping(t.APIMapping, newPath)
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
// e.g. for mapping "a.b"  -> "1.2", the input path "a.b.c.d" would yield "1.2.c.d".
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

	if m.inPath == "" {
		return "", m
	}

	d := len(path) - len(p)
	out := m.inPath + "." + path[len(path)-d:].String()
	dbgPrint("translating %s to %s", path, out)
	return out, m
}

// setTree sets the YAML path in the given Tree to the given value, creating any required intermediate nodes.
func setTree(root util.Tree, path util.Path, value interface{}) error {
	dbgPrint("setTree %s:%v", path, value)
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
	setTree(root[path[0]].(util.Tree), path[1:], value)
	return nil
}

// defaultTranslationFunc is the default translation to values. It maps a Go data path into a YAML path.
func defaultTranslationFunc(m *Translation, root util.Tree, valuesPath string, value interface{}) error {
	var path []string

	if util.IsEmptyString(value) {
		dbgPrint("Skip empty string value for path %s", m.outPath)
		return nil
	}
	if valuesPath == "" {
		dbgPrint("Not mapping to values, resources path is %s", m.outPath)
		return nil
	}

	for _, p := range util.PathFromString(valuesPath) {
		path = append(path, firstCharToLower(p))
	}

	return setTree(root, path, value)
}

func dbgPrint(v ...interface{}) {
	return
	fmt.Println(fmt.Sprintf(v[0].(string), v[1:]...))
}

func firstCharToLower(s string) string {
	return strings.ToLower(s[0:1]) + s[1:]
}
