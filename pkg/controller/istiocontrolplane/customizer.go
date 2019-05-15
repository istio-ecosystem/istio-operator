package istiocontrolplane

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"istio.io/operator/pkg/apis/istio/v1alpha1"
	"istio.io/operator/pkg/helmreconciler"
)

type IstioRenderingCustomizerFactory struct{}

var _ helmreconciler.RenderingCustomizerFactory

// NewCustomizer returns a RenderingCustomizer for Istio
func (f *IstioRenderingCustomizerFactory) NewCustomizer(instance runtime.Object) (helmreconciler.RenderingCustomizer, error) {
	istioControlPlane, ok := instance.(*v1alpha1.IstioControlPlane)
	if !ok {
		return nil, fmt.Errorf("object is not an IstioControlPlane resource")
	}
	return &helmreconciler.SimpleRenderingCustomizer{
		InputValue:    NewIstioRenderingInput(istioControlPlane),
		MarkingsValue: NewIstioResourceMarkings(istioControlPlane),
		ListenerValue: NewIstioRenderingListener(istioControlPlane),
	}, nil
}
