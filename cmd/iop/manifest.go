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

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/component/component"
	"istio.io/operator/pkg/component/controlplane"
	"istio.io/operator/pkg/helm"
	"istio.io/operator/pkg/util"
	"istio.io/operator/pkg/validate"
	"github.com/spf13/cobra"
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
	overlaycps := &v1alpha2.IstioControlPlaneSpec{}
	if err := util.UnmarshalWithJSONPB(overlayYAML, overlaycps); err != nil {
		fatalf(err.Error())
	}
	if errs := validate.CheckIstioControlPlaneSpec(overlaycps); len(errs) != 0 {
		fatalf(errs.ToError().Error())
	}

	// Now read the base profile specified in the user spec.
	b, err = ioutil.ReadFile(util.GetLocalFilePath(overlaycps.BaseProfilePath))
	if err != nil {
		fmt.Printf("1")
		fatalf(err.Error())
	}
	baseYAML := string(b)

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

	cp := controlplane.NewIstioControlPlane(mergedcps, component.V12DirLayout)
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
