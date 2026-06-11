package commands

import (
	"os"
	"path/filepath"
	"testing"

	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
)

func TestSortDirsFirst(t *testing.T) {
	entries := []smbclient.FileInfo{
		{Name: "zebra.txt"},
		{Name: "Alpha", FileAttributes: attrDirectory},
		{Name: "apple.txt"},
		{Name: "beta", FileAttributes: attrDirectory},
	}
	sortDirsFirst(entries)

	got := []string{}
	for _, e := range entries {
		got = append(got, e.Name)
	}
	want := []string{"Alpha", "beta", "apple.txt", "zebra.txt"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sortDirsFirst = %v, want %v", got, want)
		}
	}
}

func TestUnixPermissions(t *testing.T) {
	dir := t.TempDir()

	file := filepath.Join(dir, "f")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if got := unixPermissions(fi); got != "-rw-r--r--" {
		t.Errorf("file perms = %q, want %q", got, "-rw-r--r--")
	}

	sub := filepath.Join(dir, "d")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	di, err := os.Stat(sub)
	if err != nil {
		t.Fatal(err)
	}
	if got := unixPermissions(di); got != "drwxr-xr-x" {
		t.Errorf("dir perms = %q, want %q", got, "drwxr-xr-x")
	}
}
