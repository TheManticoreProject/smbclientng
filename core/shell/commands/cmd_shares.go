package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TheManticoreProject/smbclient-ng/core/client"
	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

// MS-SRVS 2.2.2.4 share type (STYPE_*) bit masks.
const (
	stypeSpecial   = 0x80000000
	stypeTemporary = 0x40000000
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:        "shares",
		Description: []string{"Lists the SMB shares served by the remote machine.", "Syntax: 'shares [-R]'"},
		Handler:     cmdShares,
	})
}

func cmdShares(s *shell.Shell, args []string) error {
	if !s.SM().IsConnected() {
		s.Errorf("Not connected to a server.")
		return nil
	}

	rights := false
	for _, arg := range args {
		switch arg {
		case "-R", "--rights":
			rights = true
		default:
			s.Errorf("Unknown argument '%s'. Syntax: 'shares [-R]'", arg)
			return nil
		}
	}

	var (
		shares []client.Share
		err    error
	)
	if rights {
		// Level 502 carries the share security descriptor we derive rights
		// from; fall back to a plain listing if the server denies it.
		shares, err = s.SM().ListSharesWithRights()
		if err != nil {
			s.Errorf("Could not read share rights (%s); listing without rights.", err)
			rights = false
			shares, err = s.SM().ListShares()
		}
	} else {
		shares, err = s.SM().ListShares()
	}
	if err != nil {
		s.Errorf("Could not list shares: %s", err)
		return nil
	}
	if len(shares) == 0 {
		s.Errorf("No share served on '%s'", s.SM().Host)
		return nil
	}

	sort.Slice(shares, func(i, j int) bool {
		return strings.ToLower(shares[i].Name) < strings.ToLower(shares[j].Name)
	})

	printSharesTable(s, shares, rights)
	return nil
}

// printSharesTable renders the share listing as an aligned table with Share,
// Visibility, Type and Description columns. Hidden
// shares (names ending in '$') are shown in blue, visible ones in yellow. When
// rights is set, an extra Rights column derived from each share ACL is shown.
func printSharesTable(s *shell.Shell, shares []client.Share, rights bool) {
	const (
		colYellow = "\x1b[1;93m"
		colBlue   = "\x1b[1;94m"
	)

	headers := []string{"Share", "Visibility", "Type", "Description"}
	if rights {
		headers = append(headers, "Rights")
	}
	rows := make([][]string, 0, len(shares))
	for _, sh := range shares {
		visibility := "Visible"
		if strings.HasSuffix(sh.Name, "$") {
			visibility = "Hidden"
		}
		row := []string{sh.Name, visibility, shareType(sh.Type), sh.Comment}
		if rights {
			row = append(row, shareRights(sh))
		}
		rows = append(rows, row)
	}

	// Compute column widths from the plain (uncolored) cell text.
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Header.
	var hb strings.Builder
	for i, h := range headers {
		fmt.Fprintf(&hb, "%s%-*s%s  ", shell.ColorBold, widths[i], h, shell.ColorReset)
	}
	s.Print(strings.TrimRight(hb.String(), " "))

	// Rows.
	for _, row := range rows {
		color := colYellow
		if row[1] == "Hidden" {
			color = colBlue
		}
		var rb strings.Builder
		for i, cell := range row {
			fmt.Fprintf(&rb, "%s%-*s%s  ", color, widths[i], cell, shell.ColorReset)
		}
		s.Print(strings.TrimRight(rb.String(), " "))
	}
}

// shareRights renders the ACL-derived access of a share. When the rights are
// unknown (no security descriptor was returned) it shows "UNKNOWN".
func shareRights(sh client.Share) string {
	if !sh.RightsKnown {
		return "UNKNOWN"
	}
	switch {
	case sh.Readable && sh.Writable:
		return "READ, WRITE"
	case sh.Readable:
		return "READ"
	case sh.Writable:
		return "WRITE"
	default:
		return "NO ACCESS"
	}
}

// shareType renders an STYPE_* value as comma-separated names, e.g.
// "DISKTREE" or "IPC, SPECIAL".
func shareType(t uint32) string {
	base := t &^ (stypeSpecial | stypeTemporary)
	var names []string
	switch base {
	case 0x00000000:
		names = append(names, "DISKTREE")
	case 0x00000001:
		names = append(names, "PRINTQ")
	case 0x00000002:
		names = append(names, "DEVICE")
	case 0x00000003:
		names = append(names, "IPC")
	default:
		names = append(names, fmt.Sprintf("0x%08x", base))
	}
	if t&stypeSpecial != 0 {
		names = append(names, "SPECIAL")
	}
	if t&stypeTemporary != 0 {
		names = append(names, "TEMPORARY")
	}
	return strings.Join(names, ", ")
}
