package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/kr/pretty"
	"github.com/kylelemons/godebug/diff"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	// debugPackage controls verbose debugging in this package. Used for offline debugging.
	debugPackage = false

	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

// RandomString returns a random string of length n.
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func PrettyJSON(b []byte) []byte {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	if err != nil {
		return []byte(fmt.Sprint(err))
	}
	return out.Bytes()
}

// IsValueNil returns true if either value is nil, or has dynamic type {ptr,
// map, slice} with value nil.
func IsValueNil(value interface{}) bool {
	if value == nil {
		return true
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice, reflect.Ptr, reflect.Map:
		return reflect.ValueOf(value).IsNil()
	}
	return false
}

// IsEmptyString returns true if value is an empty string.
func IsEmptyString(value interface{}) bool {
	if value == nil {
		return true
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.String:
		return value.(string) == ""
	}
	return false
}

func IsString(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.String
}

// IsSlice reports whether value is a slice type.
func IsSlice(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Slice
}

// IsSlicePtr reports whether v is a slice ptr type.
func IsSlicePtr(v interface{}) bool {
	t := reflect.TypeOf(v)
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Slice
}

// IsSliceInterfacePtr reports whether v is a slice ptr type.
func IsSliceInterfacePtr(v interface{}) bool {
	// Must use ValueOf because Elem().Elem() type resolves dynamically.
	vv := reflect.ValueOf(v)
	return vv.Kind() == reflect.Ptr && vv.Elem().Kind() == reflect.Interface && vv.Elem().Elem().Kind() == reflect.Slice
}

// IsPtr reports whether value is a ptr type.
func IsPtr(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Ptr
}

// IsInterfacePtr reports whether v is a slice ptr type.
func IsInterfacePtr(v interface{}) bool {
	t := reflect.TypeOf(v)
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Interface
}

// IsMap reports whether value is a map type.
func IsMap(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Map
}

// IsMapPtr reports whether v is a map ptr type.
func IsMapPtr(v interface{}) bool {
	t := reflect.TypeOf(v)
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Map
}

// IsNilOrInvalidValue reports whether v is nil or reflect.Zero.
func IsNilOrInvalidValue(v reflect.Value) bool {
	return !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) || IsValueNil(v.Interface())
}

// AppendToSlicePtr inserts value into parent which must be a slice ptr.
func AppendToSlicePtr(parentSlice interface{}, value interface{}) error {
	dbgPrint("AppendToSlicePtr slice=\n%s\nvalue=\n%v", pretty.Sprint(parentSlice), value)
	pv := reflect.ValueOf(parentSlice)
	v := reflect.ValueOf(value)

	if !IsSliceInterfacePtr(parentSlice) {
		return fmt.Errorf("AppendToSlicePtr parent type is %T, must be *[]interface{}", parentSlice)
	}

	pv.Elem().Set(reflect.Append(pv.Elem(), v))

	return nil
}

func DeleteFromSlicePtr(parentSlice interface{}, index int) error {
	dbgPrint("DeleteFromSlicePtr index=%d, slice=\n%s", index, pretty.Sprint(parentSlice))
	pv := reflect.ValueOf(parentSlice)

	if !IsSliceInterfacePtr(parentSlice) {
		return fmt.Errorf("DeleteFromSlicePtr parent type is %T, must be *[]interface{}", parentSlice)
	}

	pvv := pv.Elem()
	if pvv.Kind() == reflect.Interface {
		pvv = pvv.Elem()
	}

	ns := reflect.AppendSlice(pvv.Slice(0, index), pvv.Slice(index+1, pvv.Len()))
	pv.Elem().Set(ns)

	return nil
}

func UpdateSlicePtr(parentSlice interface{}, index int, value interface{}) error {
	dbgPrint("UpdateSlicePtr parent=\n%s\n, index=%d, value=\n%v", pretty.Sprint(parentSlice), index, value)
	pv := reflect.ValueOf(parentSlice)
	v := reflect.ValueOf(value)

	if !IsSliceInterfacePtr(parentSlice) {
		return fmt.Errorf("UpdateSlicePtr parent type is %T, must be *[]interface{}", parentSlice)
	}

	pvv := pv.Elem()
	if pvv.Kind() == reflect.Interface {
		pv.Elem().Elem().Index(index).Set(v)
		return nil
	}
	pv.Elem().Index(index).Set(v)

	return nil
}

// InsertIntoMap inserts value with key into parent which must be a map, map ptr, or interface to map.
func InsertIntoMap(parentMap interface{}, key interface{}, value interface{}) error {
	dbgPrint("InsertIntoMap key=%v, value=%s, map=\n%s", key, pretty.Sprint(value), pretty.Sprint(parentMap))
	v := reflect.ValueOf(parentMap)
	kv := reflect.ValueOf(key)
	vv := reflect.ValueOf(value)

	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Type().Kind() == reflect.Interface {
		v = v.Elem()
	}

	if v.Type().Kind() != reflect.Map {
		dbgPrint("error %v", v.Type().Kind())
		return fmt.Errorf("InsertIntoMap parent type is %T, must be map", parentMap)
	}

	v.SetMapIndex(kv, vv)

	return nil
}

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

// dbgPrint prints v if the package global variable debugPackage is set.
// v has the same format as Printf. A trailing newline is added to the output.
func dbgPrint(v ...interface{}) {
	if !debugPackage {
		return
	}
	fmt.Println(fmt.Sprintf(v[0].(string), v[1:]...))
}
