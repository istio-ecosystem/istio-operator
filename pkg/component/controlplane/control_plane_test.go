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

package controlplane

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/name"
	"istio.io/operator/pkg/object"
	"istio.io/operator/pkg/translate"
	"istio.io/operator/pkg/util"
	"istio.io/operator/pkg/version"
)

var (
	repoRootDir string
	testDataDir string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	repoRootDir = filepath.Join(wd, "../../..")
	testDataDir = filepath.Join(wd, "testdata")

}

func TestRenderInstallationSuccessV13(t *testing.T) {
	tests := []struct {
		desc        string
		installSpec string
	}{
		{
			desc: "all_off",
			installSpec: `

`,
		},
		{
			desc: "pilot_default",
			installSpec: `


`,
		},
		{
			desc: "pilot_k8s_settings",
			installSpec: `

`,
		},
		{
			desc: "pilot_override_values",
			installSpec: `

`,
		},
		{
			desc: "pilot_override_kubernetes",
			installSpec: `

`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var is v1alpha2.IstioControlPlaneSpec

			err := util.UnmarshalWithJSONPB(tt.installSpec, &is)
			if err != nil {
				t.Fatalf("yaml.Unmarshal(%s): got error %s", tt.desc, err)
			}

			tr, err := translate.NewTranslator(version.NewMinorVersion(1, 3))
			if err != nil {
				t.Fatal(err)
			}

			ins := NewIstioControlPlane(&is, tr)
			if err = ins.Run(); err != nil {
				t.Fatal(err)
			}

			got, errs := ins.RenderManifest()
			if len(errs) != 0 {
				t.Fatal(errs.Error())
			}
			want, err := readFile(tt.desc + ".yaml")
			if err != nil {
				t.Fatal(err)
			}
			diff, err := object.ManifestDiffWithSelectAndIgnore(manifestMapToStr(got), want, "", "")
			if err != nil {
				t.Fatal(err)
			}
			if diff != "" {
				t.Errorf("%s: got:\n%s\nwant:\n%s\n(-got, +want)\n%s\n", tt.desc, "", "", diff)
			}

		})
	}
}

func manifestMapToStr(mm name.ManifestMap) string {
	out := ""
	for _, m := range mm {
		out += m
	}
	return out
}

func readFile(path string) (string, error) {
	b, err := ioutil.ReadFile(filepath.Join(testDataDir, path))
	return string(b), err
}
