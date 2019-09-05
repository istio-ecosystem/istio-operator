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

package util

import (
	"encoding/json"
	"fmt"
)

func K8sJSONMerge(baseJSON, overlayJSON []byte, path Path) ([]byte, error) {
	var base, overlay interface{}
	err := json.Unmarshal(baseJSON, &base)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(overlayJSON, &overlay)
	if err != nil {
		return nil, err
	}

	dst, err := doK8sJSONMerge(base, overlay, path)
	if err != nil {
		return nil, err
	}

	dstJSON, err := json.Marshal(dst)
	if err != nil {
		return nil, err
	}

	return dstJSON, nil
}

func doK8sJSONMerge(base, overlay interface{}, path Path) (interface{}, error) {
	var remainPath Path
	var currentNode string
	if len(path) == 0 {
		currentNode = ""
		remainPath = PathFromString("")
	} else {
		currentNode = path[0]
		remainPath = path[1:]
	}
	switch base := base.(type) {
	case map[string]interface{}:
		overlay, ok := overlay.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid type for current node: %s, expect object type", currentNode)
		}
		for k, v2 := range overlay {
			if v1, ok := base[k]; ok {
				merged, err := doK8sJSONMerge(v1, v2, remainPath)
				if err != nil {
					return nil, err
				}
				base[k] = merged
			} else {
				base[k] = v2
			}
		}
	case []interface{}:
		overlay, ok := overlay.([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid type for current node: %s, expect slice type", currentNode)
		}
		dst := base[:0]
		dst = append(dst, base...)
		for _, v2 := range overlay {
			// non-leaf list, expect to match item by key.
			if v2, ok := v2.(map[interface{}]interface{}); ok {
				exists := false
				for j, v1 := range dst {
					if v1, ok := v1.(map[interface{}]interface{}); ok {
						_, nodeKey, err := PathKV(currentNode)
						if err != nil {
							return nil, fmt.Errorf("invalid current node: %s, expect [nodeName:nodekey]", currentNode)
						}
						if v2[nodeKey] == v1[nodeKey] {
							// if the item with matching key exists in base, then override base with merged value.
							exists = true
							merged, err := doK8sJSONMerge(v1, v2, remainPath)
							if err != nil {
								return nil, err
							}
							base[j] = merged
							break
						}
					}
				}
				if !exists {
					// if the item with matching key doesn't exist in base, then append to base.
					base = append(base, v2)
				}
			} else {
				base = append(base, v2)
			}
		}
	case nil:
		// merge nil base with overlay: (nil, map[string]interface{...}) -> map[string]interface{...}
		overlay, ok := overlay.(map[string]interface{})
		if ok {
			return overlay, nil
		}
	}
	return base, nil
}
