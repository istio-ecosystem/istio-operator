package component

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	protobuf "github.com/gogo/protobuf/types"
	"gopkg.in/yaml.v2"
	"istio.io/operator/pkg/apis/istio/v1alpha1"
	"istio.io/operator/pkg/helm"
	"istio.io/operator/pkg/patch"
	"istio.io/operator/pkg/util"

	"istio.io/pkg/log"
)

const (
	// IstioComponent names corresponding to the IstioControlPlane proto component names. Must be the same, since these
	// are used for struct traversal.
	IstioBaseComponentName       = "crds"
	PilotComponentName           = "Pilot"
	GalleyComponentName          = "Galley"
	SidecarInjectorComponentName = "SidecarInjector"
	PolicyComponentName          = "Policy"
	TelemetryComponentName       = "Telemetry"
	CitadelComponentName         = "Citadel"
	CertManagerComponentName     = "CertManager"
	NodeAgentComponentName       = "NodeAgent"
	IngressComponentName         = "Ingress"
	EgressComponentName          = "Egress"

	// String to emit for any component which is disabled.
	componentDisabledStr = " component is disabled."
	yamlCommentStr       = "# "

	// localFilePrefix is a prefix for local files.
	localFilePrefix = "file://"
)

// ComponentDirLayout is a mapping between a component name and a subdir path to its chart from the helm charts root.
type ComponentDirLayout map[string]string

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
	RenderManifest() (string, error)
}

// CommonComponentFields is a struct common to all components.
type CommonComponentFields struct {
	*ComponentOptions
	enabled   bool
	namespace string
	name      string
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

// disabledYAMLStr returns the YAML comment string that the given component is disabled.
func disabledYAMLStr(componentName string) string {
	return yamlCommentStr + componentName + componentDisabledStr
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
	return helm.OverlayYAML(string(by), string(py))
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

// isComponentEnabled reports whether the given feature and component are enabled in the given spec. The logic is, in
// order of evaluation:
// 1. if the feature is not defined, the component is disabled, else
// 2. if the feature is disabled, the component is disabled, else
// 3. if the component is not defined, it is reported disabled, else
// 4. if the component disabled, it is reported disabled, else
// 5. the component is enabled.
// This follows the logic description in IstioControlPlane proto.
func isComponentEnabled(featureName, componentName string, installSpec *v1alpha1.IstioControlPlaneSpec) bool {
	featureNodeI, err := GetFromStructPath(installSpec, featureName+".Enabled")
	if err != nil {
		log.Error(err.Error())
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

	componentNodeI, err := GetFromStructPath(installSpec, featureName+".Components."+componentName+".Enabled")
	if err != nil {
		log.Error(err.Error())
		return false
	}
	if componentNodeI == nil {
		return true
	}
	componentNode, ok := componentNodeI.(*protobuf.BoolValue)
	if !ok {
		log.Errorf("component %s enabled has bad type %T, expect *protobuf.BoolValue", componentNodeI)
	}
	if componentNode == nil {
		return false
	}
	return componentNode.Value
}

// renderManifest renders the manifest for the component defined by c and returns the resulting string.
func renderManifest(c *CommonComponentFields) (string, error) {
	if !isComponentEnabled(c.FeatureName, c.name, c.InstallSpec) {
		fmt.Printf("disabled\n")
		return disabledYAMLStr(c.name), nil
	}

	var vals, valsUnvalidated map[string]interface{}
	err := SetFromPath(c.ComponentOptions.InstallSpec, "TrafficManagement.Components."+c.name+".Common.ValuesOverrides", vals)
	if err != nil {
		return "", err
	}
	err = SetFromPath(c.ComponentOptions.InstallSpec, "TrafficManagement.Components."+c.name+".Common.UnvalidatedValuesOverrides", valsUnvalidated)
	if err != nil {
		return "", err
	}

	valsYAML, err := patchTree(vals, valsUnvalidated)
	if err != nil {
		return "", err
	}

	my, err := c.renderer.RenderManifest(valsYAML)
	if err != nil {
		return "", err
	}
	my += helm.YAMLSeparator + "\n"

	var overlays []*v1alpha1.K8SObjectOverlay
	err = SetFromPath(c.InstallSpec, "TrafficManagement.Components."+c.name+".Common.K8s.Overlays", overlays)
	if err != nil {
		return "", err
	}

	return patch.PatchYAMLManifest(my, c.namespace, overlays)
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
		return helm.NewFileTemplateRenderer(valuesPath, chartSubdir, c.name, c.namespace), nil
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

// SetFromPath sets out with the value at path from node. out is not set if the path doesn't exist or the value is nil.
// Node and all intermediate along path must be type struct ptr.
func SetFromPath(node interface{}, path string, out interface{}) error {
	val, err := GetFromStructPath(node, path)
	if err != nil {
		return err
	}
	if util.IsValueNil(val) {
		return nil
	}
	if reflect.TypeOf(val) != reflect.TypeOf(out) {
		return fmt.Errorf("SetFromPath from type %T != to type %T", val, out)
	}
	reflect.ValueOf(out).Set(reflect.ValueOf(val))
	return nil
}

// GetFromStructPath returns the value at path from the given node.
// Node and all intermediate along path must be type struct ptr.
func GetFromStructPath(node interface{}, path string) (interface{}, error) {
	return getFromStructPath(node, util.PathFromString(path))
}

// getFromStructPath is the internal implementation of GetFromStructPath which recurses through a tree of Go structs
// given a path. It terminates when the end of the path is reached or a path element does not exist.
func getFromStructPath(node interface{}, path util.Path) (interface{}, error) {
	if reflect.TypeOf(node).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("GetFromStructPath path %s, expected struct ptr, got %T", path, node)
	}
	structElems := reflect.ValueOf(node).Elem()
	if reflect.TypeOf(structElems).Kind() != reflect.Struct {
		return nil, fmt.Errorf("GetFromStructPath path %s, expected struct ptr, got %T", path, node)
	}

	if len(path) == 0 {
		return node, nil
	}

	if util.IsNilOrInvalidValue(structElems) {
		return nil, nil
	}

	for i := 0; i < structElems.NumField(); i++ {
		fieldName := structElems.Type().Field(i).Name

		if fieldName != path[0] {
			continue
		}

		fv := structElems.Field(i)
		kind := structElems.Type().Field(i).Type.Kind()
		if kind != reflect.Ptr {
			return nil, fmt.Errorf("struct field %s is %T, expect struct ptr", fieldName, fv.Interface())
		}

		return getFromStructPath(fv.Interface(), path[1:])
	}

	return nil, fmt.Errorf("path %s not found from node type %T", path, node)
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
