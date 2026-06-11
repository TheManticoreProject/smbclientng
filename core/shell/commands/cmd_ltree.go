package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "ltree",
		Description:  []string{"Displays a tree view of the local directories.", "Syntax: 'ltree [directory]'"},
		Autocomplete: []string{"local_directory"},
		Handler:      cmdLtree,
	})
}

func cmdLtree(s *shell.Shell, args []string) error {
	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}
	for _, arg := range paths {
		root := localPath(s, arg)
		s.Print(fmt.Sprintf("%s%s%s%c", shell.ColorBoldCyan, arg, shell.ColorReset, os.PathSeparator))
		localTreeWalk(s, root, "")
	}
	return nil
}

// localTreeWalk recursively prints a local directory tree using the same
// ├──/└── connectors and cyan directory styling as the remote `tree` command.
func localTreeWalk(s *shell.Shell, path, prefix string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		s.Print(fmt.Sprintf("%s└── \x1b[1;91m%s\x1b[0m", prefix, err))
		return
	}

	// Sort case-sensitively (uppercase before lowercase).
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for i, e := range entries {
		last := i == len(entries)-1
		connector, childPrefix := "├── ", prefix+"│   "
		if last {
			connector, childPrefix = "└── ", prefix+"    "
		}

		if e.IsDir() {
			s.Print(fmt.Sprintf("%s%s%s%s%s%c", prefix, connector, shell.ColorBoldCyan, e.Name(), shell.ColorReset, os.PathSeparator))
			localTreeWalk(s, filepath.Join(path, e.Name()), childPrefix)
		} else {
			s.Print(fmt.Sprintf("%s%s%s%s%s", prefix, connector, shell.ColorBold, e.Name(), shell.ColorReset))
		}
	}
}
