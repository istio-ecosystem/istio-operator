// Copyright 2018 Istio Authors
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

package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"istio.io/pkg/log"
	"istio.io/pkg/version"
)

// Client is a helper wrapper around the Kube RESTClient for istioctl -> Pilot/Envoy/Mesh related things
type Client struct {
	Config *rest.Config
	*rest.RESTClient
}

// ExecClient is an interface for remote execution
type ExecClient interface {
	GetIstioVersions(namespace string) (*version.MeshInfo, error)
	PodsForSelector(namespace, labelSelector string) (*v1.PodList, error)
	ConfigMapForSelector(namespace, labelSelector string) (*v1.ConfigMapList, error)
}

// NewClient is the constructor for the client wrapper
func NewClient(kubeconfig, configContext string) (*Client, error) {
	config, err := defaultRestConfig(kubeconfig, configContext)
	if err != nil {
		return nil, err
	}
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}
	return &Client{config, restClient}, nil
}

// PodExec takes a command and the pod data to run the command in the specified pod
func (client *Client) PodExec(podName, podNamespace, container string, command []string) (*bytes.Buffer, *bytes.Buffer, error) {
	req := client.Post().
		Resource("pods").
		Name(podName).
		Namespace(podNamespace).
		SubResource("exec").
		Param("container", container).
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(client.Config, "POST", req.URL())
	if err != nil {
		return nil, nil, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	return &stdout, &stderr, err
}

// ExtractExecResult wraps PodExec and return the execution result and error if has any.
func (client *Client) ExtractExecResult(podName, podNamespace, container string, cmd []string) ([]byte, error) {
	stdout, stderr, err := client.PodExec(podName, podNamespace, container, cmd)
	if err != nil {
		if stderr.String() != "" {
			return nil, fmt.Errorf("error execing into %v/%v %v container: %v\n%s", podName, podNamespace, container, err, stderr.String())
		}
		return nil, fmt.Errorf("error execing into %v/%v %v container: %v", podName, podNamespace, container, err)
	}
	return stdout.Bytes(), nil
}

// GetIstioPods retrieves the pod objects for Istio deployments
func (client *Client) GetIstioPods(namespace string, params map[string]string) ([]v1.Pod, error) {
	req := client.Get().
		Resource("pods").
		Namespace(namespace)
	for k, v := range params {
		req.Param(k, v)
	}

	res := req.Do()
	if res.Error() != nil {
		return nil, fmt.Errorf("unable to retrieve Pods: %v", res.Error())
	}
	list := &v1.PodList{}
	if err := res.Into(list); err != nil {
		return nil, fmt.Errorf("unable to parse PodList: %v", res.Error())
	}
	return list.Items, nil
}

type podDetail struct {
	binary    string
	container string
}

// GetIstioVersions gets the version for each Istio component
func (client *Client) GetIstioVersions(namespace string) (*version.MeshInfo, error) {
	pods, err := client.GetIstioPods(namespace, map[string]string{
		"labelSelector": "istio",
		"fieldSelector": "status.phase=Running",
	})
	if err != nil {
		log.Warnf("will use `--remote=false` to retrieve version info due to %q", err)
		return nil, nil
	}
	if len(pods) == 0 {
		log.Warnf("will use `--remote=false` to retrieve version info due to `no Istio pods in namespace %q`", namespace)
		return nil, nil
	}

	labelToPodDetail := map[string]podDetail{
		"pilot":            {"/usr/local/bin/pilot-discovery", "discovery"},
		"citadel":          {"/usr/local/bin/istio_ca", "citadel"},
		"egressgateway":    {"/usr/local/bin/pilot-agent", "istio-proxy"},
		"galley":           {"/usr/local/bin/galley", "galley"},
		"ingressgateway":   {"/usr/local/bin/pilot-agent", "istio-proxy"},
		"ilbgateway":       {"/usr/local/bin/pilot-agent", "istio-proxy"},
		"telemetry":        {"/usr/local/bin/mixs", "mixer"},
		"policy":           {"/usr/local/bin/mixs", "mixer"},
		"sidecar-injector": {"/usr/local/bin/sidecar-injector", "sidecar-injector-webhook"},
	}

	var errs error
	res := version.MeshInfo{}
	for _, pod := range pods {
		component := pod.Labels["istio"]

		// Special cases
		switch component {
		case "statsd-prom-bridge":
			continue
		case "mixer":
			component = pod.Labels["istio-mixer-type"]
		}

		server := version.ServerInfo{Component: component}

		if detail, ok := labelToPodDetail[component]; ok {
			cmd := []string{detail.binary, "version"}
			cmdJSON := append(cmd, "-o", "json")

			var info version.BuildInfo
			var v version.Version

			stdout, stderr, err := client.PodExec(pod.Name, pod.Namespace, detail.container, cmdJSON)

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error execing into %v %v container: %v", pod.Name, detail.container, err))
				continue
			}

			// At first try parsing stdout
			err = json.Unmarshal(stdout.Bytes(), &v)
			if err == nil && v.ClientVersion.Version != "" {
				info = *v.ClientVersion
			} else {
				// If stdout fails, try the old behavior
				if strings.HasPrefix(stderr.String(), "Error: unknown shorthand flag") {
					stdout, err := client.ExtractExecResult(pod.Name, pod.Namespace, detail.container, cmd)
					if err != nil {
						errs = multierror.Append(errs, fmt.Errorf("error execing into %v %v container: %v", pod.Name, detail.container, err))
						continue
					}

					info, err = version.NewBuildInfoFromOldString(string(stdout))
					if err != nil {
						errs = multierror.Append(errs, fmt.Errorf("error converting server info from JSON: %v", err))
						continue
					}
				} else {
					errs = multierror.Append(errs, fmt.Errorf("error execing into %v %v container: %v", pod.Name, detail.container, stderr.String()))
					continue
				}
			}

			server.Info = info
		}
		res = append(res, server)
	}
	return &res, errs
}

func (client *Client) PodsForSelector(namespace, labelSelector string) (*v1.PodList, error) {
	podGet := client.Get().Resource("pods").Namespace(namespace).Param("labelSelector", labelSelector)
	obj, err := podGet.Do().Get()
	if err != nil {
		return nil, fmt.Errorf("failed retrieving pod: %v", err)
	}
	return obj.(*v1.PodList), nil
}

func (client *Client) ConfigMapForSelector(namespace, labelSelector string) (*v1.ConfigMapList, error) {
	cmGet := client.Get().Resource("configmaps").Namespace(namespace).Param("labelSelector", labelSelector)
	obj, err := cmGet.Do().Get()
	if err != nil {
		return nil, fmt.Errorf("failed retrieving configmap: %v", err)
	}
	return obj.(*v1.ConfigMapList), nil
}
