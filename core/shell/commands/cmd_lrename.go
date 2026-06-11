package commands

import (
	"os"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "lrename",
		Description:  []string{"Renames a local file.", "Syntax: 'lrename <oldfilename> <newfilename>'"},
		Autocomplete: []string{"local_file"},
		Handler:      cmdLrename,
	})
}

func cmdLrename(s *shell.Shell, args []string) error {
	if len(args) != 2 {
		s.Errorf("Syntax: 'lrename <oldfilename> <newfilename>'")
		return nil
	}
	if err := os.Rename(localPath(s, args[0]), localPath(s, args[1])); err != nil {
		s.Errorf("Error renaming '%s' to '%s': %s", args[0], args[1], err)
	}
	return nil
}
