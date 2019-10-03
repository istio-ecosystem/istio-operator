// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helmreconciler

import (
	"sync"

	"istio.io/operator/pkg/util"

	"istio.io/pkg/log"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"istio.io/operator/pkg/name"
)

// HelmReconciler reconciles resources rendered by a set of helm charts for a specific instances of a custom resource,
// or deletes all resources associated with a specific instance of a custom resource.
type HelmReconciler struct {
	client     client.Client
	logger     logr.Logger
	customizer RenderingCustomizer
	instance   runtime.Object
}

var _ LoggerProvider = &HelmReconciler{}
var _ ClientProvider = &HelmReconciler{}

// Factory is a factory for creating HelmReconciler objects using the specified CustomizerFactory.
type Factory struct {
	// CustomizerFactory is a factory for creating the Customizer object for the HelmReconciler.
	CustomizerFactory RenderingCustomizerFactory
}

// New Returns a new HelmReconciler for the custom resource.
// instance is the custom resource to be reconciled/deleted.
// client is the kubernetes client
// logger is the logger
func (f *Factory) New(instance runtime.Object, client client.Client, logger logr.Logger) (*HelmReconciler, error) {
	delegate, err := f.CustomizerFactory.NewCustomizer(instance)
	if err != nil {
		return nil, err
	}
	wrappedcustomizer, err := wrapCustomizer(instance, delegate)
	if err != nil {
		return nil, err
	}
	reconciler := &HelmReconciler{client: client, logger: logger, customizer: wrappedcustomizer, instance: instance}
	wrappedcustomizer.RegisterReconciler(reconciler)
	return reconciler, nil
}

// wrapCustomizer creates a new internalCustomizer object wrapping the delegate, by inject a LoggingRenderingListener,
// an OwnerReferenceDecorator, and a PruningDetailsDecorator into a CompositeRenderingListener that includes the listener
// from the delegate.  This ensures the HelmReconciler can properly implement pruning, etc.
// instance is the custom resource to be processed by the HelmReconciler
// delegate is the delegate
func wrapCustomizer(instance runtime.Object, delegate RenderingCustomizer) (*SimpleRenderingCustomizer, error) {
	ownerReferenceDecorator, err := NewOwnerReferenceDecorator(instance)
	if err != nil {
		return nil, err
	}
	return &SimpleRenderingCustomizer{
		InputValue:          delegate.Input(),
		PruningDetailsValue: delegate.PruningDetails(),
		ListenerValue: &CompositeRenderingListener{
			Listeners: []RenderingListener{
				&LoggingRenderingListener{Level: 1},
				ownerReferenceDecorator,
				NewPruningMarkingsDecorator(delegate.PruningDetails()),
				delegate.Listener(),
			},
		},
	}, nil
}

// Reconcile the resources associated with the custom resource instance.
func (h *HelmReconciler) Reconcile() error {
	// any processing required before processing the charts
	err := h.customizer.Listener().BeginReconcile(h.instance)
	if err != nil {
		return err
	}

	// render charts
	manifestMap, err := h.renderCharts(h.customizer.Input())
	if err != nil {
		return err
	}

	errs := h.processRecursive(manifestMap)

	// delete any obsolete resources
	errs = util.AppendErr(errs, h.customizer.Listener().BeginPrune(false))
	errs = util.AppendErr(errs, h.Prune(false))
	errs = util.AppendErr(errs, h.customizer.Listener().EndPrune())
	errs = util.AppendErr(errs, h.customizer.Listener().EndReconcile(h.instance, utilerrors.NewAggregate(errs)))

	return errs.ToError()
}

// processRecursive processes the given manifests in an order of dependencies defined in h. Dependencies are a tree,
// where a child must wait for the parent to complete before starting.
func (h *HelmReconciler) processRecursive(manifests ChartManifestsMap) util.Errors {
	deps, dch := h.customizer.Input().GetProcessingOrder(manifests)
	var errs []error

	var wg sync.WaitGroup
	for c, m := range manifests {
		c := c
		m := m
		wg.Add(1)
		go func() {
			if s := dch[name.ComponentName(c)]; s != nil {
				log.Infof("%s is waiting on parent dependency...", c)
				<-s
				log.Infof("Parent dependency for %s has unblocked, proceeding.", c)
			}

			if len(m) != 0 {
				errs = util.AppendErr(errs, h.ProcessManifest(m[0]))
			}

			// Signal all the components that depend on us.
			for _, ch := range deps[name.ComponentName(c)] {
				log.Infof("unblocking child dependency %s.", ch)
				dch[ch] <- struct{}{}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	return errs
}

// Delete resources associated with the custom resource instance
func (h *HelmReconciler) Delete() error {
	allErrors := []error{}

	// any processing required before processing the charts
	err := h.customizer.Listener().BeginDelete(h.instance)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	err = h.customizer.Listener().BeginPrune(true)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	err = h.Prune(true)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	err = h.customizer.Listener().EndPrune()
	if err != nil {
		allErrors = append(allErrors, err)
	}

	// any post processing required after deleting
	err = utilerrors.NewAggregate(allErrors)
	if listenerErr := h.customizer.Listener().EndDelete(h.instance, err); listenerErr != nil {
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

// GetClient returns the customizer associated with this HelmReconciler
func (h *HelmReconciler) GetCustomizer() RenderingCustomizer {
	return h.customizer
}

// GetClient returns the instance associated with this HelmReconciler
func (h *HelmReconciler) GetInstance() runtime.Object {
	return h.instance
}
