package helmreconciler

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Factory is a factory for creating HelmReconciler objects using the specified CustomizerFactory.
type Factory struct {
	// CustomizerFactory is a factory for creating the Customizer object for the HelmReconciler.
	CustomizerFactory *CustomizerFactory
}

// New Returns a new HelmReconciler for the custom resource.
// instance is the custom resource to be reconciled/deleted.
// client is the kubernetes client
// logger is the logger
func (f *Factory) New(instance runtime.Object, client client.Client, logger logr.Logger) (*HelmReconciler, error) {
	customizer, err := f.CustomizerFactory.NewCustomizer(instance)
	if err != nil {
		return nil, err
	}
	return &HelmReconciler{client: client, logger: logger, customizer: customizer, instance: instance}, nil
}

// HelmReconciler reconciles resources rendered by a set of helm charts for a specific instances of a custom resource,
// or deletes all resources associated with a specific instance of a custom resource.
type HelmReconciler struct {
	client     client.Client
	logger     logr.Logger
	customizer *Customizer
	instance   runtime.Object
}

var _ LoggerProvider = &HelmReconciler{}
var _ ClientProvider = &HelmReconciler{}

// Reconcile the resources associated with the custom resource instance.
func (h *HelmReconciler) Reconcile() error {
	// any processing required before processing the charts
	err := h.customizer.BeginReconcile(h.instance)
	if err != nil {
		return err
	}

	// render charts
	manifestMap, err := h.renderCharts(h.customizer)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("error rendering charts"))
		listenerErr := h.customizer.EndReconcile(h.instance, err)
		if listenerErr != nil {
			h.logger.Error(listenerErr, "unexpected error invoking EndReconcile")
		}
		return err
	}

	// determine processing order
	chartOrder, err := h.customizer.GetProcessingOrder(manifestMap)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("error ordering charts"))
		listenerErr := h.customizer.EndReconcile(h.instance, err)
		if listenerErr != nil {
			h.logger.Error(listenerErr, "unexpected error invoking EndReconcile")
		}
		return err
	}

	// collect the errors.  from here on, we'll process everything with the assumption that any error is not fatal.
	allErrors := []error{}

	// process the charts
	for _, chartName := range chartOrder {
		chartManifests, ok := manifestMap[chartName]
		if !ok {
			// TODO: log warning about missing chart
			continue
		}
		chartManifests, err := h.customizer.BeginChart(chartName, chartManifests)
		if err != nil {
			allErrors = append(allErrors, err)
		}
		err = h.processManifests(chartManifests)
		if err != nil {
			allErrors = append(allErrors, err)
		}
		err = h.customizer.EndChart(chartName)
		if err != nil {
			allErrors = append(allErrors, err)
		}
	}

	// delete any obsolete resources
	err = h.customizer.BeginPrune(false)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	err = h.prune(false)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	err = h.customizer.EndPrune()
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// any post processing required after updating
	err = h.customizer.EndReconcile(h.instance, utilerrors.NewAggregate(allErrors))
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// return any errors
	return utilerrors.NewAggregate(allErrors)
}

// Delete resources associated with the custom resource instance
func (h *HelmReconciler) Delete() error {
	allErrors := []error{}

	// any processing required before processing the charts
	err := h.customizer.BeginDelete(h.instance)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	err = h.customizer.BeginPrune(true)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	err = h.prune(true)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	err = h.customizer.EndPrune()
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// any post processing required after deleting
	err = utilerrors.NewAggregate(allErrors)
	if listenerErr := h.customizer.EndDelete(h.instance, err); listenerErr != nil {
		h.logger.Error(listenerErr, "error calling listener")
	}

	// return any errors
	return err
}

// GetLogger returns the logger associated with this HelmReconciler
func (h *HelmReconciler) GetLogger() logr.Logger {
	return h.logger
}

// GetClient returns the kubernetes client associated with this HelmReconciler
func (h *HelmReconciler) GetClient() client.Client {
	return h.client
}
