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

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"istio.io/operator/pkg/util"

	"github.com/ghodss/yaml"

	"istio.io/operator/pkg/apis/istio/v1alpha2"

	"istio.io/operator/pkg/object"

	// For kubeclient GCP auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	"istio.io/operator/pkg/kubeclient"
	"istio.io/operator/pkg/name"
	"istio.io/operator/pkg/version"
	"istio.io/pkg/log"
)

const (
	// cRDPollInterval is how often the state of CRDs is polled when waiting for their creation.
	cRDPollInterval = 500 * time.Millisecond
	// cRDPollTimeout is the maximum wait time for all CRDs to be created.
	cRDPollTimeout = 60 * time.Second

	// operatorReconcileStr indicates that the operator will reconcile the resource.
	operatorReconcileStr = "Reconcile"
)

var (
	// operatorLabelStr indicates Istio operator is managing this resource.
	operatorLabelStr = name.OperatorAPINamespace + "/managed"
	// istioComponentLabelStr indicates which Istio component a resource belongs to.
	istioComponentLabelStr = name.OperatorAPINamespace + "/component"
	// istioVersionLabelStr indicates the Istio version of the installation.
	istioVersionLabelStr = name.OperatorAPINamespace + "/version"
)

// CompositeOutput is used to capture errors and stdout/stderr outputs for a command, per component.
type CompositeOutput struct {
	// Stdout is the stdout output.
	Stdout map[name.ComponentName]string
	// Stderr is the stderr output.
	Stderr map[name.ComponentName]string
	// Error is the error output.
	Err map[name.ComponentName]error
}

// NewCompositeOutput creates a new CompositeOutput and returns a ptr to it.
func NewCompositeOutput() *CompositeOutput {
	return &CompositeOutput{
		Stdout: make(map[name.ComponentName]string),
		Stderr: make(map[name.ComponentName]string),
		Err:    make(map[name.ComponentName]error),
	}
}

type componentNameToListMap map[name.ComponentName][]name.ComponentName
type componentTree map[name.ComponentName]interface{}

var (
	componentDependencies = componentNameToListMap{
		name.IstioBaseComponentName: {
			name.PilotComponentName,
			name.PolicyComponentName,
			name.TelemetryComponentName,
			name.GalleyComponentName,
			name.CitadelComponentName,
			name.NodeAgentComponentName,
			name.CertManagerComponentName,
			name.SidecarInjectorComponentName,
			name.IngressComponentName,
			name.EgressComponentName,
		},
	}

	installTree      = make(componentTree)
	dependencyWaitCh = make(map[name.ComponentName]chan struct{})

	kc            *kubeclient.Client
	k8sRESTConfig *rest.Config
)

func init() {
	buildInstallTree()
	for _, parent := range componentDependencies {
		for _, child := range parent {
			dependencyWaitCh[child] = make(chan struct{}, 1)
		}
	}

}

// ParseK8SYAMLToIstioControlPlaneSpec parses a YAML string IstioControlPlane CustomResource and unmarshals in into
// an IstioControlPlaneSpec object. It returns the object and an API group/version with it.
func ParseK8SYAMLToIstioControlPlaneSpec(yml string) (*v1alpha2.IstioControlPlaneSpec, *schema.GroupVersionKind, error) {
	o, err := object.ParseYAMLToK8sObject([]byte(yml))
	if err != nil {
		return nil, nil, err
	}
	y, err := yaml.Marshal(o.UnstructuredObject().Object["spec"])
	if err != nil {
		return nil, nil, err
	}
	icp := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(string(y), icp); err != nil {
		return nil, nil, err
	}
	gvk := o.GroupVersionKind()
	return icp, &gvk, nil
}

// RenderToDir writes manifests to a local filesystem directory tree.
func RenderToDir(manifests name.ManifestMap, outputDir string, dryRun, verbose bool) error {
	logAndPrint("Component dependencies tree: \n%s", installTreeString())
	logAndPrint("Rendering manifests to output dir %s", outputDir)
	return renderRecursive(manifests, installTree, outputDir, dryRun, verbose)
}

func renderRecursive(manifests name.ManifestMap, installTree componentTree, outputDir string, dryRun, verbose bool) error {
	for k, v := range installTree {
		componentName := string(k)
		ym := manifests[k]
		if ym == "" {
			logAndPrint("Manifest for %s not found, skip.", componentName)
			continue
		}
		logAndPrint("Rendering: %s", componentName)
		dirName := filepath.Join(outputDir, componentName)
		if !dryRun {
			if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
				return fmt.Errorf("could not create directory %s; %s", outputDir, err)
			}
		}
		fname := filepath.Join(dirName, componentName) + ".yaml"
		logAndPrint("Writing manifest to %s", fname)
		if !dryRun {
			if err := ioutil.WriteFile(fname, []byte(ym), 0644); err != nil {
				return fmt.Errorf("could not write manifest config; %s", err)
			}
		}

		kt, ok := v.(componentTree)
		if !ok {
			// Leaf
			return nil
		}
		if err := renderRecursive(manifests, kt, dirName, dryRun, verbose); err != nil {
			return err
		}
	}
	return nil
}

// ApplyAll applies all given manifests using kubectl client.
func ApplyAll(manifests name.ManifestMap, version version.Version, dryRun, verbose bool, kubeconfig, context string) (map[name.ComponentName]util.Errors, error) {
	var err error
	kc, err = kubeclient.NewClient(kubeconfig, context)
	if err != nil {
		return nil, err
	}
	logAndPrint("Applying manifests for these components:")
	for c := range manifests {
		logAndPrint("- %s", c)
	}
	logAndPrint("Component dependencies tree: \n%s", installTreeString())
	return applyRecursive(manifests, version, dryRun, verbose), nil
}

func applyRecursive(manifests name.ManifestMap, version version.Version, dryRun, verbose bool) map[name.ComponentName]util.Errors {
	var wg sync.WaitGroup
	out := make(map[name.ComponentName]util.Errors)
	for c, m := range manifests {
		c := c
		m := m
		wg.Add(1)
		go func() {
			if s := dependencyWaitCh[c]; s != nil {
				logAndPrint("%s is waiting on parent dependency...", c)
				<-s
				logAndPrint("Parent dependency for %s has unblocked, proceeding.", c)
			}
			out[c] = applyManifest(c, m, version, dryRun, verbose)

			// Signal all the components that depend on us.
			for _, ch := range componentDependencies[c] {
				logAndPrint("unblocking child dependency %s.", ch)
				dependencyWaitCh[ch] <- struct{}{}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	return out
}

func versionString(version version.Version) string {
	return version.String()
}

func applyManifest(componentName name.ComponentName, manifestStr string, version version.Version, dryRun, verbose bool) util.Errors {
	errsOut := util.Errors{}
	objects, err := object.ParseK8sObjectsFromYAMLManifest(manifestStr)
	if err != nil {
		return util.AppendErr(errsOut, err)
	}
	if len(objects) == 0 {
		return errsOut
	}

	namespace := ""
	for _, o := range objects {
		o.AddLabels(map[string]string{istioComponentLabelStr: string(componentName)})
		o.AddLabels(map[string]string{operatorLabelStr: operatorReconcileStr})
		o.AddLabels(map[string]string{istioVersionLabelStr: versionString(version)})
		if o.Namespace != "" {
			// All objects in a component have the same namespace.
			namespace = o.Namespace
		}
	}
	objects.Sort(defaultObjectOrder())

	var prune bool
	operatorSelector := map[string]string{
		operatorLabelStr: operatorReconcileStr,
	}

	logAndPrint("applying manifest for component %s", componentName)

	crdObjects, nonCrdObjects := filterCRDKindObjects(objects)
	if len(crdObjects) > 0 {
		for _, obj := range crdObjects {
			if err := kc.Apply(dryRun, verbose, prune, namespace, obj, operatorSelector); err != nil {
				errsOut = util.AppendErr(errsOut, err)
			}
		}

		if len(errsOut) > 0 {
			// Not all Istio components are robust to not yet created CRDs.
			if err := waitForCRDs(crdObjects, dryRun); err != nil {
				errsOut = util.AppendErr(errsOut, err)
			}
		}
	}

	log.Infof("applying the following manifest:\n%s", manifestStr)

	for _, obj := range nonCrdObjects {
		if err := kc.Apply(dryRun, verbose, prune, namespace, obj, operatorSelector); err != nil {
			errsOut = util.AppendErr(errsOut, err)
		}
	}

	logAndPrint("finished applying manifest for component %s", componentName)
	return errsOut
}

func defaultObjectOrder() func(o *object.K8sObject) int {
	return func(o *object.K8sObject) int {
		gk := o.Group + "/" + o.Kind
		switch gk {
		// Create CRDs asap - both because they are slow and because we will likely create instances of them soon
		case "apiextensions.k8s.io/CustomResourceDefinition":
			return -1000

			// We need to create ServiceAccounts, Roles before we bind them with a RoleBinding
		case "/ServiceAccount", "rbac.authorization.k8s.io/ClusterRole":
			return 1
		case "rbac.authorization.k8s.io/ClusterRoleBinding":
			return 2

			// Pods might need configmap or secrets - avoid backoff by creating them first
		case "/ConfigMap", "/Secrets":
			return 100

			// Create the pods after we've created other things they might be waiting for
		case "extensions/Deployment", "app/Deployment":
			return 1000

			// Autoscalers typically act on a deployment
		case "autoscaling/HorizontalPodAutoscaler":
			return 1001

			// Create services late - after pods have been started
		case "/Service":
			return 10000

		default:
			return 1000
		}
	}
}

// filterCRDKindObjects filter the CRD kind objects and others
func filterCRDKindObjects(objects object.K8sObjects) (crdObjects, nonCrdObjects object.K8sObjects) {
	for _, o := range objects {
		if o.Kind == "CustomResourceDefinition" {
			crdObjects = append(crdObjects, o)
		} else {
			nonCrdObjects = append(nonCrdObjects, o)
		}
	}
	return
}

func waitForCRDs(objects object.K8sObjects, dryRun bool) error {
	if dryRun {
		logAndPrint("Not waiting for CRDs in dry run mode.")
		return nil
	}

	logAndPrint("Waiting for CRDs to be applied.")
	cs, err := apiextensionsclient.NewForConfig(k8sRESTConfig)
	if err != nil {
		return fmt.Errorf("k8s client error: %s", err)
	}

	var crdNames []string
	for _, o := range objects {
		crdNames = append(crdNames, o.Name)
	}

	errPoll := wait.Poll(cRDPollInterval, cRDPollTimeout, func() (bool, error) {
	descriptor:
		for _, crdName := range crdNames {
			crd, errGet := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
			if errGet != nil {
				return false, errGet
			}
			for _, cond := range crd.Status.Conditions {
				switch cond.Type {
				case apiextensionsv1beta1.Established:
					if cond.Status == apiextensionsv1beta1.ConditionTrue {
						log.Infof("established CRD %q", crdName)
						continue descriptor
					}
				case apiextensionsv1beta1.NamesAccepted:
					if cond.Status == apiextensionsv1beta1.ConditionFalse {
						log.Warnf("name conflict: %v", cond.Reason)
					}
				}
			}
			log.Infof("missing status condition for %q", crdName)
			return false, nil
		}
		return true, nil
	})

	if errPoll != nil {
		logAndPrint("failed to verify CRD creation; %s", errPoll)
		return fmt.Errorf("failed to verify CRD creation: %s", errPoll)
	}

	logAndPrint("CRDs applied.")
	return nil
}

func buildInstallTree() {
	// Starting with root, recursively insert each first level child into each node.
	insertChildrenRecursive(name.IstioBaseComponentName, installTree, componentDependencies)
}

func insertChildrenRecursive(componentName name.ComponentName, tree componentTree, children componentNameToListMap) {
	tree[componentName] = make(componentTree)
	for _, child := range children[componentName] {
		insertChildrenRecursive(child, tree[componentName].(componentTree), children)
	}
}

func installTreeString() string {
	var sb strings.Builder
	buildInstallTreeString(name.IstioBaseComponentName, "", &sb)
	return sb.String()
}

func buildInstallTreeString(componentName name.ComponentName, prefix string, sb io.StringWriter) {
	_, _ = sb.WriteString(prefix + string(componentName) + "\n")
	if _, ok := installTree[componentName].(componentTree); !ok {
		return
	}
	for k := range installTree[componentName].(componentTree) {
		buildInstallTreeString(k, prefix+"  ", sb)
	}
}

func logAndPrint(v ...interface{}) {
	s := fmt.Sprintf(v[0].(string), v[1:]...)
	log.Infof(s)
	fmt.Println(s)
}
