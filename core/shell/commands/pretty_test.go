package commands

import (
	"reflect"
	"testing"

	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
)

func TestLimitLines(t *testing.T) {
	const text = "l1\nl2\nl3\nl4\nl5"
	cases := []struct {
		name string
		n    int
		tail bool
		want string
	}{
		{"head-3", 3, false, "l1\nl2\nl3"},
		{"tail-2", 2, true, "l4\nl5"},
		{"head-all", 10, false, text},
		{"tail-all", 10, true, text},
		{"zero", 0, false, text},
	}
	for _, c := range cases {
		if got := limitLines(text, c.n, c.tail); got != c.want {
			t.Errorf("%s: limitLines = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestParseLinesFlag(t *testing.T) {
	cases := []struct {
		in       []string
		wantN    int
		wantRest []string
	}{
		{[]string{"file.txt"}, 10, []string{"file.txt"}},
		{[]string{"-n", "5", "file.txt"}, 5, []string{"file.txt"}},
		{[]string{"--lines", "3", "a", "b"}, 3, []string{"a", "b"}},
		{[]string{"-n", "notanumber", "f"}, 10, []string{"f"}},
	}
	for _, c := range cases {
		n, rest := parseLinesFlag(c.in)
		if n != c.wantN || !reflect.DeepEqual(rest, c.wantRest) {
			t.Errorf("parseLinesFlag(%v) = (%d, %v), want (%d, %v)", c.in, n, rest, c.wantN, c.wantRest)
		}
	}
}

func TestFindMatch(t *testing.T) {
	dir := smbclient.FileInfo{Name: "Logs", FileAttributes: attrDirectory}
	file := smbclient.FileInfo{Name: "Report.TXT"}

	cases := []struct {
		name  string
		entry smbclient.FileInfo
		opts  findOptions
		want  bool
	}{
		{"type-f-on-dir", dir, findOptions{fileType: "f"}, false},
		{"type-d-on-dir", dir, findOptions{fileType: "d"}, true},
		{"name-mismatch-case", file, findOptions{name: "*.txt"}, false},
		{"iname-match", file, findOptions{iname: "*.txt"}, true},
		{"name-exact", file, findOptions{name: "Report.TXT"}, true},
		{"no-filter", file, findOptions{}, true},
	}
	for _, c := range cases {
		if got := findMatch(c.entry, &c.opts); got != c.want {
			t.Errorf("%s: findMatch = %v, want %v", c.name, got, c.want)
		}
	}
}
