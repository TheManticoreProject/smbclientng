package client

import (
	"fmt"
	"io"

	"github.com/TheManticoreProject/smbclient-ng/core/utils"

	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// transferChunk bounds how many bytes are buffered in memory per read/write
// iteration. The SMB client further splits each call to fit the negotiated
// MaxBufferSize, so this only caps our own buffering, not the wire frame size.
const transferChunk = 0xFF00

// requireShare returns an error if no share is currently selected.
func (sm *SessionManager) requireShare() error {
	if sm.client == nil {
		return fmt.Errorf("not connected")
	}
	if sm.share == "" {
		return fmt.Errorf("no share selected; use 'use <share>' first")
	}
	return nil
}

// DownloadTo streams the remote file at the share-relative path to w and returns
// the number of bytes transferred.
func (sm *SessionManager) DownloadTo(path string, w io.Writer) (int64, error) {
	if err := sm.requireShare(); err != nil {
		return 0, err
	}

	h, err := sm.client.OpenFile(path, smbclient.OpenOptions{
		DesiredAccess:     fileflags.GENERIC_READ,
		ShareAccess:       fileflags.FILE_SHARE_READ,
		CreateDisposition: fileflags.FILE_OPEN,
		CreateOptions:     fileflags.FILE_NON_DIRECTORY_FILE,
	})
	if err != nil {
		return 0, fmt.Errorf("cannot open remote file %q: %w", path, err)
	}
	defer sm.client.CloseFile(h)

	var total int64
	for {
		chunk, err := sm.client.ReadFile(h, uint64(total), transferChunk)
		if len(chunk) > 0 {
			n, werr := w.Write(chunk)
			total += int64(n)
			if werr != nil {
				return total, werr
			}
		}
		if err != nil {
			return total, err
		}
		// A short (or empty) read means we have reached the end of the file.
		if uint32(len(chunk)) < transferChunk {
			break
		}
	}
	return total, nil
}

// UploadFrom streams r into the remote file at the share-relative path, creating
// or overwriting it, and returns the number of bytes transferred.
func (sm *SessionManager) UploadFrom(path string, r io.Reader) (int64, error) {
	if err := sm.requireShare(); err != nil {
		return 0, err
	}

	h, err := sm.client.OpenFile(path, smbclient.OpenOptions{
		DesiredAccess:     fileflags.GENERIC_WRITE,
		ShareAccess:       fileflags.FILE_SHARE_READ | fileflags.FILE_SHARE_WRITE,
		CreateDisposition: fileflags.FILE_OVERWRITE_IF,
		CreateOptions:     fileflags.FILE_NON_DIRECTORY_FILE,
	})
	if err != nil {
		return 0, fmt.Errorf("cannot create remote file %q: %w", path, err)
	}
	defer sm.client.CloseFile(h)

	buf := make([]byte, transferChunk)
	var total int64
	for {
		n, rerr := r.Read(buf)
		if n > 0 {
			if _, werr := sm.client.WriteFile(h, uint64(total), buf[:n]); werr != nil {
				return total, werr
			}
			total += int64(n)
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return total, rerr
		}
	}
	return total, nil
}

// Delete removes the remote file at the share-relative path.
func (sm *SessionManager) Delete(path string) error {
	if err := sm.requireShare(); err != nil {
		return err
	}
	return sm.client.DeleteFile(path)
}

// MakeDir creates a directory at the share-relative path.
func (sm *SessionManager) MakeDir(path string) error {
	if err := sm.requireShare(); err != nil {
		return err
	}
	return sm.client.CreateDirectory(path)
}

// RemoveDir removes the directory at the share-relative path.
func (sm *SessionManager) RemoveDir(path string) error {
	if err := sm.requireShare(); err != nil {
		return err
	}
	return sm.client.DeleteDirectory(path)
}

// Rename moves/renames a remote file or directory.
func (sm *SessionManager) Rename(oldPath, newPath string) error {
	if err := sm.requireShare(); err != nil {
		return err
	}
	return sm.client.RenameFile(oldPath, newPath)
}

// ListPattern enumerates remote entries matching the share-relative pattern
// (with SMB wildcards), resolved against the current working directory.
func (sm *SessionManager) ListPattern(arg string) ([]smbclient.FileInfo, error) {
	if err := sm.requireShare(); err != nil {
		return nil, err
	}
	full := sm.Resolve(arg)
	return sm.client.ListDirectory(utils.RemoteDir(full), utils.RemoteBase(full))
}
