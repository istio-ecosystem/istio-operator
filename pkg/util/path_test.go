package util

import (
	"testing"
)

func TestSplitEscaped(t *testing.T) {
	tests := []struct {
		desc string
		in   string
		want []string
	}{
		{
			desc: "empty",
			in:   "",
			want: []string{},
		},
		{
			desc: "no match",
			in:   "foo",
			want: []string{"foo"},
		},
		{
			desc: "first",
			in:   ":foo",
			want: []string{"", "foo"},
		},
		{
			desc: "last",
			in:   "foo:",
			want: []string{"foo", ""},
		},
		{
			desc: "multiple",
			in:   "foo:bar:baz",
			want: []string{"foo", "bar", "baz"},
		},
		{
			desc: "multiple with escapes",
			in:   `foo\:bar:baz\:qux`,
			want: []string{`foo\:bar`, `baz\:qux`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got, want := splitEscaped(tt.in, kvSeparatorRune), tt.want; !stringSlicesEqual(got, want) {
				t.Errorf("%s: got:%v, want:%v", tt.desc, got, want)
			}
		})
	}
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, aa := range a {
		if aa != b[i] {
			return false
		}
	}
	return true
}
