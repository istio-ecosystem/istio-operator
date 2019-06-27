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

package manifest

import "testing"

func TestHash(t *testing.T) {
	hashTests := []struct {
		kind      string
		namespace string
		name      string
		want      string
	}{
		{"aaa", "nnn", "mmm", "aaa:nnn:mmm"},
		{"a_", "n_", "m_", "a_:n_:m_"},
		{"a:a", "n:n", "m:m", "a:a:n:n:m:m"},
		{"", "", "", "::"},
	}

	for _, tt := range hashTests {
		got := Hash(tt.kind, tt.namespace, tt.name)
		if got != tt.want {
			t.Errorf("got %s for kind %s, namespace %s, name %s, want %s", got, tt.kind, tt.namespace, tt.name, tt.want)
		}
	}
}

func TestHashNameKind(t *testing.T) {
	hashNameKindTests := []struct {
		kind string
		name string
		want string
	}{
		{"aaa", "nnn", "aaa:nnn"},
		{"a_", "n_", "a_:n_"},
		{"a:a", "n:n", "a:a:n:n"},
		{"", "", ":"},
	}

	for _, tt := range hashNameKindTests {
		got := HashNameKind(tt.kind, tt.name)
		if got != tt.want {
			t.Errorf("got %s for kind %s, name %s, want %s", got, tt.kind, tt.name, tt.want)
		}
	}
}

func TestParseJSONToK8sObject(t *testing.T) {
	testDeployment := `{
	"apiVersion": "apps/v1",
	"kind": "Deployment",
	"metadata": {
		"name": "nginx-deployment",
		"namespace": "test-apps",
		"labels": {
			"app": "nginx"
		}
	},
	"spec": {
		"replicas": 3,
		"selector": {
			"matchLabels": {
				"app": "nginx"
			}
		},
		"template": {
			"metadata": {
				"labels": {
					"app": "nginx"
				}
			},
			"spec": {
				"containers": [
					{
						"name": "nginx",
						"image": "nginx:1.7.9",
						"ports": [
							{
								"containerPort": 80
							}
						]
					}
				]
			}
		}
	}
}`
	testPod := `{
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
		"name": "myapp-pod",
		"namespace": "test-apps",
		"labels": {
			"app": "myapp"
		}
	},
	"spec": {
		"containers": [
			{
				"name": "myapp-container",
				"image": "busybox",
				"command": [
					"sh",
					"-c",
					"echo Hello Kubernetes! && sleep 3600"
				]
			}
		]
	}
}`
	testService := `{
	"apiVersion": "v1",
	"kind": "Service",
	"metadata": {
		"name": "my-service",
		"namespace": "test-apps"
	},
	"spec": {
		"selector": {
			"app": "MyApp"
		},
		"ports": [
			{
				"protocol": "TCP",
				"port": 80,
				"targetPort": 9376
			}
		]
	}
}`

	parseJSONToK8sObjectTests := []struct {
		objString     string
		wantGroup     string
		wantKind      string
		wantName      string
		wantNamespace string
	}{
		{testDeployment, "apps", "Deployment", "nginx-deployment", "test-apps"},
		{testPod, "", "Pod", "myapp-pod", "test-apps"},
		{testService, "", "Service", "my-service", "test-apps"},
	}

	for _, tt := range parseJSONToK8sObjectTests {
		k8sObj, err := ParseJSONToK8sObject([]byte(tt.objString))
		if err != nil {
			k8sObjStr, err := k8sObj.YAMLDebugString()
			if err != nil {
				if k8sObj.Group != tt.wantGroup {
					t.Errorf("got group %s for k8s object %s, want %s", k8sObj.Group, k8sObjStr, tt.wantGroup)
				}
				if k8sObj.Group != tt.wantGroup {
					t.Errorf("got kind %s for k8s object %s, want %s", k8sObj.Kind, k8sObjStr, tt.wantKind)
				}
				if k8sObj.Name != tt.wantName {
					t.Errorf("got name %s for k8s object %s, want %s", k8sObj.Name, k8sObjStr, tt.wantName)
				}
				if k8sObj.Namespace != tt.wantNamespace {
					t.Errorf("got group %s for k8s object %s, want %s", k8sObj.Namespace, k8sObjStr, tt.wantNamespace)
				}
			}
		}
	}
}

func TestParseYAMLToK8sObject(t *testing.T) {
	testManifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: test-apps
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
---
kind: Pod
metadata:
  name: myapp-pod
  namespace: test-apps
  labels:
    app: myapp
spec:
  containers:
  - name: myapp-container
    image: busybox
    command: ['sh', '-c', 'echo Hello Kubernetes! && sleep 3600']
---
apiVersion: v1
kind: Service
metadata:
  name: my-service
  namespace: test-apps
spec:
  selector:
    app: MyApp
  ports:
  - protocol: TCP
    port: 80
    targetPort: 9376`

	parseK8sObjectsFromYAMLManifestTests := []struct {
		objString string
		hashes    []string
	}{
		{testManifest, []string{"Deployment:test-apps:nginx-deployment", "Pod:test-apps:myapp-pod", "Service:test-apps:my-service"}},
	}

	for _, tt := range parseK8sObjectsFromYAMLManifestTests {
		k8sObjs, err := ParseK8sObjectsFromYAMLManifest(tt.objString)
		if err != nil {
			for i, k8sObj := range k8sObjs {
				gotHash := k8sObj.Hash()
				k8sObjStr, err := k8sObj.YAMLDebugString()
				if err != nil {
					if k8sObj.Hash() != tt.hashes[i] {
						t.Errorf("got hash %s for k8s object %s, want %s", gotHash, k8sObjStr, tt.hashes[i])
					}
				}
			}
		}
	}
}
