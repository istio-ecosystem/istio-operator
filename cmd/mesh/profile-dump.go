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
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/component/component"
	"istio.io/operator/pkg/helm"
	"istio.io/operator/pkg/tpath"
	"istio.io/operator/pkg/translate"
	"istio.io/operator/pkg/util"
	"istio.io/operator/pkg/validate"
	"istio.io/operator/pkg/version"
)

type profileDumpArgs struct {
	// inFilename is the path to the input IstioControlPlane CR.
	inFilename string
	// If set, display the translated Helm values rather than IstioControlPlaneSpec.
	helmValues bool
	// configPath sets the root node for the subtree to display the config for.
	configPath string
	// set is a string with element format "path=value" where path is an IstioControlPlane path and the value is a
	// value to set the node at that path to.
	set string
}

func addProfileDumpFlags(cmd *cobra.Command, args *profileDumpArgs) {
	cmd.PersistentFlags().StringVarP(&args.inFilename, "filename", "f", "", filenameFlagHelpStr)
	cmd.PersistentFlags().StringVarP(&args.configPath, "config-path", "p", "",
		"The path the root of the configuration subtree to dump e.g. trafficManagement.components.pilot. By default, dump whole tree. ")
	cmd.PersistentFlags().BoolVarP(&args.helmValues, "helm-values", "", false,
		"If set, dumps the Helm values that IstioControlPlaceSpec is translated to before manifests are rendered.")
	cmd.PersistentFlags().StringVarP(&args.set, "set", "s", "", setFlagHelpStr)
}

func profileDumpCmd(rootArgs *rootArgs, pdArgs *profileDumpArgs) *cobra.Command {
	return &cobra.Command{
		Use:   "dump",
		Short: "Dumps an Istio configuration profile.",
		Long:  "The dump subcommand is used to dump the values in an Istio configuration profile.",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			profileDump(rootArgs, pdArgs)
		}}

}

func profileDump(args *rootArgs, pdArgs *profileDumpArgs) {
	checkLogsOrExit(args)

	writer, err := getWriter("")
	if err != nil {
		logAndFatalf(args, err.Error())
	}
	defer func() {
		if err := writer.Close(); err != nil {
			logAndFatalf(args, "Did not close output successfully: %v", err)
		}
	}()

	mergedYAML := ""
	mergedcps := &v1alpha2.IstioControlPlaneSpec{}

	// TODO(mostrowski): load profile if set using --set flag.
	overlayYAML := ""
	if pdArgs.inFilename != "" {
		b, err := ioutil.ReadFile(pdArgs.inFilename)
		if err != nil {
			logAndFatalf(args, "Could not read values file %f: %s", pdArgs.inFilename, err)
		}
		overlayYAML = string(b)
	}

	// Start with unmarshaling and validating the user CR (which is an overlay on the base profile).
	overlayICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(overlayYAML, overlayICPS); err != nil {
		logAndFatalf(args, "Could not unmarshal the input file: %s\n\nOriginal YAML:\n%s\n", err, overlayYAML)
	}
	if errs := validate.CheckIstioControlPlaneSpec(overlayICPS, false); len(errs) != 0 {
		logAndFatalf(args, "Input file failed validation with the following errors: %s\n\nOriginal YAML:\n%s\n", errs, overlayYAML)
	}

	// Now read the base profile specified in the user spec.
	fname, err := helm.FilenameFromProfile(overlayICPS.Profile)
	if err != nil {
		logAndFatalf(args, "Could not get filename from profile: %s", err)
	}

	baseYAML, err := helm.ReadValuesYAML(overlayICPS.Profile)
	if err != nil {
		logAndFatalf(args, "Could not read the profile values for %s: %s", fname, err)
	}

	mergedYAML, err = helm.OverlayYAML(baseYAML, overlayYAML)
	if err != nil {
		logAndFatalf(args, "Could not overlay user config over base: %s", err)
	}
	// Now unmarshal and validate the combined base profile and user CR overlay.
	if err := util.UnmarshalWithJSONPB(mergedYAML, mergedcps); err != nil {
		logAndFatalf(args, err.Error())
	}
	if errs := validate.CheckIstioControlPlaneSpec(mergedcps, true); len(errs) != 0 {
		logAndFatalf(args, err.Error())
	}

	t, err := translate.NewTranslator(version.NewMinorVersion(1, 2))
	if err != nil {
		logAndFatalf(args, "%s", err)
	}

	if pdArgs.helmValues {
		mergedYAML, err = component.TranslateHelmValues(mergedcps, t, "")
		if err != nil {
			logAndFatalf(args, err.Error())
		}
	}

	finalYAML, err := getConfigSubtree(mergedYAML, pdArgs.configPath)
	if err != nil {
		logAndFatalf(args, "%s", err)
	}

	if _, err := writer.WriteString(finalYAML); err != nil {
		logAndFatalf(args, "Could not write values; %s", err)
	}
}

func getConfigSubtree(manifest, path string) (string, error) {
	root := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(manifest), &root); err != nil {
		return "", err
	}

	nc, _, err := tpath.GetPathContext(root, util.PathFromString(path))
	if err != nil {
		return "", err
	}
	out, err := yaml.Marshal(nc.Node)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
