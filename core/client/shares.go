package client

import (
	"fmt"

	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0/structures"
	mssrvs "github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/ms-srvs"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"

	"github.com/TheManticoreProject/winacl/ace/acetype"
	"github.com/TheManticoreProject/winacl/securitydescriptor"
)

// Share describes a single SMB share exported by the server, as returned by the
// srvsvc NetrShareEnum enumeration.
type Share struct {
	Name    string
	Type    uint32 // raw STYPE_* value, including the SPECIAL/TEMPORARY high bits
	Comment string

	// RightsKnown is set when the share was enumerated with its security
	// descriptor (level 502) and Readable/Writable were derived from its ACL.
	RightsKnown bool
	Readable    bool
	Writable    bool
}

// Access-mask bits used to map a DACL grant to read/write rights ([MS-DTYP]
// 2.4.3 ACCESS_MASK; the generic bits and the file-specific data bits).
const (
	maskGenericRead    = 0x80000000
	maskGenericWrite   = 0x40000000
	maskGenericAll     = 0x10000000
	maskFileReadData   = 0x00000001
	maskFileWriteData  = 0x00000002
	maskFileAppendData = 0x00000004
)

// worldSIDs are the well-known principals that every authenticated SMB session
// belongs to. Share rights are evaluated for these so the reported access
// reflects what an ordinary authenticated user is granted by the share ACL.
var worldSIDs = map[string]bool{
	"S-1-1-0":      true, // Everyone
	"S-1-5-11":     true, // Authenticated Users
	"S-1-5-32-545": true, // BUILTIN\Users
}

// restoreShare reconnects the caller's originally selected share (or tears down
// the temporary IPC$ tree when none was selected). It is used to undo the
// IPC$ tree switch performed for srvsvc calls.
func (sm *SessionManager) restoreShare() {
	if sm.share != "" {
		_ = sm.client.TreeConnect(sm.share)
	} else {
		_ = sm.client.TreeDisconnect()
	}
}

// ListShares enumerates the shares served by the remote host via the srvsvc
// DCE/RPC interface (MS-SRVS NetrShareEnum, info level 1). The srvsvc named pipe
// lives on the IPC$ tree, so this temporarily tree-connects to IPC$ and restores
// the previously selected share (if any) before returning.
func (sm *SessionManager) ListShares() ([]Share, error) {
	if sm.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	// The srvsvc pipe is opened on the current tree, which must be IPC$.
	if err := sm.client.TreeConnect("IPC$"); err != nil {
		return nil, fmt.Errorf("cannot connect to IPC$: %w", err)
	}
	defer sm.restoreShare()

	infos, err := mssrvs.New(sm.client).ListShares()
	if err != nil {
		return nil, err
	}

	shares := make([]Share, 0, len(infos))
	for _, info := range infos {
		shares = append(shares, Share{
			Name:    info.Name,
			Type:    info.Type,
			Comment: info.Comment,
		})
	}
	return shares, nil
}

// ListSharesWithRights enumerates shares at info level 502, which carries each
// share's security descriptor, and derives Readable/Writable from the share
// ACL (no access is actively probed on the server). Level 502 is more
// privileged than level 1; servers may return access-denied, in which case the
// error is returned and the caller can fall back to ListShares.
func (sm *SessionManager) ListSharesWithRights() ([]Share, error) {
	if sm.client == nil {
		return nil, fmt.Errorf("not connected")
	}

	if err := sm.client.TreeConnect("IPC$"); err != nil {
		return nil, fmt.Errorf("cannot connect to IPC$: %w", err)
	}
	defer sm.restoreShare()

	transport, err := sm.client.RPCTransport(srvsvc.PipeName)
	if err != nil {
		return nil, fmt.Errorf("open srvsvc pipe: %w", err)
	}
	rpc := dcerpcclient.NewClient(transport)
	if err := rpc.Bind(srvsvc.SyntaxID()); err != nil {
		return nil, fmt.Errorf("bind srvsvc: %w", err)
	}
	defer rpc.Close()

	resume := ndr.DWORD(0)
	info := structures.SHARE_ENUM_STRUCT{
		Level: 502,
		ShareInfo: structures.SHARE_ENUM_UNION{
			Tag:      502,
			Level502: &structures.SHARE_INFO_502_CONTAINER{},
		},
	}

	out, _, _, err := functions.NetrShareEnum(rpc, "", info, ndr.DWORD(mssrvs.MaxPreferredLength), &resume)
	if err != nil {
		return nil, err
	}
	container := out.ShareInfo.Level502
	if container == nil {
		return nil, nil
	}

	shares := make([]Share, 0, len(container.Buffer))
	for _, s := range container.Buffer {
		share := Share{
			Name:    string(s.Shi502Netname),
			Type:    uint32(s.Shi502Type),
			Comment: string(s.Shi502Remark),
		}
		if len(s.Shi502SecurityDescriptor) > 0 {
			share.Readable, share.Writable = shareACLRights(s.Shi502SecurityDescriptor)
			share.RightsKnown = true
		}
		shares = append(shares, share)
	}
	return shares, nil
}

// shareACLRights parses a share security-descriptor blob and reports the read
// and write access its DACL grants to the well-known world principals. A NULL
// DACL (no DACL present) grants everyone full control.
func shareACLRights(sd []byte) (readable, writable bool) {
	descriptor := securitydescriptor.NewSecurityDescriptor()
	if _, err := descriptor.Unmarshal(sd); err != nil {
		return false, false
	}
	if descriptor.DACL == nil {
		// A NULL DACL grants full access to everyone.
		return true, true
	}

	var allowed, denied uint32
	for _, entry := range descriptor.DACL.Entries {
		if !worldSIDs[entry.Identity.SID.String()] {
			continue
		}
		switch entry.Header.Type.Value {
		case acetype.ACE_TYPE_ACCESS_ALLOWED, acetype.ACE_TYPE_ACCESS_ALLOWED_OBJECT:
			allowed |= entry.Mask.RawValue
		case acetype.ACE_TYPE_ACCESS_DENIED, acetype.ACE_TYPE_ACCESS_DENIED_OBJECT:
			denied |= entry.Mask.RawValue
		}
	}

	return rightsFromMask(allowed &^ denied)
}

// rightsFromMask maps an effective ACCESS_MASK to read/write rights, treating
// the generic read/write/all bits and the file data bits as the indicators of
// share read and write access.
func rightsFromMask(effective uint32) (readable, writable bool) {
	readable = effective&(maskGenericRead|maskGenericAll|maskFileReadData) != 0
	writable = effective&(maskGenericWrite|maskGenericAll|maskFileWriteData|maskFileAppendData) != 0
	return readable, writable
}
