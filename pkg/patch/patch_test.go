package patch

import (
	"bytes"
	"fmt"
	"github.com/ostromart/istio-installer/pkg/util"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
)

func TestPatchYAMLManifestSuccess(t *testing.T) {
	base := `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
a:
  b:
  - name: n1
    value: v1
  - name: n2
    list: 
    - v1
    - v2
    - v3_regex
`
	tests := []struct {
		desc    string
		path    string
		value   string
		want    string
		wantErr string
	}{
		{
			desc: "ModifyListEntryValue",
			path: `a.b.name:n1.value`,
			value: `v2`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
a:
  b:
  - name: n1
    value: v2
  - list:
    - v1
    - v2
    - v3_regex
    name: n2
`,
		},
		{
			desc: "ModifyListEntry",
			path: `a.b.name:n2.list.:v2`,
			value: `v3`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
a:
  b:
  - name: n1
    value: v1
  - list:
    - v1
    - v3
    - v3_regex
    name: n2
`,
		},
		{
			desc: "DeleteListEntry",
			path: `a.b.name:n1`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
a:
  b:
  - list:
    - v1
    - v2
    - v3_regex
    name: n2
`,
		},
		{
			desc: "DeleteListEntryValue",
			path: `a.b.name:n2.list.:v2`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
a:
  b:
  - name: n1
    value: v1
  - list:
    - v1
    - v3_regex
    name: n2
`,
		},
		{
			desc: "DeleteListEntryValueRegex",
			path: `a.b.name:n2.list.:v3`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
a:
  b:
  - name: n1
    value: v1
  - list:
    - v1
    - v2
    name: n2
`,
		},	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			rc := &v1alpha1.KubernetesResourcesSpec{}
			err := unmarshalWithJSONPB(makeOverlay(tt.path, tt.value), rc)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			got, err := PatchYAMLManifest(base, "istio-system", rc.Overlays)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Fatalf("PatchYAMLManifest(%s): gotErr:%s, wantErr:%s", tt.desc, gotErr, wantErr)
			}
			if want := tt.want; !util.IsYAMLEqual(got, want) {
				t.Errorf("PatchYAMLManifest(%s): got:\n%s\n\nwant:\n%s\nDiff:\n%s\n", tt.desc, got, want, util.YAMLDiff(got, want))
			}
		})
	}
}

func TestPatchYAMLManifestRealYAMLSuccess(t *testing.T) {
	base := `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
spec:
  template:
    spec:
      containers:
      - name: deleteThis
        foo: bar
      - name: galley
        ports:
        - containerPort: 443
        - containerPort: 15014
        - containerPort: 9901
        command:
        - /usr/local/bin/galley
        - server
        - --meshConfigFile=/etc/mesh-config/mesh
        - --livenessProbeInterval=1s
        - --validation-webhook-config-file
`

	tests := []struct {
		desc    string
		path    string
		value   string
		want    string
		wantErr string
	}{
		{
			desc: "DeleteLeafListLeaf",
			path: `spec.template.spec.containers.name:galley.command.:--validation-webhook-config-file`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
spec:
  template:
    spec:
      containers:
      - foo: bar
        name: deleteThis
      - command:
        - /usr/local/bin/galley
        - server
        - --meshConfigFile=/etc/mesh-config/mesh
        - --livenessProbeInterval=1s
        name: galley
        ports:
        - containerPort: 443
        - containerPort: 15014
        - containerPort: 9901
`,
		},
		{
			desc:  "UpdateListItem",
			path:  `spec.template.spec.containers.name:galley.command.:--livenessProbeInterval`,
			value: `--livenessProbeInterval=1111s`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
spec:
  template:
    spec:
      containers:
      - foo: bar
        name: deleteThis
      - command:
        - /usr/local/bin/galley
        - server
        - --meshConfigFile=/etc/mesh-config/mesh
        - --livenessProbeInterval=1111s
        - --validation-webhook-config-file
        name: galley
        ports:
        - containerPort: 443
        - containerPort: 15014
        - containerPort: 9901
`,
		},
		{
			desc:  "UpdateLeaf",
			path:  `spec.template.spec.containers.name:galley.ports.containerPort:15014.containerPort`,
			value: `22222`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
spec:
  template:
    spec:
      containers:
      - foo: bar
        name: deleteThis
      - command:
        - /usr/local/bin/galley
        - server
        - --meshConfigFile=/etc/mesh-config/mesh
        - --livenessProbeInterval=1s
        - --validation-webhook-config-file
        name: galley
        ports:
        - containerPort: 443
        - containerPort: 22222
        - containerPort: 9901
`,
		},
		{
			desc: "DeleteLeafList",
			path: `spec.template.spec.containers.name:galley.ports.containerPort:9901`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
spec:
  template:
    spec:
      containers:
      - foo: bar
        name: deleteThis
      - command:
        - /usr/local/bin/galley
        - server
        - --meshConfigFile=/etc/mesh-config/mesh
        - --livenessProbeInterval=1s
        - --validation-webhook-config-file
        name: galley
        ports:
        - containerPort: 443
        - containerPort: 15014
`,
		},
		{
			desc: "DeleteInternalNode",
			path: `spec.template.spec.containers.name:deleteThis`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
spec:
  template:
    spec:
      containers:
      - command:
        - /usr/local/bin/galley
        - server
        - --meshConfigFile=/etc/mesh-config/mesh
        - --livenessProbeInterval=1s
        - --validation-webhook-config-file
        name: galley
        ports:
        - containerPort: 443
        - containerPort: 15014
        - containerPort: 9901
`,
		},
		{
			desc: "DeleteLeafListentry",
			path: `spec.template.spec.containers.name:galley.command.:--validation-webhook-config-file`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
spec:
  template:
    spec:
      containers:
      - foo: bar
        name: deleteThis
      - command:
        - /usr/local/bin/galley
        - server
        - --meshConfigFile=/etc/mesh-config/mesh
        - --livenessProbeInterval=1s
        name: galley
        ports:
        - containerPort: 443
        - containerPort: 15014
        - containerPort: 9901
`,
		},
		{
			desc: "UpdateInteriorNode",
			path: `spec.template.spec.containers.name:galley.ports.containerPort:15014`,
			value: `
      fooPort: 15015`,
			want: `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: istio-citadel
  namespace: istio-system
spec:
  template:
    spec:
      containers:
      - foo: bar
        name: deleteThis
      - command:
        - /usr/local/bin/galley
        - server
        - --meshConfigFile=/etc/mesh-config/mesh
        - --livenessProbeInterval=1s
        - --validation-webhook-config-file
        name: galley
        ports:
        - containerPort: 443
        - fooPort: 15015
        - containerPort: 9901

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			rc := &v1alpha1.KubernetesResourcesSpec{}
			err := unmarshalWithJSONPB(makeOverlay(tt.path, tt.value), rc)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			got, err := PatchYAMLManifest(base, "istio-system", rc.Overlays)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Fatalf("PatchYAMLManifest(%s): gotErr:%s, wantErr:%s", tt.desc, gotErr, wantErr)
			}
			if want := tt.want; !util.IsYAMLEqual(got, want) {
				t.Errorf("PatchYAMLManifest(%s): got:\n%s\n\nwant:\n%s\nDiff:\n%s\n", tt.desc, got, want, util.YAMLDiff(got, want))
			}
		})
	}
}

func makeOverlay(path, value string) string {
	const (
		patchCommon = `
overlays:
- kind: Deployment
  name: istio-citadel
  patches:
`
		pathStr  = `  - path: `
		valueStr = `    value: `
	)

	ret := patchCommon
	ret += fmt.Sprintf("%s%s\n", pathStr, path)
	if value != "" {
		ret += fmt.Sprintf("%s%s\n", valueStr, value)
	}
	return ret
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

// errToString returns the string representation of err and the empty string if
// err is nil.
func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
