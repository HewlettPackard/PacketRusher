package tunnel

import (
	"testing"

	"github.com/free5gc/go-gtp5gnl"
)

func TestParseOID(t *testing.T) {
	cases := []struct {
		name    string
		s       string
		oid     gtp5gnl.OID
		wantErr bool
	}{
		{
			s:   "456",
			oid: gtp5gnl.OID{456},
		},
		{
			s:   "123:456",
			oid: gtp5gnl.OID{123, 456},
		},
		{
			s:       "a",
			wantErr: true,
		},
		{
			s:       "123:a",
			wantErr: true,
		},
		{
			s:       "a:456",
			wantErr: true,
		},
		{
			s:       ":456",
			wantErr: true,
		},
		{
			s:       "123:",
			wantErr: true,
		},
		{
			s:       ":",
			wantErr: true,
		},
		{
			s:       "",
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			oid, err := ParseOID(tc.s)
			if err != nil {
				if !tc.wantErr {
					t.Fatal(err)
				}
				return
			}
			if !tc.oid.Equal(oid) {
				t.Errorf("want %v; but got %v\n", tc.oid, oid)
			}
		})
	}
}
