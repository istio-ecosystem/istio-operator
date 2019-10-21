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
	"fmt"
	"os"
	"testing"

	goversion "github.com/hashicorp/go-version"

	"istio.io/operator/pkg/version"
)

func TestGetVersionCompatibleMap(t *testing.T) {
	type args struct {
		versionsURI string
		binVersion  *goversion.Version
		l           *logger
	}
	goVer131000, _ := goversion.NewVersion("1.3.1000")
	goVer132, _ := goversion.NewVersion("1.3.2")
	scVer132, _ := goversion.NewConstraint(">=1.3.0,<=1.3.2")
	rcVer132, _ := goversion.NewConstraint("1.3.2")
	vm132 := &version.CompatibilityMapping{
		OperatorVersion:          goVer132,
		SupportedIstioVersions:   scVer132,
		RecommendedIstioVersions: rcVer132,
	}
	l := newLogger(true, os.Stdout, os.Stderr)
	tests := []struct {
		name    string
		args    args
		want    *version.CompatibilityMapping
		wantErr error
	}{
		{
			name: "built-in version map",
			args: args{
				versionsURI: "__nonexistent-versions.yaml",
				binVersion:  goVer132,
				l:           l,
			},
			want:    vm132,
			wantErr: nil,
		},
		{
			name: "read from github",
			args: args{
				versionsURI: versionsMapURL,
				binVersion:  goVer132,
				l:           l,
			},
			want:    vm132,
			wantErr: nil,
		},
		{
			name: "go version not found in version map",
			args: args{
				versionsURI: "__nonexistent-versions.yaml",
				binVersion:  goVer131000,
				l:           l,
			},
			want: nil,
			wantErr: fmt.Errorf("this operator version %s was not found in the global manifestVersions map",
				goVer131000.String()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getVersionCompatibleMap(tt.args.versionsURI, tt.args.binVersion, tt.args.l)
			if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", tt.want) {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
			if errToString(gotErr) != errToString(tt.wantErr) {
				t.Errorf("gotErr: %v, wantErr: %v", gotErr, tt.wantErr)
			}
		})
	}
}

// errToString returns the string representation of err and the empty string if
// err is nil.
func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
