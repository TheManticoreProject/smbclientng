package commands

import (
	"testing"

	"github.com/TheManticoreProject/smbclient-ng/core/client"
)

func TestShareRights(t *testing.T) {
	cases := []struct {
		name string
		in   client.Share
		want string
	}{
		{"unknown", client.Share{RightsKnown: false}, "UNKNOWN"},
		{"read-write", client.Share{RightsKnown: true, Readable: true, Writable: true}, "READ, WRITE"},
		{"read", client.Share{RightsKnown: true, Readable: true}, "READ"},
		{"write", client.Share{RightsKnown: true, Writable: true}, "WRITE"},
		{"no-access", client.Share{RightsKnown: true}, "NO ACCESS"},
	}
	for _, c := range cases {
		if got := shareRights(c.in); got != c.want {
			t.Errorf("%s: shareRights = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestShareType(t *testing.T) {
	cases := []struct {
		in   uint32
		want string
	}{
		{0x00000000, "DISKTREE"},
		{0x00000001, "PRINTQ"},
		{0x00000003, "IPC"},
		{0x80000003, "IPC, SPECIAL"},      // IPC$ (admin/special)
		{0x80000000, "DISKTREE, SPECIAL"}, // e.g. C$, ADMIN$
		{0x40000000, "DISKTREE, TEMPORARY"},
		{0xC0000000, "DISKTREE, SPECIAL, TEMPORARY"},
	}
	for _, c := range cases {
		if got := shareType(c.in); got != c.want {
			t.Errorf("shareType(0x%08x) = %q, want %q", c.in, got, c.want)
		}
	}
}
