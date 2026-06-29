package shell

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/chzyer/readline"
)

func TestLastWord(t *testing.T) {
	cases := map[string]string{
		"use c":       "c",
		"use ":        "",
		"use":         "use",
		"use  C$":     "C$",
		"info --shar": "--shar",
	}
	for in, want := range cases {
		if got := lastWord(in); got != want {
			t.Errorf("lastWord(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCaseInsensitiveComplete(t *testing.T) {
	shares := []string{"C$", "ADMIN$", "IPC$", "Users"}

	cases := []struct {
		partial string
		want    []string
	}{
		// Lowercase typing matches case-insensitively; the typed casing is kept
		// for the matched prefix (readline cannot rewrite typed characters).
		{"c", []string{"c$"}},
		// Exact-case typing completes to the real share name.
		{"C", []string{"C$"}},
		{"ad", []string{"adMIN$"}},
		{"ADMIN$", []string{"ADMIN$"}},
		// Empty partial lists every share in its real case.
		{"", []string{"ADMIN$", "C$", "IPC$", "Users"}},
		// No case-insensitive match.
		{"x", []string{}},
		// 'i' matches IPC$ only.
		{"i", []string{"iPC$"}},
	}
	for _, c := range cases {
		got := caseInsensitiveComplete(c.partial, shares)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("caseInsensitiveComplete(%q) = %#v, want %#v", c.partial, got, c.want)
		}
	}
}

func TestDirPrefix(t *testing.T) {
	cases := []struct {
		word string
		seps string
		want string
	}{
		// No separator: the word is a leaf name in the current directory.
		{"sub", "\\/", ""},
		{"", "\\/", ""},
		// The prefix keeps everything up to and including the last separator.
		{"a\\b", "\\/", "a\\"},
		{"a\\b\\c", "\\/", "a\\b\\"},
		{"a/b/c", "\\/", "a/b/"},
		{"a\\b/c", "\\/", "a\\b/"},
		// Trailing separator: completing the contents of that directory.
		{"a\\b\\", "\\/", "a\\b\\"},
		// Local split honours only the requested separator, so a backslash is an
		// ordinary character.
		{"a\\b", "/", ""},
		{"a/b", "/", "a/"},
	}
	for _, c := range cases {
		if got := dirPrefix(c.word, c.seps); got != c.want {
			t.Errorf("dirPrefix(%q, %q) = %q, want %q", c.word, c.seps, got, c.want)
		}
	}
}

func TestDynamicPathCompleterTrailingSpace(t *testing.T) {
	// Directories (trailing separator) must NOT receive a trailing space, so
	// that completing one lets the next TAB descend into it. Files must, so the
	// argument is terminated.
	pc := newPathCompleter(func(string) []string {
		return []string{"dir\\", "dir/", "file.txt"}
	})
	got := pc.GetDynamicNames(nil)
	want := [][]rune{[]rune("dir\\"), []rune("dir/"), []rune("file.txt ")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetDynamicNames() = %v, want %v", got, want)
	}
}

func TestCompleteLocalMultiLevel(t *testing.T) {
	root := t.TempDir()
	// root/
	//   alpha/
	//     beta/
	//       leaf.txt
	//     note.txt
	mustMkdir(t, filepath.Join(root, "alpha", "beta"))
	mustWrite(t, filepath.Join(root, "alpha", "note.txt"))
	mustWrite(t, filepath.Join(root, "alpha", "beta", "leaf.txt"))

	s := &Shell{localCwd: root}
	sep := string(os.PathSeparator)

	// Top level lists entries of the working directory.
	if got := s.completeLocal(true)("lcd "); !reflect.DeepEqual(sorted(got), []string{"alpha" + sep}) {
		t.Fatalf("top-level dir completion = %v, want [alpha%s]", got, sep)
	}

	// One level deep: completing "alpha/" must list alpha's entries, each
	// prefixed with the typed directory portion — this is the behaviour that
	// was previously broken.
	want := sorted([]string{"alpha" + sep + "beta" + sep, "alpha" + sep + "note.txt"})
	if got := sorted(s.completeLocal(false)("get alpha" + sep)); !reflect.DeepEqual(got, want) {
		t.Fatalf("one-level completion = %v, want %v", got, want)
	}

	// Two levels deep: completion continues into nested directories.
	wantDeep := []string{"alpha" + sep + "beta" + sep + "leaf.txt"}
	if got := s.completeLocal(false)("get alpha" + sep + "beta" + sep); !reflect.DeepEqual(got, wantDeep) {
		t.Fatalf("two-level completion = %v, want %v", got, wantDeep)
	}

	// onlyDirs filters files out at depth.
	wantDirs := []string{"alpha" + sep + "beta" + sep}
	if got := s.completeLocal(true)("lcd alpha" + sep); !reflect.DeepEqual(got, wantDirs) {
		t.Fatalf("one-level dir-only completion = %v, want %v", got, wantDirs)
	}
}

// TestCompleterEngineMultiLevel drives the real chzyer/readline completion
// engine (readline.Do) end to end to prove that a directory completed at one
// level can be descended into on the next TAB — the behaviour the dynamic-path
// completer restores.
func TestCompleterEngineMultiLevel(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "alpha", "beta", "gamma"))

	s := &Shell{localCwd: root}
	tree := readline.NewPrefixCompleter(
		readline.PcItem("lcd", newPathCompleter(s.completeLocal(true))),
	)
	sep := string(os.PathSeparator)

	// complete returns the single suffix readline would insert, requiring the
	// completion to be unambiguous at each step.
	complete := func(line string) string {
		cands, _ := tree.Do([]rune(line), len(line))
		if len(cands) != 1 {
			t.Fatalf("completing %q: expected 1 candidate, got %d (%v)", line, len(cands), runesToStrings(cands))
		}
		return string(cands[0])
	}

	// "lcd a" -> completes the only top-level dir to "alpha\" (no trailing space).
	if got := "lcd a" + complete("lcd a"); got != "lcd alpha"+sep {
		t.Fatalf("first level: got %q, want %q", got, "lcd alpha"+sep)
	}
	// "lcd alpha\" -> descends, completing the only child to "beta\".
	if got := complete("lcd alpha" + sep); got != "beta"+sep {
		t.Fatalf("second level: got %q, want %q", got, "beta"+sep)
	}
	// "lcd alpha\beta\" -> the engine keeps descending to arbitrary depth.
	if cands, _ := tree.Do([]rune("lcd alpha"+sep+"beta"+sep), len("lcd alpha"+sep+"beta"+sep)); len(cands) == 0 {
		t.Fatalf("third level: expected a candidate inside beta, got none")
	}
}

func runesToStrings(in [][]rune) []string {
	out := make([]string, len(in))
	for i, r := range in {
		out[i] = string(r)
	}
	return out
}

func sorted(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}
