package commands

import (
	"path"
	"strconv"
	"strings"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"

	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name: "find",
		Description: []string{
			"Search for files in a directory hierarchy.",
			"Syntax: 'find [-name PATTERN] [-iname PATTERN] [-type f|d] [-maxdepth N] [-mindepth N] [-ls] [PATH ...]'",
		},
		Autocomplete: []string{"remote_directory"},
		Handler:      cmdFind,
	})
}

// findOptions holds the parsed filtering criteria for the find command.
type findOptions struct {
	name     string // case-sensitive glob, "" if unset
	iname    string // case-insensitive glob, "" if unset
	fileType string // "f", "d", or "" for any
	maxDepth int    // -1 for unlimited
	minDepth int
	ls       bool
	paths    []string
}

func cmdFind(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}

	opts := findOptions{maxDepth: -1}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-name":
			if i+1 < len(args) {
				opts.name = args[i+1]
				i++
			}
		case "-iname":
			if i+1 < len(args) {
				opts.iname = args[i+1]
				i++
			}
		case "-type":
			if i+1 < len(args) {
				opts.fileType = args[i+1]
				i++
			}
		case "-maxdepth":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					opts.maxDepth = v
				}
				i++
			}
		case "-mindepth":
			if i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					opts.minDepth = v
				}
				i++
			}
		case "-ls":
			opts.ls = true
		default:
			opts.paths = append(opts.paths, args[i])
		}
	}

	if len(opts.paths) == 0 {
		opts.paths = []string{"."}
	}

	for _, p := range opts.paths {
		findWalk(s, s.SM().Resolve(p), 1, &opts)
	}
	return nil
}

// findWalk recursively walks the share-relative directory at the given depth,
// printing entries that match the find options.
func findWalk(s *shell.Shell, dir string, depth int, opts *findOptions) {
	if opts.maxDepth >= 0 && depth > opts.maxDepth {
		return
	}

	entries, err := s.SM().Client().ListDirectory(dir, "*")
	if err != nil {
		s.Errorf("[!] SMB Error: %s", err)
		return
	}
	sortDirsFirst(entries)

	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		full := utils.JoinRemotePath(dir, e.Name)

		if depth >= opts.minDepth && findMatch(e, opts) {
			if opts.ls {
				s.Print(windowsLsEntry(e, utils.ToWirePath(full)))
			} else {
				// Print matches with forward slashes.
				s.Print(strings.ReplaceAll(utils.ToWirePath(full), "\\", "/"))
			}
		}

		if e.IsDir() {
			findWalk(s, full, depth+1, opts)
		}
	}
}

// findMatch reports whether an entry satisfies the type and name filters.
func findMatch(e smbclient.FileInfo, opts *findOptions) bool {
	switch opts.fileType {
	case "f":
		if e.IsDir() {
			return false
		}
	case "d":
		if !e.IsDir() {
			return false
		}
	}

	if opts.name != "" {
		if ok, _ := path.Match(opts.name, e.Name); !ok {
			return false
		}
	}
	if opts.iname != "" {
		if ok, _ := path.Match(strings.ToLower(opts.iname), strings.ToLower(e.Name)); !ok {
			return false
		}
	}
	return true
}
