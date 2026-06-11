package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "mount",
		Description:  []string{"Creates a mount point of the remote share on the local machine.", "Syntax: 'mount <remote_path> <local_mountpoint>'"},
		Autocomplete: []string{"remote_directory", "local_directory"},
		Handler:      cmdMount,
	})
}

// cmdMount mounts the current share (at the given remote sub-path) onto a local
// directory by shelling out to the platform's native SMB mount tool, mirroring
// the Python smbclient-ng behaviour: 'net use' on Windows, 'mount -t cifs' on
// Linux, and 'mount_smbfs' on macOS.
func cmdMount(s *shell.Shell, args []string) error {
	if s.SM().CurrentShare() == "" {
		s.Errorf("You must open a share first, try the 'use <share>' command.")
		return nil
	}
	if len(args) != 2 {
		s.Errorf("Syntax: 'mount <remote_path> <local_mountpoint>'")
		return nil
	}

	remotePath := s.SM().Resolve(args[0]) // share-relative, backslash-separated, no leading sep
	localMount := args[1]

	if _, err := os.Stat(localMount); os.IsNotExist(err) {
		s.Debugf("Local mountpoint '%s' does not exist, creating it.", localMount)
		if err := os.MkdirAll(localMount, 0o755); err != nil {
			s.Errorf("Could not create local mountpoint '%s': %s", localMount, err)
			return nil
		}
	}

	name, cmdArgs, err := mountCommand(s, remotePath, localMount)
	if err != nil {
		s.Errorf("%s", err)
		return nil
	}

	s.Debugf("Executing '%s' with arguments: %v", name, cmdArgs)
	out, runErr := exec.Command(name, cmdArgs...).CombinedOutput()
	if runErr != nil {
		s.Errorf("Mount command failed: %s: %s", runErr, strings.TrimSpace(string(out)))
		return nil
	}
	if trimmed := strings.TrimSpace(string(out)); trimmed != "" {
		s.Debugf("Command output: %s", trimmed)
	}
	s.Infof("Successfully mounted '\\\\%s\\%s\\%s' to local '%s'.", s.SM().Host, s.SM().CurrentShare(), remotePath, localMount)
	return nil
}

// mountCommand builds the platform-specific mount executable and arguments.
// remotePath is the share-relative path (backslash separated, no leading sep).
func mountCommand(s *shell.Shell, remotePath, localMount string) (string, []string, error) {
	host := s.SM().Host
	share := s.SM().CurrentShare()
	creds := s.SM().Creds

	switch runtime.GOOS {
	case "windows":
		unc := fmt.Sprintf("\\\\%s\\%s", host, share)
		if remotePath != "" {
			unc += "\\" + remotePath
		}
		return "net", []string{"use", localMount, unc}, nil

	case "linux":
		src := fmt.Sprintf("//%s/%s", host, share)
		if remotePath != "" {
			src += "/" + strings.ReplaceAll(remotePath, "\\", "/")
		}
		opts := mountOptions(creds)
		return "mount", []string{"-t", "cifs", src, localMount, "-o", opts}, nil

	case "darwin":
		user := ""
		if creds != nil {
			if u := creds.GetUsername(); u != "" {
				user = u
				if p := creds.GetPassword(); p != "" {
					user += ":" + p
				}
				user += "@"
			}
		}
		src := fmt.Sprintf("//%s%s/%s", user, host, share)
		if remotePath != "" {
			src += "/" + strings.ReplaceAll(remotePath, "\\", "/")
		}
		return "mount_smbfs", []string{src, localMount}, nil

	default:
		return "", nil, fmt.Errorf("unsupported platform '%s' for mounting an SMB share", runtime.GOOS)
	}
}

// mountOptions builds the comma-separated '-o' option string for the Linux cifs
// mount: credentials and, when present, the authentication domain.
func mountOptions(creds *credentials.Credentials) string {
	if creds == nil {
		return "guest"
	}
	parts := []string{fmt.Sprintf("username=%s", creds.GetUsername())}
	parts = append(parts, fmt.Sprintf("password=%s", creds.GetPassword()))
	if dom := creds.GetDomain(); dom != "" && dom != "." {
		parts = append(parts, fmt.Sprintf("domain=%s", dom))
	}
	return strings.Join(parts, ",")
}
