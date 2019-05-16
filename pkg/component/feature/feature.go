package feature

import (
	"fmt"
	"path/filepath"
	"strings"

	protobuf "github.com/gogo/protobuf/types"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/component/component"
	"github.com/ostromart/istio-installer/pkg/util"
)

type ComponentDirLayout map[string]string

var (
	V12 = ComponentDirLayout{
		component.PilotComponentName:           "istio-control/istio-discovery",
		component.GalleyComponentName:          "istio-control/istio-config",
		component.SidecarInjectorComponentName: "istio-control/istio-autoinject",
		component.PolicyComponentName:          "istio-policy",
		component.TelemetryComponentName:       "istio-telemetry",
		component.CitadelComponentName:         "security/citadel",
		component.NodeAgentComponentName:       "security/nodeagent",
		component.CertManagerComponentName:     "security/certmanager",
		component.IngressComponentName:         "gateways/istio-ingress",
		component.EgressComponentName:          "gateways/istio-egress",
	}
)

type Feature interface {
	RenderManifest() (string, util.Errors)
}

type FeatureOptions struct {
	InstallSpec   *v1alpha1.InstallSpec
	Dirs          ComponentDirLayout
	HelmChartName string
	HelmChartDir  string
}

type CommonFeatureFields struct {
	FeatureOptions
	enabled    bool
	namespace  string
	components []component.Component
}

type TrafficManagementFeature struct {
	CommonFeatureFields
}

func newComponentOptions(cff *CommonFeatureFields, componentName string) *component.ComponentOptions {
	return &component.ComponentOptions{
		InstallSpec:      cff.InstallSpec,
		FeatureEnabled:   cff.enabled,
		FeatureNamespace: cff.namespace,
		HelmChartName:    cff.HelmChartName,
		HelmChartDir:     filepath.Join(cff.HelmChartDir, cff.Dirs[componentName]),
	}
}

func NewTrafficManagementFeature(opts *FeatureOptions) *TrafficManagementFeature {
	cff := &CommonFeatureFields{
		FeatureOptions: *opts,
		enabled:        withOverrideBool(false, opts.InstallSpec.TrafficManagement.Enabled),
		namespace:      withOverrideString(opts.InstallSpec.DefaultNamespacePrefix, opts.InstallSpec.TrafficManagement.Namespace),
	}
	cff.components = []component.Component{
		component.NewPilotComponent(newComponentOptions(cff, component.PilotComponentName)),
		component.NewSidecarComponent(newComponentOptions(cff, component.SidecarInjectorComponentName)),
	}

	return &TrafficManagementFeature{
		CommonFeatureFields: *cff,
	}
}

func (f *TrafficManagementFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

type SecurityFeature struct {
	CommonFeatureFields
}

func NewSecurityFeature(install *v1alpha1.InstallSpec) *SecurityFeature {
	return &SecurityFeature{
		CommonFeatureFields: CommonFeatureFields{
			enabled:   install.Security.Enabled.Value,
			namespace: install.Security.Namespace,
			components: []component.Component{
				component.NewCitadelComponent(install),
				component.NewCertManagerComponent(install),
				component.NewNodeAgentComponent(install),
			},
		},
	}
}

func (f *SecurityFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

type PolicyFeature struct {
	CommonFeatureFields
}

func NewPolicyFeature(install *v1alpha1.InstallSpec) *PolicyFeature {
	return &PolicyFeature{
		CommonFeatureFields: CommonFeatureFields{
			enabled:   install.Policy.Enabled.Value,
			namespace: install.Policy.Namespace,
		},
	}
}

func (f *PolicyFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

type TelemetryFeature struct {
	CommonFeatureFields
}

func NewTelemetryFeature(install *v1alpha1.InstallSpec) *TelemetryFeature {
	return &TelemetryFeature{
		CommonFeatureFields: CommonFeatureFields{
			enabled:   install.Policy.Enabled.Value,
			namespace: install.Policy.Namespace,
		},
	}
}

func (f *TelemetryFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

type ConfigManagementFeature struct {
	CommonFeatureFields
}

func NewConfigManagementFeature(install *v1alpha1.InstallSpec) *ConfigManagementFeature {
	return &ConfigManagementFeature{
		CommonFeatureFields: CommonFeatureFields{
			enabled:   install.ConfigManagement.Enabled.Value,
			namespace: install.ConfigManagement.Namespace,
		},
	}
}

func (f *ConfigManagementFeature) RenderManifest() (string, util.Errors) {
	return renderComponents(f.components)
}

func renderComponents(cs []component.Component) (manifest string, errsOut util.Errors) {
	var sb strings.Builder
	for _, c := range cs {
		s, errs := c.RenderManifest()
		errsOut = util.AppendErrs(errsOut, errs)
		_, err := sb.WriteString(s)
		errsOut = util.AppendErr(errsOut, err)
	}
	if len(errsOut) > 0 {
		return "", errsOut
	}
	return sb.String(), nil
}

func withOverrideBool(base bool, override *protobuf.BoolValue) bool {
	if override == nil {
		return base
	}
	return override.Value
}

func withOverrideString(base string, override string) string {
	if override == "" {
		return base
	}
	return override
}

func checkOptions(fo *FeatureOptions) error {
	if fo.InstallSpec == nil {
		return fmt.Errorf("InstallSpec must be set")
	}
	if fo.HelmChartName == "" {
		return fmt.Errorf("HelmChartName must be set")
	}
	if fo.HelmChartDir == "" {
		return fmt.Errorf("HelmChartDir must be set")
	}
	if fo.Dirs == nil {
		return fmt.Errorf("Dirs must be set")
	}
	return nil
}
