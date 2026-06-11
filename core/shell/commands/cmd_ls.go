package commands

import (
	"sort"
	"strings"

	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "ls",
		Aliases:      []string{"dir"},
		Description:  []string{"List the contents of the current remote working directory.", "Syntax: 'ls [directory]'"},
		Autocomplete: []string{"remote_directory"},
		Handler:      cmdLs,
	})
}

func cmdLs(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}

	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}

	for _, path := range paths {
		if len(paths) > 1 {
			s.Printf("%s:", path)
		}

		entries, err := s.SM().List(path)
		if err != nil {
			s.Errorf("[!] SMB Error: %s", err)
			continue
		}

		// Sort case-insensitively by name.
		sort.Slice(entries, func(i, j int) bool {
			return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		})

		for _, e := range entries {
			if e.Name == "." || e.Name == ".." {
				continue
			}
			s.Print(windowsLsEntry(e, e.Name))
		}

		if len(paths) > 1 {
			s.Print("")
		}
	}
	return nil
}

// listSorted is a small helper shared by other commands that need a
// name-sorted directory listing of a remote path.
func listSorted(entries []smbclient.FileInfo) []smbclient.FileInfo {
	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries
}

// sortDirsFirst orders entries directories-first, then case-insensitively by
// name. Used by tree and find.
func sortDirsFirst(entries []smbclient.FileInfo) []smbclient.FileInfo {
	sort.Slice(entries, func(i, j int) bool {
		di, dj := entries[i].IsDir(), entries[j].IsDir()
		if di != dj {
			return di
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries
}
