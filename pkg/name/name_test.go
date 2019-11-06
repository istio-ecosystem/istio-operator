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

package name

import (
	"reflect"
	"testing"

	"istio.io/operator/pkg/util"
)

func TestGetFromTreePath(t *testing.T) {
	type args struct {
		inputTree map[string]interface{}
		path      util.Path
	}

	tests := []struct {
		name    string
		args    args
		want    interface{}
		found   bool
		wantErr bool
	}{
		{
			name: "found string node",
			args: args{
				inputTree: map[string]interface{}{
					"k1": "v1",
				},
				path: util.Path{"k1"},
			},
			want:    "v1",
			found:   true,
			wantErr: false,
		},
		{
			name: "found tree node",
			args: args{
				inputTree: map[string]interface{}{
					"k1": map[string]interface{}{
						"k2": "v2",
					},
				},
				path: util.Path{"k1"},
			},
			want: map[string]interface{}{
				"k2": "v2",
			},
			found:   true,
			wantErr: false,
		},
		{
			name: "path is longer than tree depth, string node",
			args: args{
				inputTree: map[string]interface{}{
					"k1": "v1",
				},
				path: util.Path{"k1", "k2"},
			},
			want:    nil,
			found:   false,
			wantErr: false,
		},
		{
			name: "path is longer than tree depth, array node",
			args: args{
				inputTree: map[string]interface{}{
					"k1": []interface{}{
						"v1",
					},
				},
				path: util.Path{"k1", "k2"},
			},
			want:    nil,
			found:   false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found, err := GetFromTreePath(tt.args.inputTree, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFromTreePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFromTreePath() got = %v, want %v", got, tt.want)
			}
			if found != tt.found {
				t.Errorf("GetFromTreePath() found = %v, want %v", found, tt.found)
			}
		})
	}
}
