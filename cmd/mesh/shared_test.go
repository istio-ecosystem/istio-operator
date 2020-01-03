package mesh

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"istio.io/operator/pkg/util"
)

func TestReadLayeredYAMLs(t *testing.T) {
	testDataDir = filepath.Join(repoRootDir, "pkg/util/testdata/yaml")
	tests := []struct {
		name     string
		overlays []string
		wantErr  bool
	}{
		{
			name:     "single",
			overlays: []string{"first"},
			wantErr:  false,
		},
		{
			name:     "double",
			overlays: []string{"first", "second"},
			wantErr:  false,
		},
		{
			name:     "triple",
			overlays: []string{"first", "second", "third"},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inDir := filepath.Join(testDataDir, "input")
			outPath := filepath.Join(testDataDir, "output", tt.name+".yaml")
			wantBytes, err := ioutil.ReadFile(outPath)
			want := string(wantBytes)
			if err != nil {
				t.Errorf("ioutil.ReadFile() error = %v, filename: %v", err, outPath)
			}

			var filenames []string
			for _, ol := range tt.overlays {
				filenames = append(filenames, filepath.Join(inDir, ol+".yaml"))
			}
			got, err := ReadLayeredYAMLs(filenames)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadLayeredYAMLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if util.YAMLDiff(got, want) != "" {
				t.Errorf("ReadLayeredYAMLs() got = %v, want %v", got, want)
			}
		})
	}
}
