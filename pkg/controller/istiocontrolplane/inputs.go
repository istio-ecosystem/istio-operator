package istiocontrolplane

import (
	"fmt"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/helm/pkg/manifest"

	"istio.io/operator/pkg/apis/istio/v1alpha1"
	"istio.io/operator/pkg/helmreconciler"
)

// defaultProcessingOrder for the rendered charts
var defaultProcessingOrder = []string{
	"istio",
	"istio/charts/security",
	"istio/charts/galley",
	"istio/charts/prometheus",
	"istio/charts/mixer",
	"istio/charts/pilot",
	"istio/charts/gateways",
	"istio/charts/sidecarInjectorWebhook",
	"istio/charts/grafana",
	"istio/charts/tracing",
	"istio/charts/kiali",
}

// IstioInputFactory creates a RenderingInput specific to an IstioControlPlane instance.
type IstioInputFactory struct{}

var _ helmreconciler.RenderingInputFactory = &IstioInputFactory{}

// IstioRenderingInput is a RenderingInput specific to an IstioControlPlane instance.
type IstioRenderingInput struct {
	instance *v1alpha1.IstioControlPlane
}

var _ helmreconciler.RenderingInput = &IstioRenderingInput{}

// NewRenderingInput creates a new IstioRenderiongInput for the specified instance.
func (f *IstioInputFactory) NewRenderingInput(instance runtime.Object) (helmreconciler.RenderingInput, error) {
	istioControlPlane, ok := instance.(*v1alpha1.IstioControlPlane)
	if !ok {
		return nil, fmt.Errorf("object is not an IstioControlPlane resource")
	}
	return &IstioRenderingInput{instance: istioControlPlane}, nil
}

// GetChartPath returns the absolute path locating the charts to be rendered.
func (i *IstioRenderingInput) GetChartPath() string {
	path := i.instance.Spec.ChartPath
	if len(path) == 0 {
		return filepath.Join(controllerOptions.BaseChartPath, controllerOptions.DefaultChartPath)
	} else if filepath.IsAbs(path) {
		return i.instance.Spec.ChartPath
	}
	return filepath.Join(controllerOptions.BaseChartPath, path)
}

// GetValues returns the values that should be used when rendering the charts.
func (i *IstioRenderingInput) GetValues() map[string]interface{} {
	return i.instance.Spec.RawValues
}

// GetTargetNamespace returns the namespace within which rendered namespaced resources should be generated
// (i.e. Release.Namespace)
func (i *IstioRenderingInput) GetTargetNamespace() string {
	return i.instance.Namespace
}

// GetProcessingOrder returns the order in which the rendered charts should be processed.
func (i *IstioRenderingInput) GetProcessingOrder(manifests map[string][]manifest.Manifest) ([]string, error) {
	seen := map[string]struct{}{}
	ordering := make([]string, 0, len(manifests))
	// known ordering
	for _, chart := range defaultProcessingOrder {
		if _, ok := manifests[chart]; ok {
			ordering = append(ordering, chart)
			seen[chart] = struct{}{}
		}
	}
	// everything else to the end
	for chart := range manifests {
		if _, ok := seen[chart]; !ok {
			ordering = append(ordering, chart)
			seen[chart] = struct{}{}
		}
	}
	return ordering, nil
}
