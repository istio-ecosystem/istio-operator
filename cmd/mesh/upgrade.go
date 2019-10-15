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
	"time"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"istio.io/operator/pkg/manifest"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	meshInfoVersion "istio.io/pkg/version"

	goversion "github.com/hashicorp/go-version"

	"istio.io/operator/pkg/compare"
	pkgversion "istio.io/operator/pkg/version"
	opversion "istio.io/operator/version"
)

const (
	// The maximum duration the command will wait until the apply deployment reaches a ready state
	upgradeWaitSecWhenApply = 300
	// The time that the command will wait between each check of the upgraded version.
	upgradeWaitSecCheckVerPerLoop = 10
	// The maximum number of attempts that the command will check for the upgrade completion,
	// which means only the target version exist and the old version pods have been terminated.
	upgradeWaitCheckVerMaxAttempts = 60
)

var (
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
	// yes means skipping the prompting confirmation for value changes in this upgrade.
	yes bool
	// force means directly applying the upgrade without eligibility checks.
	force bool
	// versionsURI is a URI pointing to a YAML formatted versions mapping.
	versionsURI string
}

// addUpgradeFlags adds upgrade related flags into cobra command
func addUpgradeFlags(cmd *cobra.Command, args *upgradeArgs) {
	cmd.PersistentFlags().StringVarP(&args.inFilename, "filename",
		"f", "", "Path to file containing IstioControlPlane CustomResource")
	cmd.PersistentFlags().StringVarP(&args.kubeConfigPath, "kubeconfig",
		"c", "", "Path to kube config")
	cmd.PersistentFlags().StringVar(&args.context, "context", "",
		"The name of the kubeconfig context to use")
	cmd.PersistentFlags().BoolVarP(&args.yes, "yes", "y", false,
		"If yes, skips the prompting confirmation for value changes in this upgrade")
	cmd.PersistentFlags().BoolVarP(&args.wait, "wait", "w", false,
		"Wait, if set will wait until all Pods, Services, and minimum number of Pods "+
			"of a Deployment are in a ready state before the command exits. "+
			"It will wait for a maximum duration of "+strconv.Itoa(upgradeWaitSecWhenApply)+" seconds")
	cmd.PersistentFlags().BoolVar(&args.force, "force", false,
		"Apply the upgrade without eligibility checks and testing for changes "+
			"in profile default values")
	cmd.PersistentFlags().StringVarP(&args.versionsURI, "versionsURI", "u",
		versionsMapURL, "URI for operator versions to Istio versions map")
}

// Upgrade command upgrades Istio control plane in-place with eligibility checks
func UpgradeCmd() *cobra.Command {
	macArgs := &upgradeArgs{}
	rootArgs := &rootArgs{}
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade Istio control plane in-place",
		Long: "The mesh upgrade command checks for upgrade version eligibility and," +
			" if eligible, upgrades the Istio control plane components in-place. Warning: " +
			"traffic may be disrupted during upgrade. Please ensure PodDisruptionBudgets " +
			"are defined to maintain service continuity.",
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

// upgrade is the main function for Upgrade command
func upgrade(rootArgs *rootArgs, args *upgradeArgs) (err error) {
	initLogsOrExit(rootArgs)
	l.logAndPrintf("Client - istioctl version: %s\n", opversion.OperatorVersionString)

	targetValues := genValuesFromFile(args.inFilename, args.force)
	targetICPS := genICPSFromFile(args.inFilename, args.force)
	targetVersion := targetICPS.GetTag()
	l.logAndPrintf("Upgrade - target version: %s\n", targetVersion)

	kubeClient := getKubeExecClient(args.kubeConfigPath, args.context)
	//TODO(elfinhe): support components distributed in multiple namespaces
	istioNamespace := targetICPS.GetDefaultNamespace()
	currentVer := retrieveControlPlaneVersion(kubeClient, istioNamespace)
	currentValues := readValuesFromInjectorConfigMap(kubeClient, istioNamespace)

	if !args.force {
		checkSupportedVersions(currentVer, targetVersion, args.versionsURI)
		checkUpgradeValues(currentValues, targetValues)
		waitForConfirmation(args.yes)
	}

	runPreUpgradeHooks(kubeClient, istioNamespace,
		currentVer, targetVersion, currentValues, targetValues, rootArgs.dryRun)
	applyUpgradeManifest(args.inFilename, args.kubeConfigPath,
		args.context, rootArgs.dryRun, rootArgs.verbose)
	runPostUpgradeHooks(kubeClient, istioNamespace,
		currentVer, targetVersion, currentValues, targetValues, rootArgs.dryRun)

	if args.wait {
		waitUpgradeComplete(kubeClient, istioNamespace, targetVersion)
		upgradeVer := retrieveControlPlaneVersion(kubeClient, istioNamespace)
		l.logAndPrintf("Success. Now the Istio control plane is running at version %v.", upgradeVer)
		return
	}
	l.logAndPrintf("Upgrade submitted. Please use `istioctl version` to check the current versions.")
	return
}

// applyUpgradeManifest applies the Istion Control Plane specs reading from inFilename to
// the cluster by given kubeConfigPath and context
func applyUpgradeManifest(inFilename, kubeConfigPath, context string, dryRun, verbose bool) {
	genApplyManifests(nil, inFilename, dryRun,
		verbose, kubeConfigPath, context, upgradeWaitSecWhenApply, l)
}

// checkUpgradeValues checks the upgrade eligibility by comparing the current values with the target values
func checkUpgradeValues(curValues string, tarValues string) {
	diff := compare.YAMLCmp(curValues, tarValues)
	if diff == "" {
		l.logAndPrintf("Upgrade check: Values unchanged. The target values are identical to the current values.\n")
	} else {
		l.logAndPrintf("Upgrade check: Warning!!! the following values will be changed as part of upgrade. "+
			"If you have not overridden these values, they will change in your cluster. Please double check they are correct:\n%s", diff)
	}
}

// waitForConfirmation waits for user's confirmation if yes is not set
func waitForConfirmation(yes bool) {
	if yes {
		return
	}
	if !confirm("Confirm to proceed [y/N]?", os.Stdout) {
		l.logAndFatalf("Abort.")
	}
}

// genICPSFromFile generates an IstioControlPlaneSpec for a spec file
func genICPSFromFile(filename string, force bool) *v1alpha2.IstioControlPlaneSpec {
	_, overlayICPS, err := genICPS(filename, "", "", force, l)
	if err != nil {
		l.logAndFatalf("Failed to generate ICPS from file %s, error: %s",
			filename, err)
	}
	return overlayICPS
}

// genValuesFromFile generates values for a spec file
func genValuesFromFile(filename string, force bool) string {
	values, err := genProfile(true, filename, "", "", "", force, l)
	if err != nil {
		l.logAndFatalf("Abort. Failed to generate values from file: %v, error: %v", filename, err)
	}
	return values
}

// readValuesFromInjectorConfigMap reads the values from the config map of sidecar-injector.
func readValuesFromInjectorConfigMap(kubeClient manifest.ExecClient, istioNamespace string) string {
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

// checkSupportedVersions checks if the upgrade cur -> tar is supported by the tool
func checkSupportedVersions(cur, tar, versionsURI string) {
	if cur == tar {
		l.logAndFatalf("Abort. The current version %v equals to the target version %v.", cur, tar)
	}

	curPkgVer, err := pkgversion.NewVersionFromString(cur)
	if err != nil {
		l.logAndFatalf("Abort. Incorrect version: %v.", cur)
	}

	tarPkgVer, err := pkgversion.NewVersionFromString(tar)
	if err != nil {
		l.logAndFatalf("Abort. Incorrect version: %v.", tar)
	}

	if curPkgVer.Major != tarPkgVer.Major {
		l.logAndFatalf("Abort. Major version upgrade is not supported: %v -> %v.", cur, tar)
	}

	if curPkgVer.Minor != tarPkgVer.Minor {
		l.logAndFatalf("Abort. Minor version upgrade is not supported: %v -> %v.", cur, tar)
	}

	if curPkgVer.Patch == tarPkgVer.Patch {
		l.logAndFatalf("Abort. The target version has been installed in the cluster.\n"+
			"istioctl: %v\nIstio control plane: %v", cur, tar)
	}

	if curPkgVer.Patch > tarPkgVer.Patch {
		l.logAndFatalf("Abort. A newer version has been installed in the cluster.\n"+
			"istioctl: %v\nIstio control plane: %v", cur, tar)
	}

	tarGoVersion, err := goversion.NewVersion(tar)
	if err != nil {
		l.logAndFatalf("Abort. Failed to parse the target version: %v", tar)
	}

	compatibleMap := getVersionCompatibleMap(versionsURI, tarGoVersion, l)

	curGoVersion, err := goversion.NewVersion(cur)
	if err != nil {
		l.logAndFatalf("Abort. Failed to parse the current version: %v", cur)
	}

	if !compatibleMap.SupportedIstioVersions.Check(curGoVersion) {
		l.logAndFatalf("Abort. Upgrade is currently not supported: %v -> %v.", cur, tar)
	}

	l.logAndPrintf("Upgrade version check passed: %v -> %v.\n", cur, tar)
}

// retrieveControlPlaneVersion retrieves the version number from the Istio control plane
func retrieveControlPlaneVersion(kubeClient manifest.ExecClient, istioNamespace string) string {
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
	l.logAndPrint("")

	return coalesceVersions(meshInfo)
}

// waitUpgradeComplete waits for the upgrade to complete by periodically comparing the current component version
// to the target version.
func waitUpgradeComplete(kubeClient manifest.ExecClient, istioNamespace string, targetVer string) {
	for i := 1; i <= upgradeWaitCheckVerMaxAttempts; i++ {
		l.logAndPrintf("Waiting for upgrade rollout to complete, attempt #%v: ", i)
		sleepSeconds(upgradeWaitSecCheckVerPerLoop)
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

// sleepSeconds sleeps for n seconds, printing a dot '.' per second
func sleepSeconds(n int) {
	l.logAndPrintf("Going to sleep for %v secounds", n)
	for i := 1; i <= n; i++ {
		time.Sleep(time.Second)
		fmt.Print(".")
	}
	fmt.Println()
}

// coalesceVersions coalesces all Istio control plane components versions
func coalesceVersions(remoteVersion *meshInfoVersion.MeshInfo) string {
	if !identicalVersions(*remoteVersion) {
		l.logAndFatalf("Different versions of Istio components found: %v", remoteVersion)
	}
	return (*remoteVersion)[0].Info.Version
}

// identicalVersions checks if Istio control plane components are on the same version
func identicalVersions(remoteVersion meshInfoVersion.MeshInfo) bool {
	exemplar := remoteVersion[0].Info
	for i := 1; i < len(remoteVersion); i++ {
		candidate := (remoteVersion)[i].Info
		if exemplar.Version != candidate.Version {
			return false
		}
	}
	return true
}
