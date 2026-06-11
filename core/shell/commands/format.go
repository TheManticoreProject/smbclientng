package commands

import (
	"fmt"

	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

// Windows file attribute bits used to render the ls metadata column. These match
// [MS-FSCC] 2.6 File Attributes.
const (
	attrReadonly   = 0x00000001
	attrHidden     = 0x00000002
	attrSystem     = 0x00000004
	attrDirectory  = 0x00000010
	attrArchive    = 0x00000020
	attrNormal     = 0x00000080
	attrTemporary  = 0x00000100
	attrCompressed = 0x00000800
)

// filesize converts a byte count to a human-readable string using the largest
// appropriate unit, formatted as "%4.2f %s" (decimal-style unit names,
// 1024-based steps).
func filesize(size uint64) string {
	units := []string{"B", "kB", "MB", "GB", "TB", "PB"}
	value := float64(size)
	k := 0
	for k < len(units)-1 && value >= 1024.0 {
		value /= 1024.0
		k++
	}
	return fmt.Sprintf("%4.2f %s", value, units[k])
}

// windowsLsEntry renders a single directory entry as an 8-character
// attribute string (dachnrst), a
// right-aligned human size, the access date, and the (colored) name. Directories
// are shown in bold cyan with a trailing backslash; files in bold.
func windowsLsEntry(entry smbclient.FileInfo, name string) string {
	attrs := entry.FileAttributes

	flag := func(bit uint32, set string) string {
		if attrs&bit != 0 {
			return set
		}
		return "-"
	}

	meta := ""
	meta += flag(attrDirectory, "d")
	meta += flag(attrArchive, "a")
	meta += flag(attrCompressed, "c")
	meta += flag(attrHidden, "h")
	meta += flag(attrNormal, "n")
	meta += flag(attrReadonly, "r")
	meta += flag(attrSystem, "s")
	meta += flag(attrTemporary, "t")

	sizeStr := filesize(entry.Size)

	// Use the access time for the listing date.
	date := entry.LastAccessTime
	if date.IsZero() {
		date = entry.LastWriteTime
	}
	dateStr := date.Format("2006-01-02 15:04")

	if entry.IsDir() {
		return fmt.Sprintf("%s %10s  %s  %s%s%s\\", meta, sizeStr, dateStr, shell.ColorBoldCyan, name, shell.ColorReset)
	}
	return fmt.Sprintf("%s %10s  %s  %s%s%s", meta, sizeStr, dateStr, shell.ColorBold, name, shell.ColorReset)
}
