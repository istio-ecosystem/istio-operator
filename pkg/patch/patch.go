/*
Package patch implements a simple patching mechanism for k8s resources.
Paths are specified in the form a.b.c.[key:value].d.[list_entry_value], where:
 -  [key:value] selects a list entry in list c which contains an entry with key:value
 -  [list_entry_value] selects a list entry in list d which is a regex match of list_entry_value.

Some examples are given below. Given a resource:

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
    - "vv1"
    - vv2=foo

values and list entries can be added, modifed or deleted.

MODIFY

1. set v1 to v1new

  path: a.b.[name:n1].value
  value: v1

2. set vv1 to vv3

  // Note the lack of quotes around vv1 (see NOTES below).
  path: a.b.[name:n2].list.[vv1]
  value: vv3

3. set vv2=foo to vv2=bar (using regex match)

  path: a.b.[name:n2].list.[vv2]
  value: vv2=bar

DELETE

1. Delete container with name: n1

  path: a.b.[name:n1]

2. Delete list value vv1

  path: a.b.[name:n2].list.[vv1]

ADD

1. Add vv3 to list

  path: a.b.[name:n2].list
  value: vv3

2. Add new key:value to container name: n1

  path: a.b.[name:n1]
  value:
    new_attr: v3

*NOTES*
- Due to loss of string quoting during unmarshaling, keys and values should not be string quoted, even if they appear
that way in the object being patched.
- [key:value] treats ':' as a special separator character. Any ':' in the key or value string must be escaped as \:.
*/

package patch

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/kr/pretty"
	"github.com/ostromart/istio-installer/pkg/apis/installer/v1alpha1"
	"github.com/ostromart/istio-installer/pkg/manifest"
	"github.com/ostromart/istio-installer/pkg/util"
	"gopkg.in/yaml.v2"
	"istio.io/istio/pkg/log"
)

var (
	// debugPackage controls verbose debugging in this package. Used for offline debugging.
	debugPackage = false
)

// pathContext provides a means for traversing a tree towards the root.
type pathContext struct {
	// parent in the parent of this pathContext.
	parent *pathContext
	// keyToChild is the key required to reach the child.
	keyToChild interface{}
	// node is the actual node in the data tree.
	node interface{}
}

// String implements the Stringer interface.
func (nc *pathContext) String() string {
	ret := "\n--------------- NodeContext ------------------\n"
	ret += fmt.Sprintf("parent.node=\n%s\n", pretty.Sprint(nc.parent.node))
	ret += fmt.Sprintf("keyToChild=%v\n", nc.parent.keyToChild)
	ret += fmt.Sprintf("node=\n%s\n", pretty.Sprint(nc.node))
	ret += "----------------------------------------------\n"
	return ret
}

// makeNodeContext returns a pathContext created from the given object.
func makeNodeContext(obj interface{}) *pathContext {
	return &pathContext{
		node: obj,
	}
}

// PatchYAMLManifest patches a base YAML in the given namespace with a list of overlays.
// Each overlay has the format described in the K8SObjectOverlay definition.
// It returns the patched manifest YAML.
func PatchYAMLManifest(baseYAML string, namespace string, overlays []*v1alpha1.K8SObjectOverlay) (string, error) {
	baseObjs, err := manifest.ParseObjectsFromYAMLManifest(context.TODO(), baseYAML)
	if err != nil {
		return "", err
	}

	bom := baseObjs.ToMap()
	oom, err := objectOverrideMap(overlays, namespace)
	if err != nil {
		return "", err
	}
	var ret strings.Builder

	// Try to apply the defined overlays.
	for k, oo := range oom {
		bo := bom[k]
		if bo == nil {
			// TODO: error log overlays with no matches in any component.
			continue
		}
		patched, err := applyPatches(bo, oo)
		if err != nil {
			log.Errorf("patch error: %s", err)
			continue
		}
		ret.Write(patched)
		ret.WriteString("\n---\n")
	}
	// Render the remaining objects with no overlays.
	for k, oo := range bom {
		if oom[k] != nil {
			// Skip objects that have overlays, these were rendered above.
			continue
		}
		oy, err := oo.YAML()
		if err != nil {
			log.Errorf("Object to YAML error (%s) for base object: \n%v", err, oo)
			continue
		}
		ret.Write(oy)
		ret.WriteString("\n---\n")
	}
	return ret.String(), nil
}

// applyPatches applies the given patches against the given object. It returns the resulting patched YAML if successful,
// or a list of errors otherwise.
func applyPatches(base *manifest.Object, patches []*v1alpha1.K8SObjectOverlay_PathValue) (outYAML []byte, errs util.Errors) {
	bo := make(map[interface{}]interface{})
	by, err := base.YAML()
	if err != nil {
		return nil, util.NewErrs(err)
	}
	err = yaml.Unmarshal(by, bo)
	if err != nil {
		return nil, util.NewErrs(err)
	}
	for _, p := range patches {
		dbgPrint("applying path=%s, value=%s\n", p.Path, p.Value)
		path := util.PathFromString(p.Path)
		inc, err := getNode(makeNodeContext(bo), path, path)
		if err != nil {
			errs = util.AppendErr(errs, err)
			continue
		}
		errs = util.AppendErr(errs, writeNode(inc, p.Value))
	}
	oy, err := yaml.Marshal(bo)
	if err != nil {
		return nil, util.AppendErr(errs, err)
	}
	return oy, errs
}

// getNode returns the node which has the given patch from the source node given by nc.
// It creates a tree of nodeContexts during the traversal so that parent structures can be updated if required.
func getNode(nc *pathContext, fullPath, remainPath util.Path) (*pathContext, error) {
	dbgPrint("getNode remainPath=%s, node=%s", remainPath, pretty.Sprint(nc.node))
	if len(remainPath) == 0 {
		dbgPrint("terminate with nc=%s", nc)
		return nc, nil
	}
	pe := remainPath[0]

	v := reflect.ValueOf(nc.node)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	ncNode := v.Interface()
	// For list types, we need a key to identify the selected list item. This can be either a a value key of the
	// form :matching_value in the case of a leaf list, or a matching key:value in the case of a non-leaf list.
	if lst, ok := ncNode.([]interface{}); ok {
		dbgPrint("list type")
		for idx, le := range lst {
			// non-leaf list, expect to match item by key:value.
			if lm, ok := le.(map[interface{}]interface{}); ok {
				k, v, err := pathKV(pe)
				if err != nil {
					return nil, fmt.Errorf("path %s: %s", fullPath, err)
				}
				if stringsEqual(lm[k], v) {
					dbgPrint("found matching kv %v:%v", k, v)
					nn := &pathContext{
						parent: nc,
						node:   lm,
					}
					nc.keyToChild = idx
					nn.keyToChild = k
					if len(remainPath) == 1 {
						dbgPrint("KV terminate")
						return nn, nil
					}
					return getNode(nn, fullPath, remainPath[1:])
				}
				continue
			}
			// leaf list, match based on value.
			v, err := pathV(pe)
			if err != nil {
				return nil, fmt.Errorf("path %s: %s", fullPath, err)
			}
			if matchesRegex(v, le) {
				dbgPrint("found matching key %v, index %d", le, idx)
				nn := &pathContext{
					parent: nc,
					node:   le,
				}
				nc.keyToChild = idx
				return getNode(nn, fullPath, remainPath[1:])
			}
		}
		return nil, fmt.Errorf("path %s: element %s not found", fullPath, pe)
	}

	dbgPrint("interior node")
	// non-list node.
	if nnt, ok := nc.node.(map[interface{}]interface{}); ok {
		var nn interface{}
		nn = nnt[pe]
		nnc := &pathContext{
			parent: nc,
			node:   nn,
		}
		if _, ok := nn.([]interface{}); ok {
			// Slices must be passed by pointer for mutations.
			nnc.node = &nn
		}
		nc.keyToChild = pe
		return getNode(nnc, fullPath, remainPath[1:])
	}

	return nil, fmt.Errorf("leaf type %T in non-leaf node %s", nc.node, remainPath)
}

// writeNode writes the given value to the node in the given pathContext.
func writeNode(nc *pathContext, value interface{}) error {
	dbgPrint("writeNode pathContext=%s, value=%v", nc, value)

	switch {
	case value == nil:
		dbgPrint("delete")
		switch {
		case nc.parent != nil && isSliceOrPtrInterface(nc.parent.node):
			if err := util.DeleteFromSlicePtr(nc.parent.node, nc.parent.keyToChild.(int)); err != nil {
				return err
			}
			// FIXME
			if isMapOrInterface(nc.parent.parent.node) {
				if err := util.InsertIntoMap(nc.parent.parent.node, nc.parent.parent.keyToChild, nc.parent.node); err != nil {
					return err
				}
			}
		}
	default:
		switch {
		case isSliceOrPtrInterface(nc.parent.node):
			idx := nc.parent.keyToChild.(int)
			if idx == -1 {
				dbgPrint("insert")

			} else {
				dbgPrint("update index %d\n", idx)
				if err := util.UpdateSlicePtr(nc.parent.node, idx, value); err != nil {
					return err
				}
			}
		default:
			dbgPrint("leaf update")
			if isMapOrInterface(nc.parent.node) {
				if err := util.InsertIntoMap(nc.parent.node, nc.parent.keyToChild, value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func isValidPathElement(pe string) bool {
	return util.ValidKeyRegex.MatchString(pe)
}

func removeBrackets(pe string) (string, bool) {
	if !strings.HasPrefix(pe, "[") || !strings.HasSuffix(pe, "]") {
		return "", false
	}
	return pe[1 : len(pe)-1], true
}

func isKVPathElement(pe string) bool {
	pe, ok := removeBrackets(pe)
	if !ok {
		return false
	}

	kv := strings.Split(pe, ":")
	if len(kv) != 2 || len(kv[0]) == 0 || len(kv[1]) == 0 {
		return false
	}
	return isValidPathElement(kv[0])
}

func isVPathElement(pe string) bool {
	pe, ok := removeBrackets(pe)
	if !ok {
		return false
	}

	return len(pe) > 0
}

func pathV(pe string) (string, error) {
	if !isVPathElement(pe) {
		return "", fmt.Errorf("%s is not a valid value path element", pe)
	}
	v, _ := removeBrackets(pe)
	return v, nil
}

func pathKV(pe string) (k, v string, err error) {
	if !isKVPathElement(pe) {
		return "", "", fmt.Errorf("%s is not a valid key:value path element", pe)
	}
	pe, _ = removeBrackets(pe)
	kv := strings.Split(pe, ":")
	return kv[0], kv[1], nil
}

func splitEscaped(s string) []string {
	var prev rune
	prevIdx := 0
	var out []string
	for i, c := range s {
		if c == ':' && i > 0 && prev != 0 {
			out = append(out, s[prevIdx:i])
		}
	}
	return out
}

// objectOverrideMap converts oos, a slice of object overlays, into a map of the same overlays where the key is the
// object manifest.Hash.
func objectOverrideMap(oos []*v1alpha1.K8SObjectOverlay, namespace string) (map[string][]*v1alpha1.K8SObjectOverlay_PathValue, error) {
	ret := make(map[string][]*v1alpha1.K8SObjectOverlay_PathValue)
	for _, o := range oos {
		ret[manifest.Hash(o.Kind, namespace, o.Name)] = o.Patches
	}
	return ret, nil
}

func stringsEqual(a, b interface{}) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}

// matchesRegex reports whether str regex matches pattern.
func matchesRegex(pattern, str interface{}) bool {
	match, err := regexp.MatchString(fmt.Sprint(pattern), fmt.Sprint(str))
	if err != nil {
		log.Errorf("bad regex expression %s", fmt.Sprint(pattern))
		return false
	}
	dbgPrint("%v regex %v? %v\n", pattern, str, match)
	return match
}

// isSliceOrPtrInterface reports whether v is a slice, a ptr to slice or interface to slice.
func isSliceOrPtrInterface(v interface{}) bool {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}
	if vv.Kind() == reflect.Interface {
		vv = vv.Elem()
	}
	return vv.Kind() == reflect.Slice
}

// isMapOrInterface reports whether v is a map, or interface to a map.
func isMapOrInterface(v interface{}) bool {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.Interface {
		vv = vv.Elem()
	}
	return vv.Kind() == reflect.Map
}

// dbgPrint prints v if the package global variable debugPackage is set.
// v has the same format as Printf. A trailing newline is added to the output.
func dbgPrint(v ...interface{}) {
	if !debugPackage {
		return
	}
	fmt.Println(fmt.Sprintf(v[0].(string), v[1:]...))
}
