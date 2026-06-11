package commands

import (
	"fmt"
	"sort"
	"strings"

	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
	"github.com/TheManticoreProject/smbclient-ng/core/shell"
	"github.com/TheManticoreProject/smbclient-ng/core/utils"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "metadata",
		Description:  []string{"Get metadata about a file or directory.", "Syntax: 'metadata <file|directory>'"},
		Autocomplete: []string{"remote_file"},
		Handler:      cmdMetadata,
	})
}

func cmdMetadata(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}

	masks := args
	if len(masks) == 0 {
		masks = []string{"*"}
	}

	for _, mask := range masks {
		entries, err := s.SM().ListPattern(mask)
		if err != nil {
			s.Errorf("Failed to access '%s': %s", mask, err)
			continue
		}
		baseDir := utils.RemoteDir(s.SM().Resolve(mask))
		matched := false
		for _, e := range entries {
			if e.Name == "." || e.Name == ".." {
				continue
			}
			matched = true
			printMetadata(s, e, utils.JoinRemotePath(baseDir, e.Name))
		}
		if !matched {
			s.Errorf("No such file or directory: '%s'", mask)
		}
	}
	return nil
}

// printMetadata renders the metadata tree for a single entry. Alternate data
// streams and security descriptors are omitted: the Manticore SMB client does
// not expose the queries needed to retrieve them yet.
func printMetadata(s *shell.Shell, entry smbclient.FileInfo, relPath string) {
	unc := fmt.Sprintf("\\\\%s\\%s\\%s", s.SM().Host, s.SM().CurrentShare(), strings.ReplaceAll(relPath, "/", "\\"))

	label := func(width int, name string) string {
		return fmt.Sprintf("\x1b[94m%-*s\x1b[0m", width, name)
	}
	value := func(v string) string { return fmt.Sprintf("\x1b[93m%s\x1b[0m", v) }

	s.Print(fmt.Sprintf("[+] Metadata of '%s'", unc))
	s.Print(fmt.Sprintf("  ├─ %s: %s", label(4, "Name"), value(entry.Name)))
	s.Print(fmt.Sprintf("  ├─ %s: %s", label(4, "Path"), value(unc)))

	s.Print("  ├─ [+] General information")
	if entry.IsDir() {
		s.Print(fmt.Sprintf("  │    ├─ %s: %s", label(10, "Type"), value("📁 Directory")))
	} else {
		s.Print(fmt.Sprintf("  │    ├─ %s: %s", label(10, "Type"), value("📄 File")))
		s.Print(fmt.Sprintf("  │    ├─ %s: %s", label(10, "Size"), value(fmt.Sprintf("%s (%d bytes)", filesize(entry.Size), entry.Size))))
	}
	s.Print(fmt.Sprintf("  │    ├─ %s: %s (%s)", label(10, "Attributes"),
		value(fmt.Sprintf("%d", entry.FileAttributes)), value(attributeNames(entry.FileAttributes))))
	s.Print("  │    └───")

	s.Print("  ├─ [+] Timestamps")
	const ts = "2006-01-02 15:04:05"
	s.Print(fmt.Sprintf("  │    ├─ %s: %s", label(10, "Created"), value(entry.CreationTime.Format(ts))))
	s.Print(fmt.Sprintf("  │    ├─ %s: %s", label(10, "Accessed"), value(entry.LastAccessTime.Format(ts))))
	s.Print(fmt.Sprintf("  │    ├─ %s: %s", label(10, "Modified"), value(entry.LastWriteTime.Format(ts))))
	s.Print(fmt.Sprintf("  │    └─ %s: %s", label(10, "Changed"), value(entry.ChangeTime.Format(ts))))
	s.Print("  └───")
	s.Print("")
}

// attributeNames decodes the FILE_ATTRIBUTE_* bitmask into a sorted,
// comma-separated list of human-readable names.
func attributeNames(attrs uint32) string {
	bits := []struct {
		bit  uint32
		name string
	}{
		{attrDirectory, "Directory"},
		{attrArchive, "Archive"},
		{attrCompressed, "Compressed"},
		{attrHidden, "Hidden"},
		{attrNormal, "Normal"},
		{attrReadonly, "ReadOnly"},
		{attrSystem, "System"},
		{attrTemporary, "Temporary"},
	}
	var names []string
	for _, b := range bits {
		if attrs&b.bit != 0 {
			names = append(names, b.name)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "None"
	}
	return strings.Join(names, ", ")
}
