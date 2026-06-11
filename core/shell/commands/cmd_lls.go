package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "lls",
		Description:  []string{"Lists the contents of the current local directory.", "Syntax: 'lls [directory]'"},
		Autocomplete: []string{"local_directory"},
		Handler:      cmdLls,
	})
}

func cmdLls(s *shell.Shell, args []string) error {
	target := s.LocalCwd()
	if len(args) == 1 {
		if filepath.IsAbs(args[0]) {
			target = args[0]
		} else {
			target = filepath.Join(s.LocalCwd(), args[0])
		}
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		s.Errorf("%s", err)
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		meta := unixPermissions(info)
		date := info.ModTime().Format("2006-01-02 15:04")
		if e.IsDir() {
			s.Print(fmt.Sprintf("%s %10s  %s  %s%s%s%c", meta, filesize(uint64(info.Size())), date, shell.ColorBoldCyan, e.Name(), shell.ColorReset, os.PathSeparator))
		} else {
			s.Print(fmt.Sprintf("%s %10s  %s  %s%s%s", meta, filesize(uint64(info.Size())), date, shell.ColorBold, e.Name(), shell.ColorReset))
		}
	}
	return nil
}

// unixPermissions renders a 10-character Unix-style permission string
// (e.g. "drwxr-xr-x") for a local entry.
func unixPermissions(info os.FileInfo) string {
	mode := info.Mode()
	buf := make([]byte, 0, 10)
	if mode.IsDir() {
		buf = append(buf, 'd')
	} else {
		buf = append(buf, '-')
	}
	const rwx = "rwxrwxrwx"
	perm := mode.Perm()
	for i := 0; i < 9; i++ {
		if perm&(1<<uint(8-i)) != 0 {
			buf = append(buf, rwx[i])
		} else {
			buf = append(buf, '-')
		}
	}
	return string(buf)
}
