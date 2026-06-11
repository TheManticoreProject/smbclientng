package commands

import (
	"os"
	"path/filepath"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "lmkdir",
		Description:  []string{"Creates a new local directory.", "Syntax: 'lmkdir <directory>'"},
		Autocomplete: []string{"local_directory"},
		Handler:      cmdLmkdir,
	})
}

func cmdLmkdir(s *shell.Shell, args []string) error {
	if len(args) < 1 {
		s.Errorf("Syntax: 'lmkdir <directory>'")
		return nil
	}

	for _, arg := range args {
		target := arg
		if !filepath.IsAbs(target) {
			target = filepath.Join(s.LocalCwd(), target)
		}
		if err := os.MkdirAll(target, 0o755); err != nil {
			s.Errorf("Error creating directory '%s': %s", arg, err)
		}
	}
	return nil
}
