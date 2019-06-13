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

// ComponentName is a component name string, typed to constrain allowed values.
type ComponentName string

const (
	// IstioComponent names corresponding to the IstioControlPlane proto component names. Must be the same, since these
	// are used for struct traversal.
	IstioBaseComponentName       ComponentName = "crds"
	PilotComponentName           ComponentName = "Pilot"
	GalleyComponentName          ComponentName = "Galley"
	SidecarInjectorComponentName ComponentName = "SidecarInjector"
	PolicyComponentName          ComponentName = "Policy"
	TelemetryComponentName       ComponentName = "Telemetry"
	CitadelComponentName         ComponentName = "Citadel"
	CertManagerComponentName     ComponentName = "CertManager"
	NodeAgentComponentName       ComponentName = "NodeAgent"
	IngressComponentName         ComponentName = "Ingress"
	EgressComponentName          ComponentName = "Egress"

	// String to emit for any component which is disabled.
	componentDisabledStr = " component is disabled."
	yamlCommentStr       = "# "

	// localFilePrefix is a prefix for local files.
	localFilePrefix = "file://"
)

// ComponentDirLayout is a mapping between a component name and a subdir path to its chart from the helm charts root.
type ComponentDirLayout map[ComponentName]string

var (
	// V12DirLayout is a ComponentDirLayout for Istio v1.2.
	V12DirLayout = ComponentDirLayout{
		PilotComponentName:           "istio-control/istio-discovery",
		GalleyComponentName:          "istio-control/istio-config",
		SidecarInjectorComponentName: "istio-control/istio-autoinject",
		PolicyComponentName:          "istio-policy",
		TelemetryComponentName:       "istio-telemetry",
		CitadelComponentName:         "security/citadel",
		NodeAgentComponentName:       "security/nodeagent",
		CertManagerComponentName:     "security/certmanager",
		IngressComponentName:         "gateways/istio-ingress",
		EgressComponentName:          "gateways/istio-egress",
	}
	// componentToHelmValuesName is the root component name used in values YAML files in component charts.
	componentToHelmValuesName = map[ComponentName]string{
		PilotComponentName:           "pilot",
		GalleyComponentName:          "galley",
		SidecarInjectorComponentName: "sidecarInjectorWebhook",
		PolicyComponentName:          "mixer.policy",
		TelemetryComponentName:       "mixer.telemetry",
		CitadelComponentName:         "citadel",
		NodeAgentComponentName:       "nodeAgent",
		CertManagerComponentName:     "certManager",
		IngressComponentName:         "gateways.istio-ingressgateway",
		EgressComponentName:          "gateways.istio-ingressgateway",
	}
)

// ComponentOptions defines options for a component.
type ComponentOptions struct {
	FeatureName string
	InstallSpec *v1alpha1.IstioControlPlaneSpec
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
	name      ComponentName
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
			name:             PilotComponentName,
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
	if !isComponentEnabled(c.FeatureName, c.name, c.InstallSpec) {
		return disabledYAMLStr(c.name), nil
	}

	vals, valsUnvalidated := make(map[string]interface{}), make(map[string]interface{})
	validatedExist, err := SetFromPath(c.ComponentOptions.InstallSpec, "TrafficManagement.Components."+string(c.name)+".Common.ValuesOverrides", &vals)
	if err != nil {
		return "", err
	}
	unvalidatedExist, err := SetFromPath(c.ComponentOptions.InstallSpec, "TrafficManagement.Components."+string(c.name)+".Common.UnvalidatedValuesOverrides", &valsUnvalidated)
	if err != nil {
		return "", err
	}

	vals = valuesOverlaysToHelmValues(vals, c.name)
	valsUnvalidated = valuesOverlaysToHelmValues(valsUnvalidated, c.name)
	valsYAML, err := patchTree(vals, valsUnvalidated)
	if err != nil {
		return "", err
	}
	if validatedExist || unvalidatedExist {
		log.Infof("patched values:\n%s\n", valsYAML)
	}

	my, err := c.renderer.RenderManifest(valsYAML)
	if err != nil {
		return "", err
	}
	my += helm.YAMLSeparator + "\n"

	var overlays []*v1alpha1.K8SObjectOverlay
	found, err := SetFromPath(c.InstallSpec, "TrafficManagement.Components."+string(c.name)+".Common.K8S.Overlays", &overlays)
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

// isComponentEnabled reports whether the given feature and component are enabled in the given spec. The logic is, in
// order of evaluation:
// 1. if the feature is not defined, the component is disabled, else
// 2. if the feature is disabled, the component is disabled, else
// 3. if the component is not defined, it is reported disabled, else
// 4. if the component disabled, it is reported disabled, else
// 5. the component is enabled.
// This follows the logic description in IstioControlPlane proto.
func isComponentEnabled(featureName string, componentName ComponentName, installSpec *v1alpha1.IstioControlPlaneSpec) bool {
	featureNodeI, found, err := GetFromStructPath(installSpec, featureName+".Enabled")
	if err != nil {
		log.Error(err.Error())
		return false
	}
	if !found {
		return false
	}
	if featureNodeI == nil {
		return false
	}
	featureNode, ok := featureNodeI.(*protobuf.BoolValue)
	if !ok {
		log.Errorf("feature %s enabled has bad type %T, expect *protobuf.BoolValue", featureNodeI)
	}
	if featureNode == nil {
		return false
	}
	if featureNode.Value == false {
		return false
	}

	componentNodeI, found, err := GetFromStructPath(installSpec, featureName+".Components."+string(componentName)+".Common.Enabled")
	if err != nil {
		log.Error(err.Error())
		return featureNode.Value
	}
	if !found {
		return featureNode.Value
	}
	if componentNodeI == nil {
		return featureNode.Value
	}
	componentNode, ok := componentNodeI.(*protobuf.BoolValue)
	if !ok {
		log.Errorf("component %s enabled has bad type %T, expect *protobuf.BoolValue", componentNodeI)
		return featureNode.Value
	}
	if componentNode == nil {
		return featureNode.Value
	}
	return componentNode.Value
}

// disabledYAMLStr returns the YAML comment string that the given component is disabled.
func disabledYAMLStr(componentName ComponentName) string {
	return yamlCommentStr + string(componentName) + componentDisabledStr
}

// patchTree patches the tree represented by patch over the tree represented by base and returns a YAML string of the
// result.
func patchTree(base, patch map[string]interface{}) (string, error) {
	by, err := yaml.Marshal(base)
	if err != nil {
		return "", err
	}
	py, err := yaml.Marshal(patch)
	if err != nil {
		return "", err
	}
	//fmt.Printf("base:\n%s\n\npatch:\n%s\n", string(by), string(py))
	return helm.OverlayYAML(string(by), string(py))
}

func valuesOverlaysToHelmValues(in map[string]interface{}, cname ComponentName) map[string]interface{} {
	out := make(map[string]interface{})
	toPath, ok := componentToHelmValuesName[cname]
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
	switch {
	case cp == "":
		return nil, fmt.Errorf("compiled in CustomPackagePath not yet supported")
	case isFilePath(cp):
		chartRoot := filepath.Join(getLocalFilePath(cp))
		chartSubdir := filepath.Join(chartRoot, c.Dirs[c.name])
		valuesPath := getValuesFilename(c.InstallSpec)
		if !isFilePath(valuesPath) {
			valuesPath = filepath.Join(chartRoot, valuesPath)
		}
		return helm.NewFileTemplateRenderer(valuesPath, chartSubdir, string(c.name), c.namespace), nil
	default:
	}
	return nil, fmt.Errorf("unsupported CustomPackagePath %s", cp)
}

// isFilePath reports whether the given URL is a local file path.
func isFilePath(path string) bool {
	return strings.HasPrefix(path, localFilePrefix)
}

// getLocalFilePath returns the local file path string of the form /a/b/c, given a file URL of the form file:///a/b/c
func getLocalFilePath(path string) string {
	return strings.TrimPrefix(path, localFilePrefix)
}

// getValuesFilename returns the global values filename, given an IstioControlPlaneSpec.
func getValuesFilename(i *v1alpha1.IstioControlPlaneSpec) string {
	if i.BaseSpecPath == "" {
		return helm.DefaultGlobalValuesFilename
	}
	return i.BaseSpecPath
}

// TODO: move these out to a separate package.
// SetFromPath sets out with the value at path from node. out is not set if the path doesn't exist or the value is nil.
// All intermediate along path must be type struct ptr. Out must be either a struct ptr or map ptr.
func SetFromPath(node interface{}, path string, out interface{}) (bool, error) {
	val, found, err := GetFromStructPath(node, path)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	if util.IsValueNil(val) {
		return true, nil
	}

	return true, Set(val, out)
}

// Set sets out with the value at path from node. out is not set if the path doesn't exist or the value is nil.
func Set(val, out interface{}) error {
	// Special case: map out type must be set through map ptr.
	if util.IsMap(val) && util.IsMapPtr(out) {
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(val))
		return nil
	}
	if util.IsSlice(val) && util.IsSlicePtr(out) {
		reflect.ValueOf(out).Elem().Set(reflect.ValueOf(val))
		return nil
	}

	if reflect.TypeOf(val) != reflect.TypeOf(out) {
		return fmt.Errorf("SetFromPath from type %T != to type %T, %v", val, out, util.IsSlicePtr(out))
	}

	if !reflect.ValueOf(out).CanSet() {
		return fmt.Errorf("can't set %v(%T) to out type %T", val, val, out)
	}
	reflect.ValueOf(out).Set(reflect.ValueOf(val))
	return nil
}

// GetFromStructPath returns the value at path from the given node, or false if the path does not exist.
// Node and all intermediate along path must be type struct ptr.
func GetFromStructPath(node interface{}, path string) (interface{}, bool, error) {
	return getFromStructPath(node, util.PathFromString(path))
}

// getFromStructPath is the internal implementation of GetFromStructPath which recurses through a tree of Go structs
// given a path. It terminates when the end of the path is reached or a path element does not exist.
func getFromStructPath(node interface{}, path util.Path) (interface{}, bool, error) {
	kind := reflect.TypeOf(node).Kind()
	var structElems reflect.Value
	switch kind {
	case reflect.Map, reflect.Slice:
		if len(path) != 0 {
			return nil, false, fmt.Errorf("GetFromStructPath path %s, unsupported leaf type %T", path, node)
		}
	case reflect.Ptr:
		structElems = reflect.ValueOf(node).Elem()
		if reflect.TypeOf(structElems).Kind() != reflect.Struct {
			return nil, false, fmt.Errorf("GetFromStructPath path %s, expected struct ptr, got %T", path, node)
		}
	default:
		return nil, false, fmt.Errorf("GetFromStructPath path %s, unsupported type %T", path, node)
	}
	if len(path) == 0 {
		return node, true, nil
	}

	if util.IsNilOrInvalidValue(structElems) {
		return nil, false, nil
	}

	for i := 0; i < structElems.NumField(); i++ {
		fieldName := structElems.Type().Field(i).Name

		if fieldName != path[0] {
			continue
		}

		fv := structElems.Field(i)
		kind = structElems.Type().Field(i).Type.Kind()
		if kind != reflect.Ptr && kind != reflect.Map && kind != reflect.Slice {
			return nil, false, fmt.Errorf("struct field %s is %T, expect struct ptr, map or slice", fieldName, fv.Interface())
		}

		return getFromStructPath(fv.Interface(), path[1:])
	}

	return nil, false, nil
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
