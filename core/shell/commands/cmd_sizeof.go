package commands

import (
	"fmt"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "sizeof",
		Description:  []string{"Recursively compute the size of a folder.", "Syntax: 'sizeof [directory|file]'"},
		Autocomplete: []string{"remote_directory"},
		Handler:      cmdSizeof,
	})
}

func cmdSizeof(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}

	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var total uint64
	for _, arg := range paths {
		target := s.SM().Resolve(arg)
		size, err := dirSize(s, target)
		if err != nil {
			s.Errorf("Failed to access '%s': %s", utils.ToWirePath(target), err)
			continue
		}
		total += size
		s.Print(fmt.Sprintf("%s\t%s%s%s", filesize(size), shell.ColorBoldCyan, utils.ToWirePath(target), shell.ColorReset))
	}

	if len(paths) > 1 {
		s.Print("──────────────────────")
		s.Print(fmt.Sprintf("Total size: %s", filesize(total)))
	}
	return nil
}

// dirSize recursively sums the size of every file under the share-relative path.
// If the path is a single file, its own size is returned.
func dirSize(s *shell.Shell, path string) (uint64, error) {
	entries, err := s.SM().Client().ListDirectory(path, "*")
	if err != nil {
		return 0, err
	}

	var total uint64
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		if e.IsDir() {
			sub, err := dirSize(s, utils.JoinRemotePath(path, e.Name))
			if err != nil {
				return total, err
			}
			total += sub
		} else {
			total += e.Size
		}
	}
	return total, nil
}
