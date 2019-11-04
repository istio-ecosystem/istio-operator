// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: pkg/apis/istio/v1alpha2/istiocontrolplane_types.proto

// IstioControlPlane is a schema for both defining and customizing Istio control plane installations.
// Running the operator with an empty user defined InstallSpec results in an control plane with default values, using the
// default charts.
//
// The simplest install specialization is to point the user InstallSpec profile to a different values file, for
// example an Istio minimal control plane, which will use the values associated with the minimal control plane profile for
// Istio.
//
// Deeper customization is possible at three levels:
//
// 1. New APIs defined in this file
//
//     Feature API: this API groups an Istio install by features and allows enabling/disabling the features, selecting base
//     control plane profiles, as well as some additional high level settings that are feature specific. Each feature contains
//     one or more components, which correspond to Istio components (Pods) in the cluster.
//
//     k8s API: this API is a pass through to k8s resource settings for Istio k8s resources. It allows customizing Istio k8s
//     resources like Affinity, Resource requests/limits, PodDisruptionBudgetSpec, Selectors etc. in a more consistent and
//     k8s specific way compared to values.yaml. See KubernetesResourcesSpec in this file for details.
//
// 1. values.yaml
//
//     The entirety of values.yaml settings is accessible through InstallSpec (see CommonComponentSpec/Values).
//     This API will gradually be deprecated and values there will be moved either into CRDs that are used to directly
//     configure components or, in the case of k8s settings, will be replaced by the new API above.
//
// 1. k8s resource overlays
//
//     Once a manifest is rendered from InstallSpec, a further customization can be applied by specifying k8s resource
//     overlays. The concept is similar to kustomize, where JSON patches are applied for object paths. This allows
//     customization at the lowest level and eliminates the need to create ad-hoc template parameters, or edit templates.
//
// Here are a few example uses:
//
// 1. Default Istio install
//
//     ```
//     spec:
//     ```
//
// 1. Default minimal profile install
//
//     ```
//     spec:
//       profile: minimal
//     ```
//
// 1. Default install with telemetry disabled
//
//     ```
//     spec:
//       telemetry:
//         enabled: false
//     ```
//
// 1. Default install with each feature installed to different namespace and security components in separate namespaces
//
//     ```
//     spec:
//       traffic_management:
//         components:
//           namespace: istio-traffic-management
//       policy:
//         components:
//           namespace: istio-policy
//       telemetry:
//         components:
//           namespace: istio-telemetry
//       config_management:
//         components:
//           namespace: istio-config-management
//       security:
//         components:
//           citadel:
//             namespace: istio-citadel
//           cert_manager:
//             namespace: istio-cert-manager
//           node_agent:
//             namespace: istio-node-agent
//     ```
//
// 1. Default install with specialized k8s settings for pilot
//
//     ```
//     spec:
//       traffic_management:
//         components:
//           pilot:
//             k8s:
//               resources:
//                 limits:
//                   cpu: 444m
//                   memory: 333Mi
//                 requests:
//                   cpu: 222m
//                   memory: 111Mi
//               readinessProbe:
//                 failureThreshold: 44
//                 initialDelaySeconds: 11
//                 periodSeconds: 22
//                 successThreshold: 33
//     ```
//
// 1. Default install with values.yaml customizations for proxy
//
//     ```
//     spec:
//       traffic_management:
//         components:
//           proxy:
//             values:
//             - global.proxy.enableCoreDump: true
//             - global.proxy.dnsRefreshRate: 10s
//     ```
//
// 1. Default install with modification to container flag in galley
//
//     ```
//     spec:
//       configuration_management:
//         components:
//           galley:
//             k8s:
//               overlays:
//               - apiVersion: extensions/v1beta1
//                 kind: Deployment
//                 name: istio-galley
//                 patches:
//                 - path: spec.template.spec.containers.[name:galley].command.[--livenessProbeInterval]
//                   value: --livenessProbeInterval=123s
//     ```

package v1alpha2

import (
	bytes "bytes"
	fmt "fmt"
	github_com_gogo_protobuf_jsonpb "github.com/gogo/protobuf/jsonpb"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/protobuf/google/protobuf"
	_ "k8s.io/api/autoscaling/v2beta1"
	_ "k8s.io/api/core/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// MarshalJSON is a custom marshaler for IstioControlPlane
func (this *IstioControlPlane) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioControlPlane
func (this *IstioControlPlane) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for IstioControlPlaneSpec
func (this *IstioControlPlaneSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IstioControlPlaneSpec
func (this *IstioControlPlaneSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TrafficManagementFeatureSpec
func (this *TrafficManagementFeatureSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TrafficManagementFeatureSpec
func (this *TrafficManagementFeatureSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TrafficManagementFeatureSpec_Components
func (this *TrafficManagementFeatureSpec_Components) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TrafficManagementFeatureSpec_Components
func (this *TrafficManagementFeatureSpec_Components) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for PolicyFeatureSpec
func (this *PolicyFeatureSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for PolicyFeatureSpec
func (this *PolicyFeatureSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for PolicyFeatureSpec_Components
func (this *PolicyFeatureSpec_Components) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for PolicyFeatureSpec_Components
func (this *PolicyFeatureSpec_Components) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TelemetryFeatureSpec
func (this *TelemetryFeatureSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TelemetryFeatureSpec
func (this *TelemetryFeatureSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TelemetryFeatureSpec_Components
func (this *TelemetryFeatureSpec_Components) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TelemetryFeatureSpec_Components
func (this *TelemetryFeatureSpec_Components) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for SecurityFeatureSpec
func (this *SecurityFeatureSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for SecurityFeatureSpec
func (this *SecurityFeatureSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for SecurityFeatureSpec_Components
func (this *SecurityFeatureSpec_Components) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for SecurityFeatureSpec_Components
func (this *SecurityFeatureSpec_Components) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ConfigManagementFeatureSpec
func (this *ConfigManagementFeatureSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ConfigManagementFeatureSpec
func (this *ConfigManagementFeatureSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ConfigManagementFeatureSpec_Components
func (this *ConfigManagementFeatureSpec_Components) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ConfigManagementFeatureSpec_Components
func (this *ConfigManagementFeatureSpec_Components) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for AutoInjectionFeatureSpec
func (this *AutoInjectionFeatureSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for AutoInjectionFeatureSpec
func (this *AutoInjectionFeatureSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for AutoInjectionFeatureSpec_Components
func (this *AutoInjectionFeatureSpec_Components) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for AutoInjectionFeatureSpec_Components
func (this *AutoInjectionFeatureSpec_Components) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for GatewayFeatureSpec
func (this *GatewayFeatureSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for GatewayFeatureSpec
func (this *GatewayFeatureSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for GatewayFeatureSpec_Components
func (this *GatewayFeatureSpec_Components) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for GatewayFeatureSpec_Components
func (this *GatewayFeatureSpec_Components) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for CNIFeatureSpec
func (this *CNIFeatureSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for CNIFeatureSpec
func (this *CNIFeatureSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for CNIFeatureSpec_Components
func (this *CNIFeatureSpec_Components) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for CNIFeatureSpec_Components
func (this *CNIFeatureSpec_Components) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for PilotComponentSpec
func (this *PilotComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for PilotComponentSpec
func (this *PilotComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ProxyComponentSpec
func (this *ProxyComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ProxyComponentSpec
func (this *ProxyComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for SidecarInjectorComponentSpec
func (this *SidecarInjectorComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for SidecarInjectorComponentSpec
func (this *SidecarInjectorComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for PolicyComponentSpec
func (this *PolicyComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for PolicyComponentSpec
func (this *PolicyComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TelemetryComponentSpec
func (this *TelemetryComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TelemetryComponentSpec
func (this *TelemetryComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for CitadelComponentSpec
func (this *CitadelComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for CitadelComponentSpec
func (this *CitadelComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for CertManagerComponentSpec
func (this *CertManagerComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for CertManagerComponentSpec
func (this *CertManagerComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for NodeAgentComponentSpec
func (this *NodeAgentComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for NodeAgentComponentSpec
func (this *NodeAgentComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for GalleyComponentSpec
func (this *GalleyComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for GalleyComponentSpec
func (this *GalleyComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for IngressGatewayComponentSpec
func (this *IngressGatewayComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for IngressGatewayComponentSpec
func (this *IngressGatewayComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for EgressGatewayComponentSpec
func (this *EgressGatewayComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for EgressGatewayComponentSpec
func (this *EgressGatewayComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for CNIComponentSpec
func (this *CNIComponentSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for CNIComponentSpec
func (this *CNIComponentSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for KubernetesResourcesSpec
func (this *KubernetesResourcesSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for KubernetesResourcesSpec
func (this *KubernetesResourcesSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for K8SObjectOverlay
func (this *K8SObjectOverlay) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for K8SObjectOverlay
func (this *K8SObjectOverlay) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for K8SObjectOverlay_PathValue
func (this *K8SObjectOverlay_PathValue) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for K8SObjectOverlay_PathValue
func (this *K8SObjectOverlay_PathValue) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for InstallStatus
func (this *InstallStatus) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for InstallStatus
func (this *InstallStatus) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for InstallStatus_VersionStatus
func (this *InstallStatus_VersionStatus) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for InstallStatus_VersionStatus
func (this *InstallStatus_VersionStatus) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for Resources
func (this *Resources) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for Resources
func (this *Resources) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ReadinessProbe
func (this *ReadinessProbe) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ReadinessProbe
func (this *ReadinessProbe) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ExecAction
func (this *ExecAction) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ExecAction
func (this *ExecAction) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for HTTPGetAction
func (this *HTTPGetAction) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for HTTPGetAction
func (this *HTTPGetAction) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for HTTPHeader
func (this *HTTPHeader) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for HTTPHeader
func (this *HTTPHeader) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TCPSocketAction
func (this *TCPSocketAction) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TCPSocketAction
func (this *TCPSocketAction) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for PodDisruptionBudgetSpec
func (this *PodDisruptionBudgetSpec) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for PodDisruptionBudgetSpec
func (this *PodDisruptionBudgetSpec) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for DeploymentStrategy
func (this *DeploymentStrategy) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for DeploymentStrategy
func (this *DeploymentStrategy) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for RollingUpdateDeployment
func (this *RollingUpdateDeployment) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for RollingUpdateDeployment
func (this *RollingUpdateDeployment) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for ObjectMeta
func (this *ObjectMeta) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for ObjectMeta
func (this *ObjectMeta) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TypeMapStringInterface
func (this *TypeMapStringInterface) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TypeMapStringInterface
func (this *TypeMapStringInterface) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TypeInterface
func (this *TypeInterface) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TypeInterface
func (this *TypeInterface) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TypeIntOrStringForPB
func (this *TypeIntOrStringForPB) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TypeIntOrStringForPB
func (this *TypeIntOrStringForPB) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

// MarshalJSON is a custom marshaler for TypeBoolValueForPB
func (this *TypeBoolValueForPB) MarshalJSON() ([]byte, error) {
	str, err := IstiocontrolplaneTypesMarshaler.MarshalToString(this)
	return []byte(str), err
}

// UnmarshalJSON is a custom unmarshaler for TypeBoolValueForPB
func (this *TypeBoolValueForPB) UnmarshalJSON(b []byte) error {
	return IstiocontrolplaneTypesUnmarshaler.Unmarshal(bytes.NewReader(b), this)
}

var (
	IstiocontrolplaneTypesMarshaler   = &github_com_gogo_protobuf_jsonpb.Marshaler{}
	IstiocontrolplaneTypesUnmarshaler = &github_com_gogo_protobuf_jsonpb.Unmarshaler{}
)
