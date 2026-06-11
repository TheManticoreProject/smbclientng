package shell

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   \n", nil},
		{"ls", []string{"ls"}},
		{"cd some/dir\n", []string{"cd", "some/dir"}},
		{"get \"file with spaces.txt\"", []string{"get", "file with spaces.txt"}},
		{"  use   C$  ", []string{"use", "C$"}},
		{"put \"\"", []string{"put", ""}},
	}
	for _, c := range cases {
		got := tokenize(c.in)
		if len(got) == 0 && len(c.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("tokenize(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}
