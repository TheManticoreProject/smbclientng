module github.com/TheManticoreProject/smbclient-ng

go 1.24.0

require (
	github.com/TheManticoreProject/Manticore v1.0.8
	github.com/TheManticoreProject/goopts v1.2.4
	github.com/TheManticoreProject/winacl v1.2.14
	github.com/chzyer/readline v1.5.1
)

require golang.org/x/sys v0.0.0-20220310020820-b874c991c1a5 // indirect

// The SMB v1.0 client file/directory operations used by this tool
// (ListDirectory, ReadFile, WriteFile, DeleteFile, ...) are only available in
// the in-development Manticore checkout, not in any published release yet.
// Point at the local working copy until a release ships them.
replace github.com/TheManticoreProject/Manticore => ../Manticore
