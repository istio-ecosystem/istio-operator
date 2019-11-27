// Copyright 2018 Istio Authors.
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

package mesh

import (
	"errors"
	"fmt"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	v1batch "k8s.io/api/batch/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes/scheme"

	"istio.io/istio/galley/pkg/config/meta/metadata"
	"istio.io/operator/pkg/translate"
)

type verifyInstallArgs struct {
	kubeConfigFlags *genericclioptions.ConfigFlags
	fileNameFlags   *genericclioptions.FileNameFlags
	istioNamespace  string
	set             []string
}

const (
	// K8sJobResourceType is the resource type of kubernetes Job.
	K8sJobResourceType = "Job"
)

var (
	verifyInstallCmd               *cobra.Command
	crdCount, istioDeploymentCount int
	compareFile                    string
	vArgs                          *verifyInstallArgs
	rArgs                          *rootArgs
	l                              *logger
)

// VerifyInstallCommand verifies Istio Installation Status
func VerifyInstallCommand() *cobra.Command {
	vArgs = &verifyInstallArgs{}
	rArgs = &rootArgs{}
	verifyInstallCmd = &cobra.Command{
		Use:   "verify-install",
		Short: "Verifies Istio Installation Status or performs pre-check for the cluster before Istio installation",
		Long: `
		verify-install verifies Istio installation status against the installation file or the IstioControlPlane 
		CustomResource you specified when you installed Istio. It loops through all the installation
		resources defined in your installation file and reports whether all of them are
		in ready status. It will report failure when any of them are not ready.

		If you do not specify installation file it will perform pre-check for your cluster
		and report whether the cluster is ready for Istio installation.
`,
		Example: `
		# Verify that Istio can be freshly installed
		istioctl verify-install
		
		# Verify the deployment matches a custom Istio deployment configuration
		istioctl verify-install -f $HOME/istio.yaml

		# Verify the deployment matches an IstioControlPlane CustomResource
		istioctl verify-install -f $HOME/istio_v1alpha2_istiocontrolplane_cr.yaml

		# Verify the deployment matches minimal profile for IstioControlPlane CustomResource
		istioctl verify-install -s profile=minimal

`,
		RunE: func(c *cobra.Command, args []string) error {
			l = newLogger(rArgs.logToStdErr, c.OutOrStderr(), c.OutOrStderr())
			return verifyInstall(vArgs, args, l)
		},
	}
	addRootFlags(verifyInstallCmd, rArgs)
	addVerifyInstallFlags(verifyInstallCmd, vArgs)

	return verifyInstallCmd
}

func addRootFlags(cmd *cobra.Command, rootArgs *rootArgs) {
	cmd.PersistentFlags().BoolVarP(&rootArgs.logToStdErr, "logtostderr", "",
		false, "Send logs to stderr.")
	cmd.PersistentFlags().BoolVarP(&rootArgs.verbose, "verbose", "",
		false, "Verbose output.")
}

func addVerifyInstallFlags(cmd *cobra.Command, v *verifyInstallArgs) {
	var (
		filenames     = []string{}
		fileNameFlags = &genericclioptions.FileNameFlags{
			Filenames: &filenames,
			Recursive: boolPtr(false),
			Usage:     "Istio YAML installation file or IstioControlPlane CustomResource file",
		}
		kubeConfigFlags = &genericclioptions.ConfigFlags{
			Context:    strPtr(""),
			Namespace:  strPtr(""),
			KubeConfig: strPtr(""),
		}
	)
	flags := cmd.PersistentFlags()
	flags.StringVarP(&v.istioNamespace, "istioNamespace", "i", "istio-system",
		"Istio system namespace")
	verifyInstallCmd.PersistentFlags().StringSliceVarP(&v.set, "set", "s", nil, setFlagHelpStr)
	kubeConfigFlags.AddFlags(flags)
	v.kubeConfigFlags = kubeConfigFlags
	fileNameFlags.AddFlags(flags)
	v.fileNameFlags = fileNameFlags
}

func verifyInstall(v *verifyInstallArgs, args []string, l *logger) error {
	options := v.fileNameFlags.ToOptions()
	if len(options.Filenames) == 0 && v.set == nil {
		if len(args) != 0 {
			l.logAndPrint(verifyInstallCmd.UsageString())
			return fmt.Errorf("verify-install takes no arguments to perform installation pre-check")
		}
		return installPreCheck(v, l)
	}
	return verifyPostInstall(v, l)

}

func verifyPostInstall(v *verifyInstallArgs, l *logger) error {

	var res *resource.Result
	overlayFromSet, err := makeTreeFromSetList(v.set, true, l)
	if err != nil {
		l.logAndFatal(err.Error())
	}
	fileName := ""
	compareFile = strings.Join(v.set, ",")
	options := v.fileNameFlags.ToOptions()
	if len(options.Filenames) != 0 {
		fileName = options.Filenames[0]
		compareFile = fileName
	}
	manifest, err := genManifests(fileName, overlayFromSet, true, l)
	manifestStr := ""
	if err == nil {
		for _, v := range manifest {
			manifestStr += v
		}
		res = resource.NewBuilder(v.kubeConfigFlags).
			Unstructured().
			Stream(strings.NewReader(manifestStr), "manifest").
			Flatten().
			Do()
		if err := res.Err(); err != nil {
			return err
		}
	} else {
		res = resource.NewBuilder(v.kubeConfigFlags).
			Unstructured().
			FilenameParam(false, &options).
			Flatten().
			Do()
		if err := res.Err(); err != nil {
			return err
		}
	}

	err = res.Visit(visit)
	if err != nil {
		return err
	}
	l.logAndPrintf("Checked %v crds\n", crdCount)
	l.logAndPrintf("Checked %v Istio Deployments\n", istioDeploymentCount)
	l.logAndPrintf("Istio is installed successfully\n")
	return nil
}

func visit(info *resource.Info, err error) error {
	if err != nil {
		return err
	}
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(info.Object)
	if err != nil {
		return err
	}
	un := &unstructured.Unstructured{Object: content}
	kind := un.GetKind()
	name := un.GetName()
	namespace := un.GetNamespace()
	kinds := findResourceInSpec(kind)
	if kinds == "" {
		kinds = strings.ToLower(kind) + "s"
	}
	if namespace == "" {
		namespace = "default"
	}
	switch kind {
	case translate.K8sDeploymentResourceType:
		deployment := &appsv1.Deployment{}
		err = info.Client.
			Get().
			Resource(kinds).
			Namespace(namespace).
			Name(name).
			VersionedParams(&meta_v1.GetOptions{}, scheme.ParameterCodec).
			Do().
			Into(deployment)
		if err != nil {
			return err
		}
		err = getDeploymentStatus(deployment, name, compareFile)
		if err != nil {
			return err
		}
		if namespace == vArgs.istioNamespace && strings.HasPrefix(name, "istio-") {
			istioDeploymentCount++
		}
	case K8sJobResourceType:
		job := &v1batch.Job{}
		err = info.Client.
			Get().
			Resource(kinds).
			Namespace(namespace).
			Name(name).
			VersionedParams(&meta_v1.GetOptions{}, scheme.ParameterCodec).
			Do().
			Into(job)
		if err != nil {
			return err
		}
		for _, c := range job.Status.Conditions {
			if c.Type == v1batch.JobFailed {
				msg := fmt.Sprintf("Istio installation failed, incomplete or"+
					" does not match \"%s\" - the required Job  %s failed", compareFile, name)
				return errors.New(msg)
			}
		}
	default:
		result := info.Client.
			Get().
			Resource(kinds).
			Name(name).
			Do()
		if result.Error() != nil {
			result = info.Client.
				Get().
				Resource(kinds).
				Namespace(namespace).
				Name(name).
				Do()
			if result.Error() != nil {
				msg := fmt.Sprintf("Istio installation failed, incomplete or"+
					" does not match \"%s\" - the required %s:%s is not ready due to: %v", compareFile, kind, name, result.Error())
				return errors.New(msg)
			}
		}
		if kind == "CustomResourceDefinition" {
			crdCount++
		}
	}
	if rArgs.verbose {
		l.logAndPrintf("%s: %s.%s checked successfully\n", kind, name, namespace)
	}
	return nil
}

func getDeploymentStatus(deployment *appsv1.Deployment, name, fileName string) error {
	cond := getDeploymentCondition(deployment.Status, appsv1.DeploymentProgressing)
	if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
		msg := fmt.Sprintf("Istio installation failed, incomplete or does not match \"%s\""+
			" - deployment %q exceeded its progress deadline", fileName, name)
		return errors.New(msg)
	}
	if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
		msg := fmt.Sprintf("Istio installation failed, incomplete or does not match \"%s\""+
			" - waiting for deployment %q rollout to finish: %d out of %d new replicas have been updated",
			fileName, name, deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas)
		return errors.New(msg)
	}
	if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
		msg := fmt.Sprintf("Istio installation failed, incomplete or does not match \"%s\""+
			" - waiting for deployment %q rollout to finish: %d old replicas are pending termination",
			fileName, name, deployment.Status.Replicas-deployment.Status.UpdatedReplicas)
		return errors.New(msg)
	}
	if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
		msg := fmt.Sprintf("Istio installation failed, incomplete or does not match \"%s\""+
			" - waiting for deployment %q rollout to finish: %d of %d updated replicas are available",
			fileName, name, deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas)
		return errors.New(msg)
	}
	return nil
}

func getDeploymentCondition(status appsv1.DeploymentStatus, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

func findResourceInSpec(kind string) string {
	for _, r := range metadata.MustGet().KubeSource().Resources() {
		if r.Kind == kind {
			return r.Plural
		}
	}
	return ""
}

func strPtr(val string) *string {
	return &val
}

func boolPtr(val bool) *bool {
	return &val
}
