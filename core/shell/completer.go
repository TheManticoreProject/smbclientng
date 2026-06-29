package shell

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

// newCompleter builds the TAB-completion tree from the registered commands,
// using the readline.PrefixCompleter mechanism. Each command's Autocomplete
// hints attach a dynamic completer that lists the relevant names (remote/local
// directory entries, command names, …).
func (s *Shell) newCompleter() *readline.PrefixCompleter {
	items := make([]readline.PrefixCompleterInterface, 0, len(s.order))
	for _, cmd := range s.order {
		var children []readline.PrefixCompleterInterface
		for _, hint := range cmd.Autocomplete {
			switch hint {
			case "remote_directory":
				children = append(children, newPathCompleter(s.completeRemote(true)))
			case "remote_file":
				children = append(children, newPathCompleter(s.completeRemote(false)))
			case "local_directory":
				children = append(children, newPathCompleter(s.completeLocal(true)))
			case "local_file":
				children = append(children, newPathCompleter(s.completeLocal(false)))
			case "command":
				children = append(children, readline.PcItemDynamic(s.completeCommandNames()))
			case "share":
				children = append(children, readline.PcItemDynamic(s.completeShares()))
			}
		}
		items = append(items, readline.PcItem(cmd.Name, children...))
	}
	return readline.NewPrefixCompleter(items...)
}

// dynamicPathCompleter is a readline dynamic completer specialised for file
// system paths. It differs from a plain PcItemDynamic in a single, crucial way:
// it appends the trailing space readline uses to terminate a token ONLY to
// non-directory candidates. Directory candidates (which the callbacks emit with
// a trailing separator) are left space-free so that, once a directory is
// completed, the next TAB descends into it instead of starting a new argument.
type dynamicPathCompleter struct {
	*readline.PrefixCompleter
}

// newPathCompleter wraps a dynamic completion callback so its directory
// candidates support continued, multi-level completion.
func newPathCompleter(callback readline.DynamicCompleteFunc) *dynamicPathCompleter {
	return &dynamicPathCompleter{readline.PcItemDynamic(callback)}
}

// GetDynamicNames overrides PrefixCompleter.GetDynamicNames. The base
// implementation appends " " to every candidate, which terminates the token and
// prevents descending into a just-completed directory. Here a trailing space is
// added only to entries that are not directories.
func (d *dynamicPathCompleter) GetDynamicNames(line []rune) [][]rune {
	var names [][]rune
	for _, name := range d.Callback(string(line)) {
		if strings.HasSuffix(name, "\\") || strings.HasSuffix(name, "/") {
			names = append(names, []rune(name))
		} else {
			names = append(names, []rune(name+" "))
		}
	}
	return names
}

// dirPrefix returns the leading directory portion of a path word, i.e. the
// literal text up to and including the last of the given separators. It is
// empty when word names an entry in the current directory. The returned prefix
// is kept verbatim so completion candidates carry the exact characters the user
// already typed, which readline requires to match them.
func dirPrefix(word, separators string) string {
	if i := strings.LastIndexAny(word, separators); i >= 0 {
		return word[:i+1]
	}
	return ""
}

// completeRemote returns a callback listing entries of the remote directory the
// user is currently typing into. The directory portion already present in the
// partial path is resolved relative to the working directory, listed, and each
// returned candidate is prefixed with that same portion so completion can
// continue to any depth. When onlyDirs is set, files are filtered out.
// Directory names get a trailing backslash so completion descends into them.
func (s *Shell) completeRemote(onlyDirs bool) func(string) []string {
	return func(line string) []string {
		if !s.sm.IsConnected() || s.sm.CurrentShare() == "" {
			return nil
		}
		// Remote (SMB) paths accept both separators; NormalizeRemotePath
		// treats forward slashes as backslashes.
		prefix := dirPrefix(lastWord(line), "\\/")
		entries, err := s.sm.List(prefix)
		if err != nil {
			return nil
		}
		var out []string
		for _, e := range entries {
			if e.Name == "." || e.Name == ".." {
				continue
			}
			if onlyDirs && !e.IsDir() {
				continue
			}
			if e.IsDir() {
				out = append(out, prefix+e.Name+"\\")
			} else {
				out = append(out, prefix+e.Name)
			}
		}
		sort.Strings(out)
		return out
	}
}

// completeLocal returns a callback listing entries of the local directory the
// user is currently typing into, mirroring completeRemote for the local file
// system. When onlyDirs is set, regular files are filtered out.
func (s *Shell) completeLocal(onlyDirs bool) func(string) []string {
	return func(line string) []string {
		// Local paths split on the host's path separator only; on Unix a
		// backslash is a valid filename character.
		prefix := dirPrefix(lastWord(line), string(os.PathSeparator))
		dir := prefix
		if dir == "" {
			dir = s.localCwd
		} else if !filepath.IsAbs(dir) {
			dir = filepath.Join(s.localCwd, dir)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}
		var out []string
		for _, e := range entries {
			if onlyDirs && !e.IsDir() {
				continue
			}
			if e.IsDir() {
				out = append(out, prefix+e.Name()+string(os.PathSeparator))
			} else {
				out = append(out, prefix+e.Name())
			}
		}
		sort.Strings(out)
		return out
	}
}

// completeShares returns a callback listing the names of the shares served by
// the remote host, enumerated over srvsvc. It returns nothing when not
// connected.
func (s *Shell) completeShares() func(string) []string {
	return func(line string) []string {
		if !s.sm.IsConnected() {
			return nil
		}
		shares, err := s.sm.ListShares()
		if err != nil {
			return nil
		}
		names := make([]string, 0, len(shares))
		for _, sh := range shares {
			names = append(names, sh.Name)
		}
		return caseInsensitiveComplete(lastWord(line), names)
	}
}

// lastWord returns the word currently being completed: the substring of the
// command line after the last run of whitespace.
func lastWord(line string) string {
	if i := strings.LastIndexAny(line, " \t"); i >= 0 {
		return line[i+1:]
	}
	return line
}

// caseInsensitiveComplete returns completion candidates for the partial word
// from names, matched case-insensitively. chzyer/readline filters candidates
// with a case-sensitive prefix check and completes by appending the suffix, so
// each candidate keeps the user's typed casing for the matched prefix and
// carries the name's real-cased remainder (e.g. partial "c" + share "C$" yields
// "c$"; partial "C" yields "C$").
func caseInsensitiveComplete(partial string, names []string) []string {
	lower := strings.ToLower(partial)
	out := make([]string, 0, len(names))
	for _, name := range names {
		if strings.HasPrefix(strings.ToLower(name), lower) {
			out = append(out, partial+name[len(partial):])
		}
	}
	sort.Strings(out)
	return out
}

// completeCommandNames returns a callback listing all command names (used by
// `help`).
func (s *Shell) completeCommandNames() func(string) []string {
	return func(string) []string {
		out := make([]string, 0, len(s.order)+1)
		for _, cmd := range s.order {
			out = append(out, cmd.Name)
		}
		// 'help format' is a built-in topic, not a command.
		out = append(out, "format")
		return out
	}
}
