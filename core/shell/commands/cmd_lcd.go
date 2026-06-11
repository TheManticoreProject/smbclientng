package commands

import (
	"os"
	"path/filepath"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "lcd",
		Description:  []string{"Changes the current local directory.", "Syntax: 'lcd <directory>'"},
		Autocomplete: []string{"local_directory"},
		Handler:      cmdLcd,
	})
}

func cmdLcd(s *shell.Shell, args []string) error {
	if len(args) != 1 {
		s.Errorf("Syntax: 'lcd <directory>'")
		return nil
	}
	target := args[0]
	if !filepath.IsAbs(target) {
		target = filepath.Join(s.LocalCwd(), target)
	}
	info, err := os.Stat(target)
	if err != nil {
		s.Errorf("Directory '%s' does not exists.", args[0])
		return nil
	}
	if !info.IsDir() {
		s.Errorf("Path '%s' is not a directory.", args[0])
		return nil
	}
	s.SetLocalCwd(target)
	return nil
}
