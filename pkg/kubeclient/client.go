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

// Apply runs the kubectl apply with the provided manifest argument
func (c *Client) Apply(dryRun, verbose bool, namespace string, manifest string) error {
	obj, err := object.ParseYAMLToK8sObject([]byte(manifest))
	if err != nil {
		logAndPrint("ParseYAMLToK8sObject from manifest string: %s with error: %v", manifest, err)
		return err
	}
	gvk := obj.GroupVersionKind()
	gk := obj.GroupKind()

	groupResources, err := restmapper.GetAPIGroupResources(c.kubeClient.Discovery())
	if err != nil {
		logAndPrint("error GetAPIGroupResources from kube-apiserver: %v", err)
		return err
	}

	rm := restmapper.NewDiscoveryRESTMapper(groupResources)
	mapping, err := rm.RESTMapping(gk, gvk.Version)

	if err != nil {
		logAndPrint("getting RESTMappings for the provided group version kind: %v with error: %v", gvk, err)
		return err
	}

	// TODO: check if the obj already exists
	_, err = c.dynamicInterface.Resource(mapping.Resource).Namespace(namespace).Create(obj.UnstructuredObject(), metav1.CreateOptions{})
	if err != nil {
		logAndPrint("creating the resource:\n%s\nwith error: %v", manifest, err)
		return err
	}
	return nil
}

func logAndPrint(v ...interface{}) {
	s := fmt.Sprintf(v[0].(string), v[1:]...)
	log.Infof(s)
	fmt.Println(s)
}
