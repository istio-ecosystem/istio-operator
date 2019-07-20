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

	"github.com/spf13/cobra"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/component/controlplane"
	"istio.io/operator/pkg/helm"
	"istio.io/operator/pkg/name"
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

func genManifests(_ *rootArgs, inFilename string) (name.ManifestMap, error) {
	overlayYAML := ""
	if inFilename != "" {
		b, err := ioutil.ReadFile(inFilename)
		if err != nil {
			log.Fatalf("Could not open input file: %s", err)
		}
		overlayYAML = string(b)
	}

	// Start with unmarshaling and validating the user CR (which is an overlay on the base profile).
	overlayICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(overlayYAML, overlayICPS); err != nil {
		return nil, err
	}
	if errs := validate.CheckIstioControlPlaneSpec(overlayICPS, false); len(errs) != 0 {
		return nil, errs.ToError()
	}

	// Now read the base profile specified in the user spec. If nothing specified, use default.
	baseYAML, err := helm.ReadValuesYAML(overlayICPS.Profile)
	if err != nil {
		return nil, err
	}
	// Unmarshal and validate the base CR.
	baseICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(baseYAML, baseICPS); err != nil {
		return nil, err
	}
	if errs := validate.CheckIstioControlPlaneSpec(baseICPS, true); len(errs) != 0 {
		return nil, err
	}

	mergedYAML, err := helm.OverlayYAML(baseYAML, overlayYAML)
	if err != nil {
		return nil, err
	}

	// Now unmarshal and validate the combined base profile and user CR overlay.
	mergedcps := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(mergedYAML, mergedcps); err != nil {
		return nil, err
	}
	if errs := validate.CheckIstioControlPlaneSpec(mergedcps, true); len(errs) != 0 {
		return nil, errs.ToError()
	}

	if yd := util.YAMLDiff(mergedYAML, util.ToYAMLWithJSONPB(mergedcps)); yd != "" {
		return nil, fmt.Errorf("validated YAML differs from input: \n%s", yd)
	}

	// TODO: remove version hard coding.
	cp := controlplane.NewIstioControlPlane(mergedcps, translate.Translators[version.NewMinorVersion(1, 2)])
	if err := cp.Run(); err != nil {
		return nil, err
	}

	manifests, errs := cp.RenderManifest()

	return manifests, errs.ToError()
}
