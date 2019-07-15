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
	"bytes"
	"io/ioutil"
	"istio.io/operator/pkg/util"
	"os"
	"path/filepath"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kylelemons/godebug/diff"

	"istio.io/operator/pkg/apis/istio/v1alpha2"
	"istio.io/operator/pkg/name"
	"istio.io/operator/pkg/translate"
	"istio.io/operator/pkg/version"
)

var (
	testDataDir      string
	helmChartTestDir string
	globalValuesFile string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	testDataDir = filepath.Join(wd, "testdata")
	helmChartTestDir = filepath.Join(testDataDir, "charts")
	globalValuesFile = filepath.Join(helmChartTestDir, "global.yaml")
}

func TestRenderInstallationSuccess(t *testing.T) {
	tests := []struct {
		desc        string
		installSpec string
	}{
		{
			desc: "all_off",
			installSpec: `
defaultNamespacePrefix: istio-system
trafficManagement:
  enabled: false
policy:
  enabled: false
telemetry:
  enabled: false
security:
  enabled: false
configManagement:
  enabled: false
autoInjection:
  enabled: false
`,
		},
		{
			desc: "pilot_default",
			installSpec: `
defaultNamespacePrefix: istio-system
policy:
  enabled: false
telemetry:
  enabled: false
security:
  enabled: false
configManagement:
  enabled: false
autoInjection:
  enabled: false
trafficManagement:
  enabled: true
  components:
    proxy:
      common:
        enabled: false
`,
		},
		{
			desc: "pilot_override_values",
			installSpec: `
defaultNamespacePrefix: istio-system
policy:
  enabled: false
telemetry:
  enabled: false
security:
  enabled: false
configManagement:
  enabled: false
autoInjection:
  enabled: false
trafficManagement:
  enabled: true
  components:
    namespace: istio-system
    proxy:
      common:
        enabled: false
    pilot:
      common:
        values:
          replicaCount: 5
          resources:
            requests:
              cpu: 111m
              memory: 222Mi
        unvalidatedValues:
          myCustomKey: someValue
`,
		},
		{
			desc: "pilot_override_kubernetes",
			installSpec: `
defaultNamespacePrefix: istio-system
policy:
  enabled: false
telemetry:
  enabled: false
security:
  enabled: false
configManagement:
  enabled: false
autoInjection:
  enabled: false
trafficManagement:
  enabled: true
  components:
    proxy:
      common:
        enabled: false
    pilot:
      common:
        k8s:
          overlays:
          - kind: Deployment
            name: istio-pilot
            patches:
            - path: spec.template.spec.containers.[name:discovery].args.[30m]
              value: "60m" # OVERRIDDEN
            - path: spec.template.spec.containers.[name:discovery].ports.[containerPort:8080].containerPort
              value: 1234 # OVERRIDDEN
          - kind: Service
            name: istio-pilot
            patches:
            - path: spec.ports.[name:grpc-xds].port
              value: 11111 # OVERRIDDEN
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var is v1alpha2.IstioControlPlaneSpec
			spec := `customPackagePath: "file://` + helmChartTestDir + `"` + "\n"
			spec += `profile: "file://` + helmChartTestDir + `/global.yaml"` + "\n"
			spec += tt.installSpec
			err := unmarshalWithJSONPB(spec, &is)
			if err != nil {
				t.Fatalf("yaml.Unmarshal(%s): got error %s", tt.desc, err)
			}

			ins := NewIstioControlPlane(&is, translate.Translators[version.NewMinorVersion(1, 2)])
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
			diff, err := util.ManifestDiff(manifestMapToStr(got), want)
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

func unmarshalWithJSONPB(y string, out proto.Message) error {
	jb, err := yaml.YAMLToJSON([]byte(y))
	if err != nil {
		return err
	}

	u := jsonpb.Unmarshaler{}
	err = u.Unmarshal(bytes.NewReader(jb), out)
	if err != nil {
		return err
	}
	return nil
}

func readFile(path string) (string, error) {
	b, err := ioutil.ReadFile(filepath.Join(testDataDir, path))
	return string(b), err
}

func YAMLDiff(a, b string) string {
	ao, bo := make(map[string]interface{}), make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(a), &ao); err != nil {
		return err.Error()
	}
	if err := yaml.Unmarshal([]byte(b), &bo); err != nil {
		return err.Error()
	}

	ay, err := yaml.Marshal(ao)
	if err != nil {
		return err.Error()
	}
	by, err := yaml.Marshal(bo)
	if err != nil {
		return err.Error()
	}

	return diff.Diff(string(ay), string(by))
}


/*func ObjectsInManifest(mstr string) string {
	ao, err := manifest.ParseObjectsFromYAMLManifest(mstr)
	if err != nil {
		return err.Error()
	}
	var out []string
	for _, v := range ao {
		out = append(out, v.Hash())
	}
	return strings.Join(out, "\n")
}*/
