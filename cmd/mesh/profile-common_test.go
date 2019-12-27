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
	"testing"

	"istio.io/operator/pkg/compare"
)

func Test_icpsToIcp(t *testing.T) {
	tests := []struct {
		name    string
		icpsy   string
		want    string
		wantErr bool
	}{
		{
			name: "pilot",
			icpsy: `
policy:
  enabled: false
values:
  global:
    useMCP: false
  pilot:
    sidecar: false
    useMCP: false
`,
			want: `
apiVersion: install.istio.io/v1alpha2
kind: IstioControlPlane
spec:
  policy:
    enabled: false
  values:
    global:
      useMCP: false
    pilot:
      sidecar: false
      useMCP: false
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := icpsToIcp(tt.icpsy)
			if (err != nil) != tt.wantErr {
				t.Errorf("icpsToIcp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ds, err := compare.ManifestDiff(got, tt.want, false)
			if err != nil {
				t.Errorf("compare.ManifestDiff() error = %v", err)
				return
			}
			if ds != "" {
				t.Errorf("icpsToIcp() got = %v, want %v", got, tt.want)
			}
		})
	}
}
