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
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/component/controlplane"
	"istio.io/operator/pkg/helm"
	"istio.io/operator/pkg/name"
	"istio.io/operator/pkg/tpath"
	"istio.io/operator/pkg/translate"
	"istio.io/operator/pkg/util"
	"istio.io/operator/pkg/validate"
	"istio.io/operator/pkg/version"
	"istio.io/pkg/log"
)

func manifestCmd(args *rootArgs) *cobra.Command {
	mc := &cobra.Command{
		Use:   "manifest",
		Short: "Commands related to Istio manifests.",
		Long:  "The manifest subcommand is used to generate, apply, diff or migrate Istio manifests.",
	}

	mgcArgs := &manifestGenerateArgs{}
	mdcArgs := &manifestDiffArgs{}
	macArgs := &manifestApplyArgs{}

	mgc := manifestGenerateCmd(args, mgcArgs)
	mdc := manifestDiffCmd(args, mdcArgs)
	mac := manifestApplyCmd(args, macArgs)
	mmc := manifestMigrateCmd(args)

	addFlags(mgc, args)
	addFlags(mdc, args)
	addFlags(mac, args)

	addManifestGenerateFlags(mgc, mgcArgs)
	addManifestDiffFlags(mdc, mdcArgs)
	addManifestApplyFlags(mac, macArgs)

	mc.AddCommand(mgc)
	mc.AddCommand(mdc)
	mc.AddCommand(mac)
	mc.AddCommand(mmc)

	return mc
}

func genManifests(_ *rootArgs, inFilename string, setOverlayYAML string) (name.ManifestMap, error) {
	overlayYAML := ""
	if inFilename != "" {
		b, err := ioutil.ReadFile(inFilename)
		if err != nil {
			log.Fatalf("Could not open input file: %s", err)
		}
		overlayYAML = string(b)
	}

	overlayFilenameLog := inFilename
	if overlayFilenameLog == "" {
		overlayFilenameLog = "[Empty Filename]"
	}

	// Start with unmarshaling and validating the user CR (which is an overlay on the base profile).
	overlayICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(overlayYAML, overlayICPS); err != nil {
		return nil, fmt.Errorf("could not unmarshal the overlay YAML from file: %s", overlayFilenameLog)
	}
	if errs := validate.CheckIstioControlPlaneSpec(overlayICPS, false); len(errs) != 0 {
		return nil, fmt.Errorf("Overlay spec failed validation against IstioControlPlaneSpec: \n%v", overlayICPS)
	}

	baseProfileName := overlayICPS.Profile
	if baseProfileName == "" {
		baseProfileName = "[Builtin Profile]"
	}

	// Now read the base profile specified in the user spec. If nothing specified, use default.
	baseYAML, err := helm.ReadValuesYAML(overlayICPS.Profile)
	if err != nil {
		return nil, fmt.Errorf("error reading YAML from profile: %s", baseProfileName)
	}
	// Unmarshal and validate the base CR.
	baseICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(baseYAML, baseICPS); err != nil {
		return nil, fmt.Errorf("could not unmarshal the base YAML from profile: %s", baseProfileName)
	}
	if errs := validate.CheckIstioControlPlaneSpec(baseICPS, true); len(errs) != 0 {
		return nil, fmt.Errorf("base spec failed validation against IstioControlPlaneSpec: \n%v", baseICPS)
	}

	mergedYAML, err := helm.OverlayYAML(baseYAML, overlayYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to merge base YAML (%s) and overlay YAML (%s)", baseYAML, overlayYAML)
	}

	// Additional overlay from set flag.
	mergedYAML, err = helm.OverlayYAML(mergedYAML, setOverlayYAML)
	if err != nil {
		return nil, fmt.Errorf("failed to overlay YAML (%s) on merged YAML (%s)", setOverlayYAML, mergedYAML)
	}

	// Now unmarshal and validate the combined base profile and user CR overlay.
	mergedICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(mergedYAML, mergedICPS); err != nil {
		return nil, fmt.Errorf("could not unmarshal the merged YAML: \n%s", mergedYAML)
	}
	if errs := validate.CheckIstioControlPlaneSpec(mergedICPS, true); len(errs) != 0 {
		return nil, fmt.Errorf("merged spec failed validation against IstioControlPlaneSpec: \n%v", mergedICPS)
	}

	if yd := util.YAMLDiff(mergedYAML, util.ToYAMLWithJSONPB(mergedICPS)); yd != "" {
		return nil, fmt.Errorf("merged YAML differs from merged spec: \n%s", yd)
	}

	// TODO: remove version hard coding.
	cp := controlplane.NewIstioControlPlane(mergedICPS, translate.Translators[version.NewMinorVersion(1, 2)])
	if err := cp.Run(); err != nil {
		return nil, fmt.Errorf("failed to create Istio control plane with spec: \n%v", mergedICPS)
	}

	manifests, errs := cp.RenderManifest()
	if errs != nil {
		return manifests, errs.ToError()
	}
	return manifests, nil
}

// makeTreeFromSetList creates a YAML tree from a string slice containing key-value pairs in the format key=value.
func makeTreeFromSetList(setOverlay []string) (string, error) {
	if len(setOverlay) == 0 {
		return "", nil
	}
	tree := make(map[string]interface{})
	// Populate a default namespace for convenience, otherwise most --set commands will error out.
	if err := tpath.WriteNode(tree, util.PathFromString("defaultNamespacePrefix"), "istio-system"); err != nil {
		return "", err
	}
	for _, kv := range setOverlay {
		kvv := strings.Split(kv, "=")
		if len(kvv) != 2 {
			return "", fmt.Errorf("bad argument %s: expect format key=value", kv)
		}
		k := kvv[0]
		v := kvv[1]
		if err := tpath.WriteNode(tree, util.PathFromString(k), v); err != nil {
			return "", err
		}
		// To make errors more user friendly, test the path and error out immediately if we cannot unmarshal.
		testTree, err := yaml.Marshal(tree)
		if err != nil {
			return "", err
		}
		icps := &v1alpha2.IstioControlPlaneSpec{}
		if err := util.UnmarshalWithJSONPB(string(testTree), icps); err != nil {
			return "", fmt.Errorf("bad path=value: %s", kv)
		}
		if errs := validate.CheckIstioControlPlaneSpec(icps, true); len(errs) != 0 {
			return "", fmt.Errorf("bad path=value (%s): %s", kv, errs)
		}

	}
	out, err := yaml.Marshal(tree)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
