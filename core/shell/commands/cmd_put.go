package commands

import (
	"os"
	"path/filepath"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "put",
		Description:  []string{"Put a local file or directory in a remote directory.", "Syntax: 'put <local_file> [remote_file]'"},
		Autocomplete: []string{"local_file"},
		Handler:      cmdPut,
	})
}

func cmdPut(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) < 1 || len(args) > 2 {
		s.Errorf("Syntax: 'put <local_file> [remote_file]'")
		return nil
	}

	localTarget := localPath(s, args[0])
	remoteName := filepath.Base(args[0])
	if len(args) == 2 {
		remoteName = args[1]
	}
	remote := s.SM().Resolve(remoteName)

	f, err := os.Open(localTarget)
	if err != nil {
		s.Errorf("Cannot open local file '%s': %s", localTarget, err)
		return nil
	}
	defer f.Close()

	n, err := s.SM().UploadFrom(remote, f)
	if err != nil {
		s.Errorf("Upload failed after %d bytes: %s", n, err)
		return nil
	}
	s.Infof("Uploaded '%s' (%s) to '%s'.", localTarget, filesize(uint64(n)), utils.ToWirePath(remote))
	return nil
}
