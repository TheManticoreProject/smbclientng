package commands

import (
	"os"
	"path/filepath"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "get",
		Description:  []string{"Get a remote file.", "Syntax: 'get <file> [local_file]'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdGet,
	})
}

// localPath resolves a user-supplied local path against the shell's local cwd.
func localPath(s *shell.Shell, arg string) string {
	if filepath.IsAbs(arg) {
		return arg
	}
	return filepath.Join(s.LocalCwd(), arg)
}

func cmdGet(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) < 1 || len(args) > 2 {
		s.Errorf("Syntax: 'get <file> [local_file]'")
		return nil
	}

	remote := s.SM().Resolve(args[0])
	localName := utils.RemoteBase(remote)
	if len(args) == 2 {
		localName = args[1]
	}
	if localName == "" {
		s.Errorf("Could not determine a local filename for '%s'.", args[0])
		return nil
	}
	localTarget := localPath(s, localName)

	f, err := os.Create(localTarget)
	if err != nil {
		s.Errorf("Cannot create local file '%s': %s", localTarget, err)
		return nil
	}
	defer f.Close()

	n, err := s.SM().DownloadTo(remote, f)
	if err != nil {
		s.Errorf("Download failed after %d bytes: %s", n, err)
		return nil
	}
	s.Infof("Downloaded '%s' (%s) to '%s'.", utils.ToWirePath(remote), filesize(uint64(n)), localTarget)
	return nil
}
