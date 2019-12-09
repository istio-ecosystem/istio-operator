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

package istiocontrolplane

import (
	"context"
	"fmt"

	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/helmreconciler"
	"istio.io/pkg/log"
)

const (
	finalizer = "istio-finalizer.install.istio.io"
	// finalizerMaxRetries defines the maximum number of attempts to add finalizers.
	finalizerMaxRetries = 10
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new IstioControlPlane Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	factory := &helmreconciler.Factory{CustomizerFactory: &IstioRenderingCustomizerFactory{}}
	return &ReconcileIstioControlPlane{client: mgr.GetClient(), scheme: mgr.GetScheme(), factory: factory}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	log.Info("Adding controller for IstioControlPlane")
	// Create a new controller
	c, err := controller.New("istiocontrolplane-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource IstioControlPlane
	err = c.Watch(&source.Kind{Type: &v1alpha2.IstioControlPlane{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	//watch for changes to Istio resources
	err = watchIstioResources(c)
	if err != nil {
		return err
	}
	log.Info("Controller added")
	return nil
}

var _ reconcile.Reconciler = &ReconcileIstioControlPlane{}

// ReconcileIstioControlPlane reconciles a IstioControlPlane object
type ReconcileIstioControlPlane struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client  client.Client
	scheme  *runtime.Scheme
	factory *helmreconciler.Factory
}

// Reconcile reads that state of the cluster for a IstioControlPlane object and makes changes based on the state read
// and what is in the IstioControlPlane.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIstioControlPlane) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconciling IstioControlPlane")

	ns := request.Namespace
	if ns == "" {
		ns = defaultNs
	} else {
		defaultNs = ns
	}
	reqNamespacedName := types.NamespacedName{
		Name:      request.Name,
		Namespace: ns,
	}
	// declare read-only icp instance to create the reconciler
	icp := &v1alpha2.IstioControlPlane{}
	if err := r.client.Get(context.TODO(), reqNamespacedName, icp); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Errorf("error getting IstioControlPlane icp: %s", err)
		return reconcile.Result{}, err
	}

	deleted := icp.GetDeletionTimestamp() != nil
	finalizers := sets.NewString(icp.GetFinalizers()...)
	if deleted {
		if !finalizers.Has(finalizer) {
			log.Info("IstioControlPlane deleted")
			return reconcile.Result{}, nil
		}
		log.Info("Deleting IstioControlPlane")

		reconciler, err := r.factory.New(icp, r.client)
		if err == nil {
			err = reconciler.Delete()
		} else {
			log.Errorf("failed to create reconciler: %s", err)
		}
		// TODO: for now, nuke the resources, regardless of errors
		finalizers.Delete(finalizer)
		icp.SetFinalizers(finalizers.List())
		finalizerError := r.client.Update(context.TODO(), icp)
		for retryCount := 0; errors.IsConflict(finalizerError) && retryCount < finalizerMaxRetries; retryCount++ {
			// workaround for https://github.com/kubernetes/kubernetes/issues/73098 for k8s < 1.14
			// TODO: make this error message more meaningful.
			log.Info("conflict during finalizer removal, retrying")
			_ = r.client.Get(context.TODO(), request.NamespacedName, icp)
			finalizers = sets.NewString(icp.GetFinalizers()...)
			finalizers.Delete(finalizer)
			icp.SetFinalizers(finalizers.List())
			finalizerError = r.client.Update(context.TODO(), icp)
		}
		if finalizerError != nil {
			log.Errorf("error removing finalizer: %s", finalizerError)
			return reconcile.Result{}, finalizerError
		}
		return reconcile.Result{}, err
	} else if !finalizers.Has(finalizer) {
		log.Infof("Adding finalizer %v to %v", finalizer, request)
		finalizers.Insert(finalizer)
		icp.SetFinalizers(finalizers.List())
		err := r.client.Update(context.TODO(), icp)
		if err != nil {
			log.Errorf("Failed to update IstioControlPlane with finalizer, %v", err)
			return reconcile.Result{}, err
		}
	}

	log.Info("Updating IstioControlPlane")
	reconciler, err := r.getOrCreateReconciler(icp)
	if err == nil {
		err = reconciler.Reconcile()
		if err != nil {
			log.Errorf("reconciling err: %s", err)
		}
	} else {
		log.Errorf("failed to create reconciler: %s", err)
	}

	return reconcile.Result{}, err
}

var (
	defaultNs   string
	reconcilers = map[string]*helmreconciler.HelmReconciler{}
)

func reconcilersMapKey(icp *v1alpha2.IstioControlPlane) string {
	return fmt.Sprintf("%s/%s", icp.Namespace, icp.Name)
}

var ownedResourcePredicates = predicate.Funcs{
	CreateFunc: func(_ event.CreateEvent) bool {
		// no action
		return false
	},
	GenericFunc: func(_ event.GenericEvent) bool {
		// no action
		return false
	},
	DeleteFunc: func(e event.DeleteEvent) bool {
		object, err := meta.Accessor(e.Object)
		log.Debugf("got delete event for %s.%s", object.GetName(), object.GetNamespace())
		if err != nil {
			return false
		}
		if object.GetLabels()[OwnerNameKey] != "" {
			return true
		}
		return false
	},
	UpdateFunc: func(e event.UpdateEvent) bool {
		object, err := meta.Accessor(e.ObjectNew)
		log.Debugf("got update event for %s.%s", object.GetName(), object.GetNamespace())
		if err != nil {
			return false
		}
		if object.GetLabels()[OwnerNameKey] != "" {
			return true
		}
		return false
	},
}

func (r *ReconcileIstioControlPlane) getOrCreateReconciler(icp *v1alpha2.IstioControlPlane) (*helmreconciler.HelmReconciler, error) {
	key := reconcilersMapKey(icp)
	var err error
	var reconciler *helmreconciler.HelmReconciler
	if reconciler, ok := reconcilers[key]; ok {
		reconciler.SetNeedUpdateAndPrune(false)
		oldInstance := reconciler.GetInstance()
		reconciler.SetInstance(icp)
		if reconciler.GetInstance().GetGeneration() != oldInstance.GetGeneration() {
			//regenerate the reconciler
			if reconciler, err = r.factory.New(icp, r.client); err == nil {
				reconcilers[key] = reconciler
			}
		}
		return reconciler, err
	}
	//not found - generate the reconciler
	if reconciler, err = r.factory.New(icp, r.client); err == nil {
		reconcilers[key] = reconciler
	}
	return reconciler, err
}

// Watch changes for Istio resources managed by the operator
func watchIstioResources(c controller.Controller) error {
	for _, t := range []runtime.Object{
		&corev1.ServiceAccount{TypeMeta: metav1.TypeMeta{Kind: "ServiceAccount", APIVersion: "v1"}},
		&rbacv1.Role{TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: "v1"}},
		&rbacv1.RoleBinding{TypeMeta: metav1.TypeMeta{Kind: "ClusterRoleBinding", APIVersion: "v1"}},
		&rbacv1.ClusterRole{TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: "v1"}},
		&rbacv1.ClusterRoleBinding{TypeMeta: metav1.TypeMeta{Kind: "ClusterRoleBinding", APIVersion: "v1"}},
		&corev1.ConfigMap{TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"}},
		&corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}},
		&appsv1.Deployment{TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "v1"}},
		&appsv1.DaemonSet{TypeMeta: metav1.TypeMeta{Kind: "DaemonSet", APIVersion: "v1"}},
		&autoscalingv2beta1.HorizontalPodAutoscaler{TypeMeta: metav1.TypeMeta{Kind: "HorizontalPodAutoscaler", APIVersion: "v2beta1"}},
		&admissionregistrationv1beta1.MutatingWebhookConfiguration{TypeMeta: metav1.TypeMeta{Kind: "MutatingWebhookConfiguration", APIVersion: "v1beta1"}},
		&v1beta1.Deployment{TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "v1beta1"}},
	} {
		err := c.Watch(&source.Kind{Type: t}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
				log.Debugf("watch a change for istio resource: %s.%s", a.Meta.GetName(), a.Meta.GetNamespace())
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Name: a.Meta.GetLabels()[OwnerNameKey],
					}},
				}
			}),
		}, ownedResourcePredicates)
		if err != nil {
			return err
		}
	}
	return nil
}
