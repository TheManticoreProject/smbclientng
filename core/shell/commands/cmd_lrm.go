package commands

import (
	"os"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "lrm",
		Description:  []string{"Removes a local file.", "Syntax: 'lrm <file>'"},
		Autocomplete: []string{"local_file"},
		Handler:      cmdLrm,
	})
}

func cmdLrm(s *shell.Shell, args []string) error {
	if len(args) < 1 {
		s.Errorf("Syntax: 'lrm <file>'")
		return nil
	}

	for _, arg := range args {
		target := localPath(s, arg)
		info, err := os.Stat(target)
		if err != nil {
			s.Errorf("Path '%s' does not exist.", arg)
			continue
		}
		if info.IsDir() {
			s.Errorf("Cannot delete '%s'. It is a directory, use 'lrmdir <directory>' instead.", arg)
			continue
		}
		if err := os.Remove(target); err != nil {
			s.Errorf("Error removing file '%s': %s", arg, err)
		}
	}
	return nil
}
