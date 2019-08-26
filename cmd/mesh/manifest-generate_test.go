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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"istio.io/operator/pkg/object"
)

type testGroup []struct {
	desc       string
	flags      string
	diffSelect string
	diffIgnore string
}

var (
	repoRootDir string
	testDataDir string
)

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	repoRootDir = filepath.Join(wd, "../..")
	testDataDir = filepath.Join(wd, "testdata/manifest-generate")

	if err := syncCharts(); err != nil {
		panic(err)
	}
}

func syncCharts() error {
	cmd, err := exec.Command(filepath.Join(repoRootDir, "scripts/run_update_charts.sh")).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s:%s", err, cmd)
	}
	return nil
}

func TestManifestGenerateFlags(t *testing.T) {
	runTestGroup(t, testGroup{
		{
			desc: "all_off",
		},
		{
			desc:       "all_on",
			diffIgnore: "ConfigMap:*:istio",
		},
		{
			desc:       "flag_set_values",
			diffIgnore: "ConfigMap:*:istio",
			flags:      "-s values.global.proxy.image=myproxy",
		},
		// TODO: test output flag
	})
}

func TestManifestGeneratePilot(t *testing.T) {
	runTestGroup(t, testGroup{
		{
			desc: "pilot_default",
			// TODO: remove istio ConfigMap
			diffIgnore: "CustomResourceDefinition:*:*,ConfigMap:*:istio",
		},
		{
			desc:       "pilot_k8s_settings",
			diffIgnore: "CustomResourceDefinition:*:*,ConfigMap:*:istio",
		},
		{
			desc:       "pilot_override_values",
			diffSelect: "Deployment:*:istio-pilot",
		},
		{
			desc:       "pilot_override_kubernetes",
			diffSelect: "Deployment:*:istio-pilot, Service:*:istio-pilot",
		},
	})
}

func TestManifestGenerateTelemetry(t *testing.T) {
	runTestGroup(t, testGroup{
		{
			desc: "all_off",
		},
		{
			desc:       "telemetry_default",
			diffIgnore: "",
		},
		{
			desc:       "telemetry_k8s_settings",
			diffSelect: "Deployment:*:istio-telemetry, HorizontalPodAutoscaler:*:istio-telemetry",
		},
		{
			desc:       "telemetry_override_values",
			diffSelect: "handler:*:prometheus",
		},
		{
			desc:       "telemetry_override_kubernetes",
			diffSelect: "Deployment:*:istio-telemetry, handler:*:prometheus",
		},
	})
}

func runTestGroup(t *testing.T, tests testGroup) {
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			inPath := filepath.Join(testDataDir, "input", tt.desc+".yaml")
			outPath := filepath.Join(testDataDir, "output", tt.desc+".yaml")

			got, err := runManifestGenerate(inPath, tt.flags)
			if err != nil {
				t.Fatal(err)
			}

			want, err := readFile(outPath)
			if err != nil {
				t.Fatal(err)
			}

			diffSelect := "*:*:*"
			if tt.diffSelect != "" {
				diffSelect = tt.diffSelect
			}
			diff, err := object.ManifestDiffWithSelectAndIgnore(got, want, diffSelect, tt.diffIgnore)
			if err != nil {
				t.Fatal(err)
			}
			if diff != "" {
				t.Errorf("%s: got:\n%s\nwant:\n%s\n(-got, +want)\n%s\n", tt.desc, "", "", diff)
			}

		})
	}
}

func runCommand(command string) (string, error) {
	var out bytes.Buffer
	rootCmd := GetRootCmd(strings.Split(command, " "))
	rootCmd.SetOutput(&out)

	if err := rootCmd.Execute(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func runManifestGenerate(path, flags string) (string, error) {
	args := "manifest generate " + flags
	if flags == "" {
		args += " -f " + path
	}
	return runCommand(args)
}

func readFile(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	return string(b), err
}
