package commands

import (
	"os"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "mget",
		Description:  []string{"Download every remote file matching a wildcard mask.", "Syntax: 'mget <mask>'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdMget,
	})
}

func cmdMget(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) != 1 {
		s.Errorf("Syntax: 'mget <mask>'")
		return nil
	}

	entries, err := s.SM().ListPattern(args[0])
	if err != nil {
		s.Errorf("[!] SMB Error: %s", err)
		return nil
	}

	matched := 0
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." || e.IsDir() {
			continue
		}
		// Matches resolve relative to the mask's directory, so a mask like
		// "sub\*.txt" downloads from "sub".
		remote := utils.JoinRemotePath(utils.RemoteDir(s.SM().Resolve(args[0])), e.Name)
		localTarget := localPath(s, e.Name)

		f, ferr := os.Create(localTarget)
		if ferr != nil {
			s.Errorf("Skipping '%s': %s", e.Name, ferr)
			continue
		}
		n, derr := s.SM().DownloadTo(remote, f)
		f.Close()
		if derr != nil {
			s.Errorf("Failed to download '%s' after %d bytes: %s", e.Name, n, derr)
			continue
		}
		s.Infof("Downloaded '%s' (%s).", e.Name, filesize(uint64(n)))
		matched++
	}

	if matched == 0 {
		s.Info("No files matched.")
	}
	return nil
}
