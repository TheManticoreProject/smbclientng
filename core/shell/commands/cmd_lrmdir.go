package commands

import (
	"os"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "lrmdir",
		Description:  []string{"Removes a local directory.", "Syntax: 'lrmdir <directory>'"},
		Autocomplete: []string{"local_directory"},
		Handler:      cmdLrmdir,
	})
}

func cmdLrmdir(s *shell.Shell, args []string) error {
	if len(args) < 1 {
		s.Errorf("Syntax: 'lrmdir <directory>'")
		return nil
	}

	for _, arg := range args {
		target := localPath(s, arg)
		info, err := os.Stat(target)
		if err != nil {
			s.Errorf("Path '%s' does not exist.", arg)
			continue
		}
		if !info.IsDir() {
			s.Errorf("Cannot delete '%s'. It is a file, use 'lrm <file>' instead.", arg)
			continue
		}
		if err := os.RemoveAll(target); err != nil {
			s.Errorf("Error removing directory '%s': %s", arg, err)
		}
	}
	return nil
}
