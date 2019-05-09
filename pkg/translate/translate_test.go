package translate

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kr/pretty"
	"github.com/kylelemons/godebug/diff"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/util"
)

func TestSetYAML(t *testing.T) {
	tests := []struct {
		desc    string
		root    util.Tree
		path    string
		value   string
		want    string
		wantErr string
	}{
		{
			desc:    "insert no path",
			path:    "",
			value:   "val1",
			want:    `val1`,
			wantErr: "path cannot be empty",
		},
		{
			desc:  "insert empty",
			path:  "a.b.c",
			value: "val1",
			want: `a:
  b:
    c: val1
`,
		},
		{
			desc: "overwrite",
			root: util.Tree{
				"a": util.Tree{
					"b": util.Tree{
						"c": "val1",
					},
				},
			},
			path:  "a.b.c",
			value: "val2",
			want: `a:
  b:
    c: val2
`,
		},
		{
			desc: "partial create",
			root: util.Tree{
				"a": util.Tree{
					"b": util.Tree{
						"c": "val1",
					},
				},
			},
			path:  "a.b.d",
			value: "val2",
			want: `a:
  b:
    c: val1
    d: val2
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			root := tt.root
			if root == nil {
				root = make(util.Tree)
			}
			p := util.PathFromString(tt.path)
			err := setYAML(root, p, tt.value)
			fmt.Println(err)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Errorf("TestSetYAML()%s: gotErr:%s, wantErr:%s", tt.desc, gotErr, wantErr)
				return
			}
			if got, want := root.String(), tt.want; err == nil && got != want {
				t.Errorf("TestSetYAML(%s) got:\n%s\nwant:\n%s\ndiff:\n%s\n", tt.desc, got, want, diff.Diff(got, want))
			}
		})
	}
}

func TestProtoToValues(t *testing.T) {
	tests := []struct {
		desc string
		// mappings defaults to defaultMappings
		mappings map[string]*Translation
		yamlStr  string
		want     string
		wantErr  string
	}{
		{
			desc: "nil success",
			want: "",
		},
		{
			desc: "IstioInstaller",
			yamlStr: `
hub: docker.io/istio
tag: 1.2.3
k8sDefaults:
  resources:
    requests:
      cpu: "250m"
`,
			want: `
global:
  hub: docker.io/istio
  tag: 1.2.3
  defaultResources:
    requests:
      cpu: "250m"
`,
		},
		{
			desc: "Security",
			yamlStr: `
security:
  controlPlaneMtls: true
  dataPlaneMtlsStrict: false
`,
			want: `
global:
  controlPlaneSecurityEnabled: true
  mtls:
    enabled: false
`,
		},
		{
			desc: "SidecarInjector",
			yamlStr: `
trafficManagement:
  sidecarInjector:
    enableNamespacesByDefault: false
`,
			want: `
sidecarInjectorWebhook:
  enableNamespacesByDefault: false
`,
		},
	}

	for _, tt := range tests {
		fmt.Println(tt.desc)
		t.Run(tt.desc, func(t *testing.T) {
			mappings := tt.mappings
			if mappings == nil {
				mappings = defaultMappings
			}
			ispec := &v1alpha1.InstallerSpec{}
			err := unmarshalWithJSONPB(tt.yamlStr, ispec)
			if err != nil {
				t.Fatalf("unmarshalWithJSONPB(%s): got error %s", tt.desc, err)
			}
			dbgPrint("ispec: \n%s\n", pretty.Sprint(ispec))
			got, err := ProtoToValues(mappings, ispec)
			if gotErr, wantErr := errToString(err), tt.wantErr; gotErr != wantErr {
				t.Errorf("ProtoToValues(%s)(%v): gotErr:%s, wantErr:%s", tt.desc, tt.yamlStr, gotErr, wantErr)
			}
			if got, want := stripNL(got), stripNL(tt.want); err == nil && !util.IsYAMLEqual(got, want) {
				t.Errorf("ProtoToValues(%s) got:\n%s\nwant:\n%s\n", tt.desc, got, want)
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
