package commands

import (
	"fmt"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "tree",
		Description:  []string{"Displays a tree view of the remote directories.", "Syntax: 'tree [directory]'"},
		Autocomplete: []string{"remote_directory"},
		Handler:      cmdTree,
	})
}

func cmdTree(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}

	root := s.SM().Cwd()
	if len(args) >= 1 {
		root = s.SM().Resolve(args[0])
	}

	treeWalk(s, root, "")
	return nil
}

// treeWalk recursively prints a directory tree rooted at the share-relative
// path, using the ├──/└── connectors and cyan directory names.
func treeWalk(s *shell.Shell, path, prefix string) {
	entries, err := s.SM().Client().ListDirectory(path, "*")
	if err != nil {
		s.Print(fmt.Sprintf("%s%s[error: %s]%s", prefix, shell.ColorGrey, err, shell.ColorReset))
		return
	}

	visible := entries[:0]
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		visible = append(visible, e)
	}
	sortDirsFirst(visible)

	for i, e := range visible {
		last := i == len(visible)-1
		connector, childPrefix := "├── ", prefix+"│   "
		if last {
			connector, childPrefix = "└── ", prefix+"    "
		}

		if e.IsDir() {
			s.Print(fmt.Sprintf("%s%s%s%s%s/", prefix, connector, shell.ColorBoldCyan, e.Name, shell.ColorReset))
			treeWalk(s, utils.JoinRemotePath(path, e.Name), childPrefix)
		} else {
			s.Print(fmt.Sprintf("%s%s%s", prefix, connector, e.Name))
		}
	}
}
