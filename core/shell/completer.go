package shell

import (
	"os"
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
				children = append(children, readline.PcItemDynamic(s.completeRemote(true)))
			case "remote_file":
				children = append(children, readline.PcItemDynamic(s.completeRemote(false)))
			case "local_directory":
				children = append(children, readline.PcItemDynamic(s.completeLocal(true)))
			case "local_file":
				children = append(children, readline.PcItemDynamic(s.completeLocal(false)))
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

// completeRemote returns a callback listing entries of the current remote
// working directory. When onlyDirs is set, files are filtered out. Directory
// names get a trailing backslash so completion can continue into them.
func (s *Shell) completeRemote(onlyDirs bool) func(string) []string {
	return func(string) []string {
		if !s.sm.IsConnected() || s.sm.CurrentShare() == "" {
			return nil
		}
		entries, err := s.sm.List("")
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
				out = append(out, e.Name+"\\")
			} else {
				out = append(out, e.Name)
			}
		}
		sort.Strings(out)
		return out
	}
}

// completeLocal returns a callback listing entries of the local working
// directory. When onlyDirs is set, regular files are filtered out.
func (s *Shell) completeLocal(onlyDirs bool) func(string) []string {
	return func(string) []string {
		entries, err := os.ReadDir(s.localCwd)
		if err != nil {
			return nil
		}
		var out []string
		for _, e := range entries {
			if onlyDirs && !e.IsDir() {
				continue
			}
			if e.IsDir() {
				out = append(out, e.Name()+string(os.PathSeparator))
			} else {
				out = append(out, e.Name())
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
