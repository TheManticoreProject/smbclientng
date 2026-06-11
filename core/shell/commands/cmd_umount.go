package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/TheManticoreProject/smbclient-ng/core/shell"
)

func init() {
	shell.RegisterCommand(&shell.Command{
		Name:         "umount",
		Description:  []string{"Removes a mount point of the remote share on the local machine.", "Syntax: 'umount <local_mount_point>'"},
		Autocomplete: []string{"local_directory"},
		Handler:      cmdUmount,
	})
}

// cmdUmount removes a local SMB mount point created by 'mount' by shelling out
// to the platform's native tool: 'net use /delete' on Windows, 'umount' on
// Linux and macOS.
func cmdUmount(s *shell.Shell, args []string) error {
	if len(args) != 1 {
		s.Errorf("Syntax: 'umount <local_mount_point>'")
		return nil
	}
	localMount := args[0]

	if _, err := os.Stat(localMount); os.IsNotExist(err) {
		s.Errorf("Cannot unmount a non-existing path: '%s'", localMount)
		return nil
	}

	name, cmdArgs, err := umountCommand(localMount)
	if err != nil {
		s.Errorf("%s", err)
		return nil
	}

	s.Debugf("Executing '%s' with arguments: %v", name, cmdArgs)
	out, runErr := exec.Command(name, cmdArgs...).CombinedOutput()
	if runErr != nil {
		s.Errorf("Unmount command failed: %s: %s", runErr, strings.TrimSpace(string(out)))
		return nil
	}
	if trimmed := strings.TrimSpace(string(out)); trimmed != "" {
		s.Debugf("Command output: %s", trimmed)
	}
	s.Infof("Successfully unmounted local path '%s'.", localMount)
	return nil
}

// umountCommand builds the platform-specific unmount executable and arguments.
func umountCommand(localMount string) (string, []string, error) {
	switch runtime.GOOS {
	case "windows":
		return "net", []string{"use", localMount, "/delete"}, nil
	case "linux", "darwin":
		return "umount", []string{localMount}, nil
	default:
		return "", nil, fmt.Errorf("unsupported platform '%s' for unmounting an SMB share", runtime.GOOS)
	}
}
