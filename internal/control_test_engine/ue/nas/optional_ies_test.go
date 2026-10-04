/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package nas

import (
	"encoding/hex"
	"testing"

	message "github.com/free5gc/nas/message"
	"github.com/stretchr/testify/require"
)

// A core may put these optional IEs in its messages (TS 24.501, TS 24.008). free5gc/nas v1.3.0
// has no codec for them and reports the message as an error, which had the UE drop for instance
// its Registration Accept: third_party/free5gc-nas adds them.
func TestOptionalIEsAreDecoded(t *testing.T) {
	for name, wire := range map[string]string{
		"registration/mico":                 "7e00420101b3",
		"registration/emergency":            "7e0042010134030201f1",
		"registration/extended-emergency":   "7e004201017a00040001f100",
		"registration/sor":                  "7e0042010173001300000000000000000000000000000000000000",
		"registration/nssai-mode":           "7e00420101a1",
		"registration/operator-categories":  "7e00420101760000",
		"registration/non3gpp-policies":     "7e00420101d1",
		"registration/eps-status":           "7e0042010160022000",
		"configuration/mico":                "7e0054b0",
		"configuration/operator-categories": "7e0054760000",
		"configuration/full-name-ucs2":      "7e00544307900046005200e9",
		"configuration/short-name-ucs2":     "7e005445039000e9",
		"configuration/sms":                 "7e0054f1",
		"security/eps-algorithms":           "7e005d220102a0205722",
		"security/s1-capabilities":          "7e005d220102a0201902a020",
		"transport/additional-information":  "7e00680100052e0101c31f240100",
		"session/always-on":                 "2e0101c211000601000320ff01060103e80103e881",
		"session/mapped-eps":                "2e0101c211000601000320ff01060103e80103e875000450000180",
		"session/allowed-ssc":               "2e0101c31ff3",
		"session/reject-congestion":         "2e0101c31f610103",
		"session/release-congestion":        "2e0101d324610103",
	} {
		plain, err := hex.DecodeString(wire)
		require.NoError(t, err, name)
		msg, err := message.Parse(plain, nil)
		require.NoError(t, err, name)
		// A decoder that skipped the IE would not encode it back.
		encoded, err := msg.MarshalBinary()
		require.NoError(t, err, name)
		require.Equal(t, plain, encoded, name)
	}
}
