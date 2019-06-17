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

/*
Package component defines an in-memory representation of IstioControlPlane.<Feature>.<Component>. It provides functions
for manipulating the component and rendering a manifest from it.
See ../README.md for an architecture overview.
*/
package component

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	protobuf "github.com/gogo/protobuf/types"
	"istio.io/operator/pkg/apis/istio/v1alpha1"
	"istio.io/operator/pkg/helm"
	"istio.io/operator/pkg/patch"
	"istio.io/operator/pkg/util"

	"istio.io/pkg/log"
)

// ComponentDirLayout is a mapping between a component name and a subdir path to its chart from the helm charts root.
type ComponentDirLayout map[name.ComponentName]string

const (
	// String to emit for any component which is disabled.
	componentDisabledStr = " component is disabled."
	yamlCommentStr       = "# "
)

var (
	// V12DirLayout is a ComponentDirLayout for Istio v1.2.
	V12DirLayout = ComponentDirLayout{
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
	}
)

// ComponentOptions defines options for a component.
type ComponentOptions struct {
	FeatureName string
	InstallSpec *v1alpha2.IstioControlPlaneSpec
	Dirs        ComponentDirLayout
}

// IstioComponent defines the interface for a component.
type IstioComponent interface {
	// Run starts the component. Must me called before the component is used.
	Run() error
	// RenderManifest returns a string with the rendered manifest for the component.
	RenderManifest() (string, error)
}

// CommonComponentFields is a struct common to all components.
type CommonComponentFields struct {
	*ComponentOptions
	enabled   bool
	namespace string
	name      name.ComponentName
	renderer  helm.TemplateRenderer
	started   bool
}

// PilotComponent is the pilot component.
type PilotComponent struct {
	*CommonComponentFields
}

// NewPilotComponent creates a new PilotComponent and returns a pointer to it.
func NewPilotComponent(opts *ComponentOptions) *PilotComponent {
	ret := &PilotComponent{
		&CommonComponentFields{
			ComponentOptions: opts,
			name:             name.PilotComponentName,
		},
	}
	return ret
}

// Run implements the IstioComponent interface.
func (c *PilotComponent) Run() error {
	return runComponent(c.CommonComponentFields)
}

// RenderManifest implements the IstioComponent interface.
func (c *PilotComponent) RenderManifest() (string, error) {
	if !c.started {
		return "", fmt.Errorf("component %s not started in RenderManifest", c.name)
	}
	return renderManifest(c.CommonComponentFields)
}

// runComponent performs startup tasks for the component defined by the given CommonComponentFields.
func runComponent(c *CommonComponentFields) error {
	r, err := createHelmRenderer(c)
	if err != nil {
		return err
	}
	if err := r.Run(); err != nil {
		return err
	}
	c.renderer = r
	c.started = true
	return nil
}

// renderManifest renders the manifest for the component defined by c and returns the resulting string.
func renderManifest(c *CommonComponentFields) (string, error) {
	if !name.IsComponentEnabled(c.FeatureName, c.name, c.InstallSpec) {
		return disabledYAMLStr(c.name), nil
	}

	// First, translate the IstioControlPlane API to helm Values.
	apiVals, err := translate.ProtoToValues(translate.V12Mappings, c.InstallSpec)
	if err != nil {
		return "", err
	}

	// Second, add any overlays coming from IstioControlPlane.Value and IstioControlPlane.Feature.Component values and
	// unvalidatedValues.
	globalVals, vals, valsUnvalidated := make(map[string]interface{}), make(map[string]interface{}), make(map[string]interface{})
	_, err = name.SetFromPath(c.ComponentOptions.InstallSpec, "Values", &globalVals)
	if err != nil {
		return "", err
	}
	_, err = name.SetFromPath(c.ComponentOptions.InstallSpec, "TrafficManagement.Components."+string(c.name)+".Common.ValuesOverrides", &vals)
	if err != nil {
		return "", err
	}
	_, err = name.SetFromPath(c.ComponentOptions.InstallSpec, "TrafficManagement.Components."+string(c.name)+".Common.UnvalidatedValuesOverrides", &valsUnvalidated)
	if err != nil {
		return "", err
	}

	globalVals = valuesOverlaysToHelmValues(vals, name.IstioBaseComponentName)
	vals = valuesOverlaysToHelmValues(vals, c.name)
	valsUnvalidated = valuesOverlaysToHelmValues(valsUnvalidated, c.name)
	valsYAML, err := mergeTrees(apiVals, globalVals, vals, valsUnvalidated)
	if err != nil {
		return "", err
	}

	log.Infof("values from IstioControlPlane:\n%s\noverlay values:\n%s\nunvalidate overlay:\n%s\nmerged values:\n%s\n",
		apiVals, vals, valsUnvalidated, valsYAML)

	my, err := c.renderer.RenderManifest(valsYAML)
	if err != nil {
		return "", err
	}
	my += helm.YAMLSeparator + "\n"

	var overlays []*v1alpha2.K8SObjectOverlay
	found, err := name.SetFromPath(c.InstallSpec, "TrafficManagement.Components."+string(c.name)+".Common.K8S.Overlays", &overlays)
	if err != nil {
		return "", err
	}
	if !found {
		return my, nil
	}
	kyo, _ := yaml.Marshal(overlays)
	log.Infof("kubernetes overlay: \n%s\n", kyo)
	return patch.PatchYAMLManifest(my, c.namespace, overlays)
}

// disabledYAMLStr returns the YAML comment string that the given component is disabled.
func disabledYAMLStr(componentName name.ComponentName) string {
	return yamlCommentStr + string(componentName) + componentDisabledStr
}

// mergeTrees overlays global values, component values and unvalidatedValues (in that order) over the YAML tree in
// apiValues and returns the result.
// The merge operation looks something like this (later items are merged on top of earlier ones):
// - values derived from translating IstioControlPlane to values
// - values in top level IstioControlPlane
// - values from component
// - unvalidateValues from component
func mergeTrees(apiValues string, globalVals, values, unvalidatedValues map[string]interface{}) (string, error) {
	gy, err := yaml.Marshal(globalVals)
	if err != nil {
		return "", err
	}
	by, err := yaml.Marshal(values)
	if err != nil {
		return "", err
	}
	py, err := yaml.Marshal(unvalidatedValues)
	if err != nil {
		return "", err
	}
	//fmt.Printf("values:\n%s\n\npatch:\n%s\n", string(by), string(py))
	yo, err := helm.OverlayYAML(apiValues, string(gy))
	if err != nil {
		return "", err
	}
	yyo, err := helm.OverlayYAML(yo, string(by))
	if err != nil {
		return "", err
	}
	return helm.OverlayYAML(yyo, string(py))
}

func valuesOverlaysToHelmValues(in map[string]interface{}, cname name.ComponentName) map[string]interface{} {
	out := make(map[string]interface{})
	toPath, ok := translate.ComponentToHelmValuesName[cname]
	if !ok {
		log.Errorf("missing translation path for %s in valuesOverlaysToHelmValues", cname)
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

// createHelmRenderer creates a helm renderer for the component defined by c and returns a ptr to it.
func createHelmRenderer(c *CommonComponentFields) (helm.TemplateRenderer, error) {
	cp := c.InstallSpec.CustomPackagePath
	chartSubdir := ""
	switch {
	case cp == "":
		return nil, fmt.Errorf("compiled in CustomPackagePath not yet supported")
	case util.IsFilePath(cp):
		chartRoot := filepath.Join(util.GetLocalFilePath(cp))
		chartSubdir = filepath.Join(chartRoot, c.Dirs[c.name])
	default:
		return nil, fmt.Errorf("unsupported CustomPackagePath type: %s", cp)
	}
	vp := c.InstallSpec.BaseProfilePath
	valuesPath := ""
	switch {
	case vp == "":
		return nil, fmt.Errorf("compiled in CustomPackagePath not yet supported")
	case util.IsFilePath(vp):
		valuesPath = util.GetLocalFilePath(vp)
	default:
		return nil, fmt.Errorf("unsupported BaseProfilePath type: %s", cp)
	}
	return helm.NewFileTemplateRenderer(valuesPath, chartSubdir, string(c.name), c.namespace), nil

}

// getValuesFilename returns the global values filename, given an IstioControlPlaneSpec.
func getValuesFilename(i *v1alpha2.IstioControlPlaneSpec) string {
	if i.BaseProfilePath == "" {
		return helm.DefaultGlobalValuesFilename
	}
	return i.BaseProfilePath
}

// TODO: implement below components once Pilot looks good.
type ProxyComponent struct {
}

func NewProxyComponent(opts *ComponentOptions) *ProxyComponent {
	return nil
}

func (c *ProxyComponent) Run() error {
	return nil
}

func (c *ProxyComponent) RenderManifest() (string, error) {
	return "", nil
}

type CitadelComponent struct {
}

func NewCitadelComponent(opts *ComponentOptions) *CitadelComponent {
	return nil
}

func (c *CitadelComponent) Run() error {
	return nil
}

func (c *CitadelComponent) RenderManifest() (string, error) {
	return "", nil
}

type CertManagerComponent struct {
}

func NewCertManagerComponent(opts *ComponentOptions) *CertManagerComponent {
	return nil
}

func (c *CertManagerComponent) Run() error {
	return nil
}

func (c *CertManagerComponent) RenderManifest() (string, error) {
	return "", nil
}

type NodeAgentComponent struct {
}

func NewNodeAgentComponent(opts *ComponentOptions) *NodeAgentComponent {
	return nil
}

func (c *NodeAgentComponent) Run() error {
	return nil
}

func (c *NodeAgentComponent) RenderManifest() (string, error) {
	return "", nil
}

type PolicyComponent struct {
}

func NewPolicyComponent(opts *ComponentOptions) *PolicyComponent {
	return nil
}

func (c *PolicyComponent) Run() error {
	return nil
}

func (c *PolicyComponent) RenderManifest() (string, error) {
	return "", nil
}

type TelemetryComponent struct {
}

func NewTelemetryComponent(opts *ComponentOptions) *TelemetryComponent {
	return nil
}

func (c *TelemetryComponent) Run() error {
	return nil
}

func (c *TelemetryComponent) RenderManifest() (string, error) {
	return "", nil
}

type GalleyComponent struct {
}

func NewGalleyComponent(opts *ComponentOptions) *GalleyComponent {
	return nil
}

func (c *GalleyComponent) Run() error {
	return nil
}

func (c *GalleyComponent) RenderManifest() (string, error) {
	return "", nil
}

type SidecarInjectorComponent struct {
}

func NewSidecarInjectorComponent(opts *ComponentOptions) *SidecarInjectorComponent {
	return nil
}

func (c *SidecarInjectorComponent) Run() error {
	return nil
}

func (c *SidecarInjectorComponent) RenderManifest() (string, error) {
	return "", nil
}
