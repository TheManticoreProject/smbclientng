package client

import "testing"

func TestRightsFromMask(t *testing.T) {
	cases := []struct {
		name         string
		mask         uint32
		wantR, wantW bool
	}{
		{"none", 0x00000000, false, false},
		{"generic-read", maskGenericRead, true, false},
		{"generic-write", maskGenericWrite, false, true},
		{"generic-all", maskGenericAll, true, true},
		{"file-read-data", maskFileReadData, true, false},
		{"file-write-data", maskFileWriteData, false, true},
		{"file-append-data", maskFileAppendData, false, true},
		{"read-control-only", 0x00020000, false, false}, // READ_CONTROL is not data read
		{"full-control", 0x001f01ff, true, true},
	}
	for _, c := range cases {
		r, w := rightsFromMask(c.mask)
		if r != c.wantR || w != c.wantW {
			t.Errorf("%s: rightsFromMask(0x%08x) = (r=%v, w=%v), want (r=%v, w=%v)",
				c.name, c.mask, r, w, c.wantR, c.wantW)
		}
	}
}
