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

package kubeclient

import (
	"fmt"

	"istio.io/istio/pkg/kube"
	"istio.io/operator/pkg/object"
	"istio.io/pkg/log"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

// Client provides an interface to kube-apiserver, it contains a kubernetes Clientset
// and a dynamic client that deals with all "runtime.Objects".
type Client struct {
	// kubeClient is a kubernetes Clientset.
	kubeClient *kubernetes.Clientset
	// dynamicInterface is a kube dynamic client that deals with all "runtime.Objects".
	dynamicInterface dynamic.Interface
}

// NewClient creates a client interface to kube-apiserver and returns a ptr to it.
func NewClient(kubeconfig, context string) (*Client, error) {
	kubeClient, err := CreateKubeInterface(kubeconfig, context)
	if err != nil {
		return nil, err
	}

	dynamicInterface, err := CreateDynamicInterface(kubeconfig, context)
	if err != nil {
		return nil, err
	}

	return &Client{
		kubeClient:       kubeClient,
		dynamicInterface: dynamicInterface,
	}, nil
}

func CreateDynamicInterface(kubeconfig, context string) (dynamic.Interface, error) {
	restConfig, err := kube.BuildClientConfig(kubeconfig, context)

	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(restConfig)
}

func CreateKubeInterface(kubeconfig, context string) (*kubernetes.Clientset, error) {
	restConfig, err := kube.BuildClientConfig(kubeconfig, context)

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(restConfig)
}

// Apply function the k8s object into kube-apiserver with kubeclient.
func (c *Client) Apply(dryRun, verbose, prune bool, namespace string, obj *object.K8sObject, selector map[string]string) error {
	gvk := obj.GroupVersionKind()
	gk := obj.GroupKind()

	groupResources, err := restmapper.GetAPIGroupResources(c.kubeClient.Discovery())
	if err != nil {
		logAndPrint("getting API groupResources from kube-apiserver failed with error %v", err)
		return err
	}

	rm := restmapper.NewDiscoveryRESTMapper(groupResources)
	mapping, err := rm.RESTMapping(gk, gvk.Version)

	if err != nil {
		logAndPrint("getting REST mappings for the provided group version kind %v failed with error %v", gvk, err)
		return err
	}

	createOpts := metav1.CreateOptions{}
	deleteOpts := &metav1.DeleteOptions{}
	if dryRun {
		createOpts.DryRun = []string{metav1.DryRunAll}
		deleteOpts.DryRun = []string{metav1.DryRunAll}
	}
	resInterface := c.dynamicInterface.Resource(mapping.Resource).Namespace(namespace)

	objName := obj.UnstructuredObject().GetName()
	existingObj, err := resInterface.Get(objName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logAndPrint("creating the resource %s", obj.HashNameKind())
			_, err := resInterface.Create(obj.UnstructuredObject(), createOpts)
			if err != nil {
				logAndPrint("creating the resource %s failed with error %v", obj.HashNameKind(), err)
				return err
			}
			return nil
		}
	}

	labels := existingObj.GetLabels()
	if containsSelector(labels, selector) {
		if prune {
			// if there already exists resource with the selector, then delete the existing resource before creating new one.
			logAndPrint("pruning the existing resource %s", obj.HashNameKind())
			err := resInterface.Delete(objName, deleteOpts)
			if err != nil {
				return fmt.Errorf("pruning existing resource failed with error %v", err)
			}
			logAndPrint("creating the resource %s", obj.HashNameKind())
			_, err = resInterface.Create(obj.UnstructuredObject(), createOpts)
			if err != nil {
				logAndPrint("creating the resource %s failed with error %v", obj.HashNameKind(), err)
				return err
			}
			return nil
		}
		return fmt.Errorf("there already exists resource %s", obj.HashNameKind())
	}
	return fmt.Errorf("there already exists resource %s", obj.HashNameKind())
}

// containsSelector check if the labels contains the specified selector.
func containsSelector(labels, selector map[string]string) bool {
	if labels == nil {
		return selector == nil
	}
	if selector == nil {
		return true
	}
	for k, v := range selector {
		if val, ok := labels[k]; !ok || val != v {
			return false
		}
	}
	return true
}

func logAndPrint(v ...interface{}) {
	s := fmt.Sprintf(v[0].(string), v[1:]...)
	log.Infof(s)
	fmt.Println(s)
}
