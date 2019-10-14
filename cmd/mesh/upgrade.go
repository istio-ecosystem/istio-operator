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

package mesh

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	istioVersion "istio.io/pkg/version"

	"istio.io/operator/pkg/compare"
	"istio.io/operator/pkg/kubernetes"
	"istio.io/operator/pkg/manifest"
	opversion "istio.io/operator/version"
)

var (
	supportedVersionMap = map[string]map[string]bool{
		"1.3.0": {
			"1.3.1": true,
			"1.3.2": true,
		},
		"1.3.1": {
			"1.3.2": true,
		},
	}
	l *logger
)

type upgradeArgs struct {
	// inFilename is the path to the input IstioControlPlane CR.
	inFilename string
	// kubeConfigPath is the path to kube config file.
	kubeConfigPath string
	// context is the cluster context in the kube config.
	context string
	// wait is flag that indicates whether to wait resources ready before exiting.
	wait bool
	// yes means don't ask for confirmation (asking for confirmation not implemented).
	yes bool
	// force means directly applying the upgrade without eligibility checks.
	force bool
}

func addUpgradeFlags(cmd *cobra.Command, args *upgradeArgs) {
	cmd.PersistentFlags().StringVarP(&args.inFilename, "filename",
		"f", "", "Path to file containing IstioControlPlane CustomResource")
	cmd.PersistentFlags().StringVarP(&args.kubeConfigPath, "kubeconfig",
		"c", "", "Path to kube config")
	cmd.PersistentFlags().StringVar(&args.context, "context", "",
		"The name of the kubeconfig context to use")
	cmd.PersistentFlags().BoolVarP(&args.yes, "yes", "y", false,
		"Do not ask for confirmation")
	cmd.PersistentFlags().BoolVarP(&args.wait, "wait", "w", false,
		"Wait, if set will wait until all Pods, Services, and minimum number of Pods "+
			"of a Deployment are in a ready state before the command exits. "+
			"It will wait for a maximum duration of --readiness-timeout seconds")
	cmd.PersistentFlags().BoolVar(&args.force, "force", false,
		"Apply the upgrade without eligibility checks")
}

// Upgrade command upgrades Istio control plane in-place with eligibility checks
func Upgrade() *cobra.Command {
	macArgs := &upgradeArgs{}
	rootArgs := &rootArgs{}
	cmd := &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrade Istio control plane in-place",
		Example: `mesh upgrade`,
		RunE: func(cmd *cobra.Command, args []string) (e error) {
			l = newLogger(rootArgs.logToStdErr, cmd.OutOrStdout(), cmd.OutOrStderr())
			return upgrade(rootArgs, macArgs)
		},
	}
	addFlags(cmd, rootArgs)
	addUpgradeFlags(cmd, macArgs)
	return cmd
}

func upgrade(rootArgs *rootArgs, args *upgradeArgs) (err error) {
	initLogsOrExit(rootArgs)
	targetVer := retrieveClientVersion()
	targetValues := genValuesFromFile(targetVer, args.inFilename)
	overlayICPS := genICPSFromFile(args.inFilename)
	istioNamespace := overlayICPS.GetDefaultNamespace()

	kubeClient := getKubeClient(args.kubeConfigPath, args.context)
	currentVer := retrieveControlPlaneVersion(kubeClient, istioNamespace)
	currentValues := readValuesFromInjectorConfigMap(kubeClient, istioNamespace)

	if !args.force {
		checkSupportedVersions(currentVer, targetVer)
		checkUpgradeValues(currentValues, targetValues, args.yes)
	}

	runUpgradeHooks(kubeClient, istioNamespace,
		currentVer, targetVer, currentValues, targetValues)
	applyUpgradeManifest(targetVer, args.inFilename, args.kubeConfigPath,
		args.context, rootArgs.dryRun, rootArgs.verbose)

	if args.wait {
		waitUpgradeComplete(kubeClient, istioNamespace, targetVer)
		upgradeVer := retrieveControlPlaneVersion(kubeClient, istioNamespace)
		l.logAndPrintf("Success. Now the Istio control plane is running at version %v.", upgradeVer)
		return
	}
	l.logAndPrintf("Upgrade submitted. Please use `istioctl version` to check the current versions.")
	return
}

func applyUpgradeManifest(targetVer, inFilename,
	kubeConfigPath, context string, dryRun, verbose bool) {
	imageSourceOverlay := getImageSourceOverlay(targetVer)

	manifests, err := genManifests(inFilename, imageSourceOverlay)
	if err != nil {
		l.logAndFatalf("Failed to generate manifest: %v", err)
	}
	opts := &manifest.InstallOptions{
		DryRun:      dryRun,
		Verbose:     verbose,
		WaitTimeout: 300 * time.Second,
		Kubeconfig:  kubeConfigPath,
		Context:     context,
	}
	out, err := manifest.ApplyAll(manifests, opversion.OperatorBinaryVersion, opts)
	if err != nil {
		l.logAndFatalf("Failed to apply manifest with kubectl client: %v", err)
	}
	for cn := range manifests {
		if out[cn].Err != nil {
			cs := fmt.Sprintf("Component %s failed install:", cn)
			l.logAndPrintf("\n%s\n%s\n", cs, strings.Repeat("=", len(cs)))
			l.logAndPrint("Error: ", out[cn].Err, "\n")
		} else {
			cs := fmt.Sprintf("Component %s installed successfully:", cn)
			l.logAndPrintf("\n%s\n%s\n", cs, strings.Repeat("=", len(cs)))
		}

		if strings.TrimSpace(out[cn].Stderr) != "" {
			l.logAndPrint("Error detail:\n", out[cn].Stderr, "\n")
		}
		if strings.TrimSpace(out[cn].Stdout) != "" {
			l.logAndPrint("Stdout:\n", out[cn].Stdout, "\n")
		}
	}
}

func getKubeClient(kubeConfig, configContext string) kubernetes.ExecClient {
	kubeClient, err := clientExecFactory(kubeConfig, configContext)
	if err != nil {
		l.logAndFatalf("Abort. Failed to connect Kubernetes API server: %v", err)
	}
	return kubeClient
}

func checkUpgradeValues(curValues string, tarValues string, yes bool) {
	diff := compare.YAMLCmp(curValues, tarValues)
	if diff == "" {
		l.logAndPrintf("Upgrade check: values are valid for upgrade.\n")
	} else {
		l.logAndPrintf("Upgrade check: values will be changed during the upgrade:\n%s", diff)
	}

	if yes {
		return
	}
	if !confirm("Confirm to proceed [y/N]?", os.Stdout) {
		l.logAndFatalf("Abort.")
	}
}

func getImageSourceOverlay(v string) string {
	setImageSource := []string{
		"hub=docker.io/istio",
		"tag=" + v,
	}
	imageSourceOverlay, err := makeTreeFromSetList(setImageSource)
	if err != nil {
		l.logAndFatal(err.Error())
	}
	return imageSourceOverlay
}

func genICPSFromFile(filename string) *v1alpha2.IstioControlPlaneSpec {
	_, overlayICPS, err := genICPS(filename, "", "")
	if err != nil {
		l.logAndFatalf("Failed to generate ICPS from file %s, error: %s",
			filename, err)
	}
	return overlayICPS
}

func genValuesFromFile(targetVer, filename string) string {
	imageSourceOverlay := getImageSourceOverlay(targetVer)
	values, err := genProfile(true, filename, "", imageSourceOverlay, "")
	if err != nil {
		l.logAndFatalf("Abort. Failed to generate values from file: %v, error: %v", filename, err)
	}
	return values
}

func clientExecFactory(kubeconfig, configContext string) (kubernetes.ExecClient, error) {
	return kubernetes.NewClient(kubeconfig, configContext)
}

func readValuesFromInjectorConfigMap(kubeClient kubernetes.ExecClient, istioNamespace string) string {
	configMapList, err := kubeClient.ConfigMapForSelector(istioNamespace, "istio=sidecar-injector")
	if err != nil || len(configMapList.Items) == 0 {
		l.logAndFatalf("Abort. Failed to retrieve sidecar-injector config map: %v", err)
	}

	jsonValues := ""
	foundValues := false
	for _, item := range configMapList.Items {
		if item.Name == "istio-sidecar-injector" && item.Data != nil {
			jsonValues, foundValues = item.Data["values"]
			if foundValues {
				break
			}
		}
	}

	if !foundValues {
		l.logAndFatalf("Abort. Failed to find values in sidecar-injector config map: %v", configMapList)
	}

	yamlValues, err := yaml.JSONToYAML([]byte(jsonValues))
	if err != nil {
		l.logAndFatalf("jsonToYAML failed to parse values:\n%v\nError:\n%v", yamlValues, err)
	}

	return string(yamlValues)
}

func checkSupportedVersions(cur string, tar string) {
	if cur == tar {
		l.logAndFatalf("Abort. The current version %v equals to the target version %v.", cur, tar)
	}

	curMajor, curMinor, curPatch := parseVersionFormat(cur)
	tarMajor, tarMinor, tarPatch := parseVersionFormat(tar)

	if curMajor != tarMajor {
		l.logAndFatalf("Abort. Major version upgrade is not supported: %v -> %v.", cur, tar)
	}

	if curMinor != tarMinor {
		l.logAndFatalf("Abort. Minor version upgrade is not supported: %v -> %v.", cur, tar)
	}

	if curPatch == tarPatch {
		l.logAndFatalf("Abort. The target version has been installed in the cluster.\n"+
			"istioctl: %v\nIstio control plane: %v", cur, tar)
	}

	if curPatch > tarPatch {
		l.logAndFatalf("Abort. A newer version has been installed in the cluster.\n"+
			"istioctl: %v\nIstio control plane: %v", cur, tar)
	}

	if !supportedVersionMap[cur][tar] {
		l.logAndFatalf("Abort. Upgrade is currently not supported: %v -> %v.", cur, tar)
	}

	l.logAndPrintf("Version check passed: %v -> %v.\n", cur, tar)
}

func parseVersionFormat(ver string) (int, int, int) {
	fullVerArray := strings.Split(ver, "-")
	if len(fullVerArray) == 0 {
		l.logAndFatalf("Abort. Incorrect version: %v.", ver)
	}
	verArray := strings.Split(fullVerArray[0], ".")
	if len(verArray) != 3 {
		l.logAndFatalf("Abort. Incorrect version: %v.", ver)
	}
	major, err := strconv.Atoi(verArray[0])
	if err != nil {
		l.logAndFatalf("Abort. Incorrect marjor version: %v.", verArray[0])
	}
	minor, err := strconv.Atoi(verArray[1])
	if err != nil {
		l.logAndFatalf("Abort. Incorrect minor version: %v.", verArray[1])
	}
	patch, err := strconv.Atoi(verArray[2])
	if err != nil {
		l.logAndFatalf("Abort. Incorrect patch version: %v.", verArray[2])
	}
	return major, minor, patch
}

func retrieveControlPlaneVersion(kubeClient kubernetes.ExecClient, istioNamespace string) string {
	meshInfo, e := kubeClient.GetIstioVersions(istioNamespace)
	if e != nil {
		l.logAndFatalf("Failed to retrieve Istio control plane version, error: %v", e)
	}

	if meshInfo == nil {
		l.logAndFatalf("Istio control plane not found in namespace: %v", istioNamespace)
	}

	for _, remote := range *meshInfo {
		l.logAndPrintf("Control Plane - %s pod - version: %s", remote.Component, remote.Info.Version)
	}

	return coalesceVersions(meshInfo)
}

func waitUpgradeComplete(kubeClient kubernetes.ExecClient, istioNamespace string, targetVer string) {
	for i := 1; i <= 60; i++ {
		l.logAndPrintf("Waiting for upgrade rollout to complete, attempt #%v: ", i)
		sleepSeconds(10)
		meshInfo, e := kubeClient.GetIstioVersions(istioNamespace)
		if e != nil {
			l.logAndPrintf("Failed to retrieve Istio control plane version, error: %v", e)
			continue
		}
		if meshInfo == nil {
			l.logAndPrintf("Failed to find Istio namespace: %v", istioNamespace)
			continue
		}
		if identicalVersions(*meshInfo) && targetVer == (*meshInfo)[0].Info.Version {
			l.logAndPrintf("Upgrade rollout completed. " +
				"All Istio control plane pods are running on the target version.\n\n")
			return
		}
		for _, remote := range *meshInfo {
			if targetVer != remote.Info.Version {
				l.logAndPrintf("Control Plane - %s pod - version %s does not match the target version %s",
					remote.Component, remote.Info.Version, targetVer)
			}
		}
	}
	l.logAndFatal("Upgrade rollout unfinished. Maximum number of attempts exceeded, quit...")
}

func sleepSeconds(n int) {
	l.logAndPrintf("Going to sleep for %v secounds", n)
	for i := 1; i <= n; i++ {
		time.Sleep(time.Second)
		fmt.Print(".")
	}
	fmt.Println()
}

func retrieveClientVersion() string {
	l.logAndPrintf("Client - istioctl version: %s\n", istioVersion.Info.Version)
	return istioVersion.Info.Version
}

func coalesceVersions(remoteVersion *istioVersion.MeshInfo) string {
	if !identicalVersions(*remoteVersion) {
		l.logAndFatalf("Different versions of Istio componets found: %v", remoteVersion)
	}
	return (*remoteVersion)[0].Info.Version
}

func identicalVersions(remoteVersion istioVersion.MeshInfo) bool {
	exemplar := remoteVersion[0].Info
	for i := 1; i < len(remoteVersion); i++ {
		candidate := (remoteVersion)[i].Info
		if exemplar.Version != candidate.Version {
			return false
		}
	}

	return true
}
