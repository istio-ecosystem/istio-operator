package common

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// IndexOf is a helper function returning the index of the specified string in the specified slice or -1 if not found.
func IndexOf(l []string, s string) int {
	for i, elem := range l {
		if elem == s {
			return i
		}
	}
	return -1
}

var accessor = meta.NewAccessor()

// HasLabel is a helper function returning true if the specified object contains the specified label.
func HasLabel(resource runtime.Object, label string) bool {
	labels, err := accessor.Labels(resource)
	if err != nil {
		// XXX: panic?
		return false
	}
	if labels == nil {
		return false
	}
	_, ok := labels[label]
	return ok
}

// DeleteLabel is a helper function which deletes the specified label from the specified object.
func DeleteLabel(resource runtime.Object, label string) {
	labels, err := accessor.Labels(resource)
	if err != nil {
		// XXX: panic?
		return
	}
	if labels == nil {
		return
	}
	delete(labels, label)
	_ = accessor.SetLabels(resource, labels)
}

// SetLabel is a helper function which sets the specified label and value on the specified object.
func SetLabel(resource runtime.Object, label, value string) {
	labels, err := accessor.Labels(resource)
	if err != nil {
		// XXX: panic?
		return
	}
	if labels == nil {
		labels = map[string]string{}
	}
	labels[label] = value
	_ = accessor.SetLabels(resource, labels)
}

// HasAnnotation is a helper function returning true if the specified object contains the specified annotation.
func HasAnnotation(resource runtime.Object, annotation string) bool {
	annotations, err := accessor.Annotations(resource)
	if err != nil {
		// XXX: panic?
		return false
	}
	if annotations == nil {
		return false
	}
	_, ok := annotations[annotation]
	return ok
}

// DeleteAnnotation is a helper function which deletes the specified annotation from the specified object.
func DeleteAnnotation(resource runtime.Object, annotation string) {
	annotations, err := accessor.Annotations(resource)
	if err != nil {
		// XXX: panic?
		return
	}
	if annotations == nil {
		return
	}
	delete(annotations, annotation)
	_ = accessor.SetAnnotations(resource, annotations)
}

// GetAnnotation is a helper function which returns the value of the specified annotation on the specified object.
// returns ok=false if the annotation was not found on the object.
func GetAnnotation(resource runtime.Object, annotation string) (value string, ok bool) {
	annotations, err := accessor.Annotations(resource)
	if err != nil {
		// XXX: panic?
		return
	}
	if annotations == nil {
		annotations = map[string]string{}
	}
	value, ok = annotations[annotation]
	return
}

// SetAnnotation is a helper function which sets the specified annotation and value on the specified object.
func SetAnnotation(resource runtime.Object, annotation, value string) {
	annotations, err := accessor.Annotations(resource)
	if err != nil {
		// XXX: panic?
		return
	}
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[annotation] = value
	_ = accessor.SetAnnotations(resource, annotations)
}
