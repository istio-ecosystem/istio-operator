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
