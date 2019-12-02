package helm

import "testing"

func TestStripPrefix(t *testing.T) {
	cases := []struct{
		in string
		prefix string
		out string
	}{
		{"foo/bar", "foo", "bar"},
		{"a/b/c/d", "a/b/c", "d"},
		{"a/b/c/d", "a/b", "c/d"},
	}
	for _, tt := range cases{
		t.Run(tt.in+"~"+tt.prefix, func(t *testing.T) {
			got := stripPrefix(tt.in, tt.prefix)
			if got != tt.out {
				t.Fatalf("Wanted %v got %v", tt.out, got)
			}
		})
	}
}
