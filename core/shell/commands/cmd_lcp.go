package commands

import (
	"io"
	"os"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "lcp",
		Description:  []string{"Create a copy of a local file.", "Syntax: 'lcp <srcfile> <dstfile>'"},
		Autocomplete: []string{"local_file"},
		Handler:      cmdLcp,
	})
}

func cmdLcp(s *shell.Shell, args []string) error {
	if len(args) != 2 {
		s.Errorf("Syntax: 'lcp <srcfile> <dstfile>'")
		return nil
	}

	src := localPath(s, args[0])
	dst := localPath(s, args[1])

	in, err := os.Open(src)
	if err != nil {
		s.Errorf("[!] File '%s' does not exists.", args[0])
		return nil
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		s.Errorf("Error creating '%s': %s", args[1], err)
		return nil
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		s.Errorf("Error copying file: %s", err)
	}
	return nil
}
