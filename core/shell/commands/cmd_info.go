package commands

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:        "info",
		Description: []string{"Get information about the server and or the share.", "Syntax: 'info <--server|--share>'"},
		Handler:     cmdInfo,
	})
}

func cmdInfo(s *shell.Shell, args []string) error {
	if !s.SM().IsConnected() {
		s.Errorf("Not connected to a server.")
		return nil
	}

	server, share := false, false
	for _, arg := range args {
		switch arg {
		case "--server", "server":
			server = true
		case "--share", "share":
			share = true
		default:
			s.Errorf("Unknown argument '%s'. Syntax: 'info <--server|--share>'", arg)
			return nil
		}
	}
	// Mutually exclusive and required.
	if server == share {
		s.Errorf("Syntax: 'info <--server|--share>'")
		return nil
	}

	if server {
		printServerInfo(s)
	}
	if share {
		if s.SM().CurrentShare() == "" {
			s.Errorf("You must open a share first, try the 'use <share>' command.")
			return nil
		}
		printShareInfo(s)
	}
	return nil
}

// infoRow renders one "<lead><label> <dashes> : <value>" line of the info tree,
// colored (blue label, grey dashes, yellow value). lead is the
// full connector prefix, e.g. "  │ ├─ " or "  ├─ ".
func infoRow(s *shell.Shell, lead, label string, dashes int, value string) {
	s.Print(formatInfoRow(lead, label, dashes, value))
}

// formatInfoRow builds the colored info-tree line. Kept pure for testing.
func formatInfoRow(lead, label string, dashes int, value string) string {
	return fmt.Sprintf("%s\x1b[94m%s\x1b[0m \x1b[90m%s\x1b[0m : \x1b[93m%s\x1b[0m",
		lead, label, strings.Repeat("─", dashes), value)
}

// printServerInfo prints the "[+] Server:" tree. "Login Required" is omitted as
// it has no source in the Manticore SMB stack. "OS Name" is populated for SMB1
// only (SMB2/SMB3 do not transmit it); it renders empty on SMB2/SMB3.
func printServerInfo(s *shell.Shell) {
	id := s.SM().Client().ServerIdentity()
	conn := s.SM().Client().ConnectionInfo()

	s.Print("[+] Server:")
	s.Print("  ├─NetBIOS:")
	infoRow(s, "  │ ├─ ", "NetBIOS Hostname", 8, id.NetBIOSComputerName)
	infoRow(s, "  │ └─ ", "NetBIOS Domain", 10, id.NetBIOSDomainName)
	s.Print("  ├─DNS:")
	infoRow(s, "  │ ├─ ", "DNS Hostname", 12, id.DNSComputerName)
	infoRow(s, "  │ └─ ", "DNS Domain", 14, id.DNSDomainName)
	s.Print("  ├─OS:")
	infoRow(s, "  │ ├─ ", "OS Name", 17, id.OSName)
	infoRow(s, "  │ └─ ", "OS Version", 14, fmt.Sprintf("%d.%d.%d", id.OSVersionMajor, id.OSVersionMinor, id.OSVersionBuild))
	s.Print("  ├─Server:")
	infoRow(s, "  │ ├─ ", "Signing Required", 8, pyBool(conn.SigningRequired))
	infoRow(s, "  │ ├─ ", "Supports NTLMv2", 9, pyBool(conn.SupportsNTLMv2))
	infoRow(s, "  │ ├─ ", "Max size of read chunk", 2, fmt.Sprintf("%d bytes (%s)", conn.MaxReadSize, filesize(uint64(conn.MaxReadSize))))
	infoRow(s, "  │ └─ ", "Max size of write chunk", 1, fmt.Sprintf("%d bytes (%s)", conn.MaxWriteSize, filesize(uint64(conn.MaxWriteSize))))
	s.Print("  └─")
}

// printShareInfo prints the "[+] Share:" block for the current share, looking up
// its metadata from the srvsvc share enumeration.
func printShareInfo(s *shell.Shell) {
	current := s.SM().CurrentShare()

	var (
		name    = current
		comment string
		typeStr string
		rawType string
	)
	if shares, err := s.SM().ListShares(); err == nil {
		for _, sh := range shares {
			if strings.EqualFold(sh.Name, current) {
				name = sh.Name
				comment = sh.Comment
				typeStr = shareType(sh.Type)
				rawType = fmt.Sprintf("%d", sh.Type)
				break
			}
		}
	}

	s.Print("")
	s.Print("[+] Share:")
	infoRow(s, "  ├─ ", "Name", 12, name)
	infoRow(s, "  ├─ ", "Description", 5, comment)
	infoRow(s, "  ├─ ", "Type", 12, typeStr)
	infoRow(s, "  └─ ", "Raw type value", 2, rawType)
}

// pyBool renders a boolean as "True"/"False".
func pyBool(b bool) string {
	if b {
		return "True"
	}
	return "False"
}
