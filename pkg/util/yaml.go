package util

import (
	"bytes"
	"strings"

	"istio.io/pkg/log"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/kylelemons/godebug/diff"
	"sigs.k8s.io/yaml"
)

// IsYAMLEqual reports whether the YAML in strings a and b are equal.
func IsYAMLEqual(a, b string) bool {
	if strings.TrimSpace(a) == "" && strings.TrimSpace(b) == "" {
		return true
	}
	ajb, err := yaml.YAMLToJSON([]byte(a))
	if err != nil {
		dbgPrint("bad YAML in isYAMLEqual:\n%s", a)
		return false
	}
	bjb, err := yaml.YAMLToJSON([]byte(b))
	if err != nil {
		dbgPrint("bad YAML in isYAMLEqual:\n%s", b)
		return false
	}

	return string(ajb) == string(bjb)
}

// ToYAML returns a YAML string representation of val, or the error string if an error occurs.
func ToYAML(val interface{}) string {
	y, err := yaml.Marshal(val)
	if err != nil {
		return err.Error()
	}
	return string(y)
}

// ToYAMLWithJSONPB returns a YAML string representation of val (using jsonpb), or the error string if an error occurs.
func ToYAMLWithJSONPB(val proto.Message) string {
	m := jsonpb.Marshaler{}
	js, err := m.MarshalToString(val)
	if err != nil {
		return err.Error()
	}
	yb, err := yaml.JSONToYAML([]byte(js))
	if err != nil {
		return err.Error()
	}
	return string(yb)
}

// UnmarshalWithJSONPB unmarshals y into out using jsonpb (required for many proto defined structs).
func UnmarshalWithJSONPB(y string, out proto.Message) error {
	jb, err := yaml.YAMLToJSON([]byte(y))
	if err != nil {
		return err
	}

	u := jsonpb.Unmarshaler{AllowUnknownFields: false}
	err = u.Unmarshal(bytes.NewReader(jb), out)
	if err != nil {
		return err
	}
	return nil
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

func ManifestDiff(a, b string) (string, error) {
	ao, err := ParseK8sObjectsFromYAMLManifest(a)
	if err != nil {
		return "", err
	}
	bo, err := ParseK8sObjectsFromYAMLManifest(b)
	if err != nil {
		return "", err
	}
	aom, bom := ao.ToMap(), bo.ToMap()
	var sb strings.Builder
	for ak, av := range aom {
		ay, err := av.YAML()
		if err != nil {
			return "", err
		}
		by, err := bom[ak].YAML()
		if err != nil {
			return "", err
		}
		diff := YAMLDiff(string(ay), string(by))
		if diff != "" {
			writeStringSafe(sb, "\n\nObject "+ak+" has diffs:\n\n")
			writeStringSafe(sb, diff)
		}
	}
	for bk, bv := range bom {
		if aom[bk] == nil {
			by, err := bv.YAML()
			if err != nil {
				return "", err
			}
			diff := YAMLDiff(string(by), "")
			if diff != "" {
				writeStringSafe(sb, "\n\nObject "+bk+" is missing:\n\n")
				writeStringSafe(sb, diff)
			}
		}
	}
	return sb.String(), err
}

func writeStringSafe(sb strings.Builder, s string) {
	_, err := sb.WriteString(s)
	if err != nil {
		log.Error(err.Error())
	}
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
