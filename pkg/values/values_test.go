package values

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/kylelemons/godebug/diff"

	"istio.io/operator/pkg/apis/istio/v1alpha1"
)

const (
	valuesFilesDir = "testdata/values"
)

/*
func TestUnmarshalValues(t *testing.T) {
	tests := []struct {
		desc    string
		yamlStr string
		want    string
	}{}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tk := &v1alpha1.Values{}
			err := yaml.Unmarshal([]byte(tt.yamlStr), tk)
			if err != nil {
				t.Fatalf("yaml.Unmarshal(%s): got error %s", tt.desc, err)
			}
			s, err := yaml.Marshal(tk)
			if err != nil {
				t.Fatalf("yaml.Marshal(%s): got error %s", tt.desc, err)
			}
			got, want := stripNL(string(s)), stripNL(tt.want)
			if want == "" {
				want = stripNL(tt.yamlStr)
			}
			if !IsYAMLEqual(got, want) {
				t.Errorf("%s: got:\n%s\nwant:\n%s\n(-got, +want)\n%s\n", tt.desc, got, want, diff.Diff(got, want))
			}
		})
	}
}
*/

func TestUnmarshalRealValues(t *testing.T) {
	files, err := getFilesInDir(valuesFilesDir)
	if err != nil {
		t.Fatalf("getFiles: %v", err)
	}

	for _, f := range files {
		fs, err := readFile(f)
		if err != nil {
			t.Fatalf("readFile: %v", err)
		}
		t.Logf("Testing file %s", f)
		v := &v1alpha1.Values{}
		err = yaml.Unmarshal([]byte(fs), v)
		if err != nil {
			t.Fatalf("yaml.Unmarshal(%s): got error %s", f, err)
		}
		s, err := yaml.Marshal(v)
		if err != nil {
			t.Fatalf("yaml.Marshal(%s): got error %s", f, err)
		}
		got, want := stripNL(string(s)), stripNL(fs)
		if !IsYAMLEqual(got, want) {
			t.Errorf("%s: got:\n%s\nwant:\n%s\n(-got, +want)\n%s\n", f, got, want, YAMLDiff(got, want))
		}

	}

}

func getFilesInDir(dirPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func readFile(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	return string(b), err
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

// TODO: move to util
func IsYAMLEqual(a, b string) bool {
	if strings.TrimSpace(a) == "" && strings.TrimSpace(b) == "" {
		return true
	}
	ajb, err := yaml.YAMLToJSON([]byte(a))
	if err != nil {
		return false
	}
	bjb, err := yaml.YAMLToJSON([]byte(b))
	if err != nil {
		return false
	}

	return string(ajb) == string(bjb)
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
