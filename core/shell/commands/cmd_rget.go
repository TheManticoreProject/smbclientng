package commands

import (
	"os"
	"path/filepath"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "rget",
		Description:  []string{"Recursively download a remote directory tree.", "Syntax: 'rget [directory]'"},
		Autocomplete: []string{"remote_directory"},
		Handler:      cmdRget,
	})
}

func cmdRget(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) > 1 {
		s.Errorf("Syntax: 'rget [directory]'")
		return nil
	}

	root := s.SM().Cwd()
	if len(args) == 1 {
		root = s.SM().Resolve(args[0])
	}
	count, err := rgetWalk(s, root, s.LocalCwd())
	if err != nil {
		s.Errorf("[!] SMB Error: %s", err)
		return nil
	}
	s.Infof("Downloaded %d file(s).", count)
	return nil
}

// rgetWalk recursively downloads the remote directory at remotePath into
// localDir, recreating the directory structure, and returns the file count.
func rgetWalk(s *shell.Shell, remotePath, localDir string) (int, error) {
	entries, err := s.SM().Client().ListDirectory(remotePath, "*")
	if err != nil {
		return 0, err
	}

	if err := os.MkdirAll(localDir, 0o755); err != nil {
		return 0, err
	}

	count := 0
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		childRemote := utils.JoinRemotePath(remotePath, e.Name)
		childLocal := filepath.Join(localDir, e.Name)

		if e.IsDir() {
			sub, werr := rgetWalk(s, childRemote, childLocal)
			if werr != nil {
				s.Errorf("%s: %s", utils.ToWirePath(childRemote), werr)
				continue
			}
			count += sub
			continue
		}

		f, ferr := os.Create(childLocal)
		if ferr != nil {
			s.Errorf("Skipping '%s': %s", childLocal, ferr)
			continue
		}
		n, derr := s.SM().DownloadTo(childRemote, f)
		f.Close()
		if derr != nil {
			s.Errorf("Failed to download '%s' after %d bytes: %s", utils.ToWirePath(childRemote), n, derr)
			continue
		}
		s.Infof("Downloaded '%s' (%s).", utils.ToWirePath(childRemote), filesize(uint64(n)))
		count++
	}
	return count, nil
}
