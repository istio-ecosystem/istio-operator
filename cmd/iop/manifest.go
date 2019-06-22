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

package iop

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/ostromart/istio-installer/pkg/apis/istio/v1alpha2"
	"github.com/ostromart/istio-installer/pkg/component/controlplane"
	"github.com/ostromart/istio-installer/pkg/helm"
	"github.com/ostromart/istio-installer/pkg/translate"
	"github.com/ostromart/istio-installer/pkg/util"
	"github.com/ostromart/istio-installer/pkg/validate"
	"github.com/ostromart/istio-installer/pkg/version"
)

func manifestCmd(rootArgs *rootArgs, printf, fatalf FormatFn) *cobra.Command {
	return &cobra.Command{
		Use:   "manifest",
		Short: "Generates Istio install manifest.",
		Long:  "The manifest subcommand is used to generate an Istio install manifest based on the input CR.",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genManifest(rootArgs, printf, fatalf)
		}}

}

func genManifest(args *rootArgs, printf, fatalf FormatFn) {
	if args.crPath == "" {
		fatalf("Must set crpath")
	}
	b, err := ioutil.ReadFile(args.crPath)
	if err != nil {
		fatalf(err.Error())
	}
	overlayYAML := string(b)

	// Start with unmarshaling and validating the user CR (which is an overlay on the base profile).
	overlayICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(overlayYAML, overlayICPS); err != nil {
		fatalf(err.Error())
	}
	if errs := validate.CheckIstioControlPlaneSpec(overlayICPS); len(errs) != 0 {
		fatalf(errs.ToError().Error())
	}

	// Now read the base profile specified in the user spec.
	baseYAML, err := helm.ReadValuesYAML(overlayICPS.CustomPackagePath, overlayICPS.BaseProfilePath)
	if err != nil {
		fatalf(err.Error())
	}
	// Unmarshal and validate the base CR.
	baseICPS := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(baseYAML, baseICPS); err != nil {
		fatalf(err.Error())
	}
	if errs := validate.CheckIstioControlPlaneSpec(baseICPS); len(errs) != 0 {
		fatalf(errs.ToError().Error())
	}

	mergedYAML, err := helm.OverlayYAML(baseYAML, overlayYAML)
	if err != nil {
		fatalf(err.Error())
	}

	// Now unmarshal and validate the combined base profile and user CR overlay.
	mergedcps := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(mergedYAML, mergedcps); err != nil {
		fatalf(err.Error())
	}
	if errs := validate.CheckIstioControlPlaneSpec(mergedcps); len(errs) != 0 {
		fatalf(errs.ToError().Error())
	}

	if yd := util.YAMLDiff(mergedYAML, util.ToYAMLWithJSONPB(mergedcps)); yd != "" {
		fatalf("Validated YAML differs from input: \n%s", yd)
	}

	// TODO: remove version hard coding.
	cp := controlplane.NewIstioControlPlane(mergedcps, translate.Translators[version.MinorVersion{1, 2}])
	if err := cp.Run(); err != nil {
		fatalf(err.Error())
	}

	y, errs := cp.RenderManifest()
	err = errs.ToError()
	if err != nil {
		fatalf(err.Error())
	}
	fmt.Println(y)
}
