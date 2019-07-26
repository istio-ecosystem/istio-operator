package version

import (
	"io/ioutil"
	"testing"

	goversion "github.com/hashicorp/go-version"
	"gopkg.in/yaml.v2"

	"istio.io/operator/pkg/version"
)

const (
	operatorVersionsMapFilePath = "./versions.yaml"
)

func TestVersions(t *testing.T) {
	operatorVersion, err := goversion.NewVersion(Version)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(operatorVersionsMapFilePath)
	if err != nil {
		t.Fatal(err)
	}
	var vs []version.IstioOperatorVersionCompatibility
	if err := yaml.Unmarshal(b, &vs); err != nil {
		t.Fatal(err)
	}

	for _, v := range vs {
		if operatorVersion.Equal(v.OperatorVersion) {
			t.Logf("Found operator version %s in %s file.", operatorVersion, operatorVersionsMapFilePath)
			return
		}
	}

}
