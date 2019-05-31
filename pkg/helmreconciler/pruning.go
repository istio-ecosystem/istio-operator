package helmreconciler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *HelmReconciler) prune(all bool) error {
	allErrors := []error{}
	namespacedResources, nonNamespacedResources := h.customizer.GetResourceTypes()
	targetNamespace := h.customizer.GetTargetNamespace()
	err := h.pruneResources(namespacedResources, all, targetNamespace)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	err = h.pruneResources(nonNamespacedResources, all, "")
	if err != nil {
		allErrors = append(allErrors, err)
	}
	return utilerrors.NewAggregate(allErrors)
}

func (h *HelmReconciler) pruneResources(gvks []schema.GroupVersionKind, all bool, namespace string) error {
	allErrors := []error{}
	ownerLabels := h.customizer.GetOwnerLabels()
	ownerAnnotations := h.customizer.GetOwnerAnnotations()
	for _, gvk := range gvks {
		objects := &unstructured.UnstructuredList{}
		objects.SetGroupVersionKind(gvk)
		err := h.client.List(context.TODO(), client.MatchingLabels(ownerLabels).InNamespace(namespace), objects)
		if err != nil {
			h.logger.Error(err, "Error retrieving resources to prune", "type", gvk.String())
			allErrors = append(allErrors, err)
			continue
		}
	objectLoop:
		for _, object := range objects.Items {
			annotations := object.GetAnnotations()
			for ownerKey, ownerValue := range ownerAnnotations {
				// we only want to delete objects that contain the annotations
				// if we're not pruning all objects, we only want to prune those whose annotation value does not match what is expected
				if value, ok := annotations[ownerKey]; !ok || (!all && value == ownerValue) {
					continue objectLoop
				}
			}
			err = h.client.Delete(context.TODO(), &object, client.PropagationPolicy(metav1.DeletePropagationBackground))
			if err == nil {
				if listenerErr := h.customizer.ResourceDeleted(&object); listenerErr != nil {
					h.logger.Error(err, "error calling listener")
				}
			} else {
				if listenerErr := h.customizer.ResourceError(&object, err); listenerErr != nil {
					h.logger.Error(err, "error calling listener")
				}
				allErrors = append(allErrors, err)
			}
		}
	}
	return utilerrors.NewAggregate(allErrors)
}
