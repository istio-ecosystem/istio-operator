package validate

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kylelemons/godebug/diff"
	"github.com/ostromart/istio-installer/pkg/util"
)

func TestUnmarshalKubernetes(t *testing.T) {
	tests := []struct {
		desc    string
		yamlStr string
		want    string
	}{
		{
			desc:    "nil success",
			yamlStr: "",
			want:    "{}",
		},
		{
			desc: "hpaSpec",
			yamlStr: `
hpaSpec:
  maxReplicas: 10
  minReplicas: 1
  targetCPUUtilizationPercentage: 80
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: php-apache
`,
		},
		{
			desc: "resources",
			yamlStr: `
resources:
  limits:
    cpu: 444m
    memory: 333Mi
  requests:
    cpu: 222m
    memory: 111Mi
`,
		},
		{
			desc: "podDisruptionBudget",
			yamlStr: `
podDisruptionBudget:
  maxUnavailable: 1
  selector:
    matchLabels:
      app: pilot
`,
		},
		{
			desc: "readinessProbe",
			yamlStr: `
readinessProbe:
  failureThreshold: 44
  initialDelaySeconds: 11
  periodSeconds: 22
  successThreshold: 33
  handler: {}
`,
		},
		{
			desc: "affinity",
			yamlStr: `
affinity:
  podAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: security
          operator: In
          values:
          - S1
      topologyKey: failure-domain.beta.kubernetes.io/zone
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: security
            operator: In
            values:
            - S2
        topologyKey: failure-domain.beta.kubernetes.io/zone
`,
		},
		{
			desc: "k8sObjectOverlay",
			yamlStr: `
overlays:
- kind: Deployment
  name: istio-citadel
  patches:
  - path: spec.template.spec.containers.name:galley.ports.containerPort:15014
    value: 12345
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tk := &v1alpha1.TestKube{}
			err := unmarshalWithJSONPB(tt.yamlStr, tk)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			s, err := marshalWithJSONPB(tk)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			got, want := stripNL(s), stripNL(tt.want)
			if want == "" {
				want = stripNL(tt.yamlStr)
			}
			if !util.IsYAMLEqual(got, want) {
				t.Errorf("%s: got:\n%s\nwant:\n%s\n(-got, +want)\n%s\n", tt.desc, got, want, diff.Diff(got, want))
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		desc     string
		yamlStr  string
		wantErrs util.Errors
	}{
		{
			desc: "nil success",
		},
		{
			desc: "TrafficManagement",
			yamlStr: `
trafficManagement:
  enabled: true
  namespace: istio-system-traffic
`,
		},
		{
			desc: "PilotConfig",
			yamlStr: `
trafficManagement:
  pilot:
    sidecar: true
`,
		},
		{
			desc: "SidecarInjectorConfig",
			yamlStr: `
trafficManagement:
  sidecarInjector:
    enableNamespacesByDefault: true
`,
		},
		{
			desc: "CommonConfig",
			yamlStr: `
hub: docker.io/istio
tag: v1.2.3
trafficManagement:
  proxy:
    common:
      enabled: true
      namespace: istio-control-system
      debug: INFO
      k8s:
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        readinessProbe:
          initialDelaySeconds: 11
          periodSeconds: 22
          successThreshold: 33
          failureThreshold: 44
        hpaSpec:
          scaleTargetRef:
            apiVersion: apps/v1
            kind: Deployment
            name: php-apache
          minReplicas: 1
          maxReplicas: 10
          targetCPUUtilizationPercentage: 80
        nodeSelector:
          disktype: ssd
`,
		},
		{
			desc: "BadTag",
			yamlStr: `
hub: ?illegal-tag!
`,
			wantErrs: makeErrors([]string{`invalid value Hub:?illegal-tag!`}),
		},
		{
			desc: "BadHub",
			yamlStr: `
hub: docker.io:tag/istio
`,
			wantErrs: makeErrors([]string{`invalid value Hub:docker.io:tag/istio`}),
		},
		{
			desc: "GoodURL",
			yamlStr: `
customPackagePath: file://local/file/path
`,
		},
		{
			desc: "BadURL",
			yamlStr: `
customPackagePath: bad_schema://local/file/path
`,
			wantErrs: makeErrors([]string{`invalid value CustomPackagePath:bad_schema://local/file/path`}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ispec := &v1alpha1.InstallerSpec{}
			err := unmarshalWithJSONPB(tt.yamlStr, ispec)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			errs := ValidateInstallerSpec(ispec)
			if gotErrs, wantErrs := errs, tt.wantErrs; !util.EqualErrors(gotErrs, wantErrs) {
				t.Errorf("ProtoToValues(%s)(%v): gotErrs:%s, wantErrs:%s", tt.desc, tt.yamlStr, gotErrs, wantErrs)
			}
		})
	}
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

func marshalWithJSONPB(in *v1alpha1.TestKube) (string, error) {
	m := jsonpb.Marshaler{}
	js, err := m.MarshalToString(in)
	if err != nil {
		return "", err
	}
	yb, err := yaml.JSONToYAML([]byte(js))
	if err != nil {
		return "", err
	}
	return string(yb), nil
}

// errToString returns the string representation of err and the empty string if
// err is nil.
func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func stripNL(s string) string {
	return strings.Trim(s, "\n")
}
