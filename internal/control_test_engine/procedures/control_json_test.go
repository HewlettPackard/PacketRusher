// SPDX-License-Identifier: Apache-2.0
package procedures

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAttachmentSessionIdentitiesJSON(t *testing.T) {
	for _, test := range []struct {
		name string
		ids  []uint8
		wire string
	}{
		{"nil", nil, "[]"},
		{"registration-only", []uint8{}, "[]"},
		{"multiple sessions", []uint8{1, 15}, "[1,15]"},
	} {
		t.Run(test.name, func(t *testing.T) {
			original := Attachment{UE: 3, Generation: 2, ConnectionGeneration: 4,
				State: "registered", GNB: "000008", Connected: true, Ready: true,
				ActivePDUSessions: test.ids}
			encoded, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			if actual := string(fields["active_pdu_sessions"]); actual != test.wire {
				t.Fatalf("session identities must be an integer array: got %s, want %s", actual, test.wire)
			}
			var decoded Attachment
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatal(err)
			}
			if original.ActivePDUSessions == nil {
				original.ActivePDUSessions = []uint8{}
			}
			if !reflect.DeepEqual(original, decoded) {
				t.Fatalf("attachment fields changed across JSON round trip: got %+v, want %+v", decoded, original)
			}
		})
	}
}
