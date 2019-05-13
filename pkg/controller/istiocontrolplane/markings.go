package istiocontrolplane

import (
	"strconv"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"istio.io/operator/pkg/helmreconciler"
)

const (
	// MetadataNamespace is the namespace for mesh metadata (labels, annotations)
	MetadataNamespace = "install.operator.istio.io"

	// OwnerNameKey represents the name of the owner to which the resource relates
	OwnerNameKey = MetadataNamespace + "/owner-name"
	// OwnerKindKey represents the kind of the owner to which the resource relates
	OwnerKindKey = MetadataNamespace + "/owner-kind"
	// OwnerGroupKey represents the group of the owner to which the resource relates
	OwnerGroupKey = MetadataNamespace + "/owner-group"

	// OwnerGenerationKey represents the generation to which the resource was last reconciled
	OwnerGenerationKey = MetadataNamespace + "/owner-generation"
)

var (
	// XXX: move this into a ConfigMap so users can override things if they add new types in customized charts
	// ordered by which types should be deleted, first to last
	namespacedResources = []schema.GroupVersionKind{
		{Group: "autoscaling", Version: "v2beta1", Kind: "HorizontalPodAutoscaler"},
		{Group: "policy", Version: "v1beta1", Kind: "PodDisruptionBudget"},
		{Group: "route.openshift.io", Version: "v1", Kind: "Route"},
		{Group: "apps.openshift.io", Version: "v1", Kind: "DeploymentConfig"},
		{Group: "apps", Version: "v1beta1", Kind: "Deployment"},
		{Group: "apps", Version: "v1beta1", Kind: "StatefulSet"},
		{Group: "apps", Version: "v1", Kind: "Deployment"},
		{Group: "batch", Version: "v1", Kind: "Job"},
		{Group: "extensions", Version: "v1beta1", Kind: "DaemonSet"},
		{Group: "extensions", Version: "v1beta1", Kind: "Deployment"},
		{Group: "extensions", Version: "v1beta1", Kind: "Ingress"},
		{Group: "", Version: "v1", Kind: "Service"},
		{Group: "", Version: "v1", Kind: "Endpoints"},
		{Group: "", Version: "v1", Kind: "ConfigMap"},
		{Group: "", Version: "v1", Kind: "PersistentVolumeClaim"},
		{Group: "", Version: "v1", Kind: "Pod"},
		{Group: "", Version: "v1", Kind: "Secret"},
		{Group: "", Version: "v1", Kind: "ServiceAccount"},
		{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Kind: "RoleBinding"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"},
		{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Kind: "Role"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"},
		{Group: "authentication.istio.io", Version: "v1alpha1", Kind: "Policy"},
		{Group: "config.istio.io", Version: "v1alpha2", Kind: "adapter"},
		{Group: "config.istio.io", Version: "v1alpha2", Kind: "attributemanifest"},
		{Group: "config.istio.io", Version: "v1alpha2", Kind: "handler"},
		{Group: "config.istio.io", Version: "v1alpha2", Kind: "kubernetes"},
		{Group: "config.istio.io", Version: "v1alpha2", Kind: "logentry"},
		{Group: "config.istio.io", Version: "v1alpha2", Kind: "metric"},
		{Group: "config.istio.io", Version: "v1alpha2", Kind: "rule"},
		{Group: "config.istio.io", Version: "v1alpha2", Kind: "template"},
		{Group: "networking.istio.io", Version: "v1alpha3", Kind: "DestinationRule"},
		{Group: "networking.istio.io", Version: "v1alpha3", Kind: "EnvoyFilter"},
		{Group: "networking.istio.io", Version: "v1alpha3", Kind: "Gateway"},
		{Group: "networking.istio.io", Version: "v1alpha3", Kind: "VirtualService"},
	}

	// ordered by which types should be deleted, first to last
	nonNamespacedResources = []schema.GroupVersionKind{
		{Group: "admissionregistration.k8s.io", Version: "v1beta1", Kind: "MutatingWebhookConfiguration"},
		{Group: "admissionregistration.k8s.io", Version: "v1beta1", Kind: "ValidatingWebhookConfiguration"},
		{Group: "certmanager.k8s.io", Version: "v1alpha1", Kind: "ClusterIssuer"},
		{Group: "oauth.openshift.io", Version: "v1", Kind: "OAuthClient"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"},
		{Group: "authentication.istio.io", Version: "v1alpha1", Kind: "MeshPolicy"},
	}
)

// IstioMarkingsFactory is a ResourceMarkingsFactory specific to IstioControlPlane resources.
type IstioMarkingsFactory struct{}

var _ helmreconciler.ResourceMarkingsFactory = &IstioMarkingsFactory{}

// NewResourceMarkings creates a new ResourceMarkings object specific to the instance.
func (f *IstioMarkingsFactory) NewResourceMarkings(instance runtime.Object) (helmreconciler.ResourceMarkings, error) {
	objectAccessor, err := meta.Accessor(instance)
	if err != nil {
		return nil, err
	}
	gvk := instance.GetObjectKind().GroupVersionKind()
	name := objectAccessor.GetName()
	generation := strconv.FormatInt(objectAccessor.GetGeneration(), 10)
	return &helmreconciler.SimpleResourceMarkings{
		OwnerLabels: map[string]string{
			OwnerNameKey:  name,
			OwnerGroupKey: gvk.Group,
			OwnerKindKey:  gvk.Kind,
		},
		OwnerAnnotations: map[string]string{
			OwnerGenerationKey: generation,
		},
		NamespacedResources:    namespacedResources,
		NonNamespacedResources: nonNamespacedResources,
	}, nil
}
