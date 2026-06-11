package commands

import (
	"fmt"
	"os"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "lbat",
		Description:  []string{"Pretty prints the contents of a local file.", "Syntax: 'lbat <file>'"},
		Autocomplete: []string{"local_file"},
		Handler:      cmdLbat,
	})
}

func cmdLbat(s *shell.Shell, args []string) error {
	if len(args) < 1 {
		s.Errorf("Syntax: 'lbat <file>'")
		return nil
	}

	for _, arg := range args {
		content, err := os.ReadFile(localPath(s, arg))
		if err != nil {
			s.Errorf("[!] Local file '%s' does not exist.", arg)
			continue
		}
		if len(args) > 1 {
			s.Print(fmt.Sprintf("\x1b[1;93m[>] %s\x1b[0m", padHeader(arg)))
		}
		printNumbered(s, string(content))
	}
	return nil
}
