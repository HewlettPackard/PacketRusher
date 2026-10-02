/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package auth

import (
	nasType "github.com/free5gc/nas/ie"
	security "github.com/free5gc/nas/message"
	"github.com/free5gc/util/ueauth"
)

// Algorithm key Derivation function defined in TS 33.501 Annex A.9
func AlgorithmKeyDerivation(cipheringAlg uint8, kamf []byte, knasEnc *[16]uint8, integrityAlg uint8, knasInt *[16]uint8) error {
	// Security Key
	P0 := []byte{security.NNASEncAlg}
	L0 := ueauth.KDFLen(P0)
	P1 := []byte{cipheringAlg}
	L1 := ueauth.KDFLen(P1)

	kenc, err := ueauth.GetKDFValue(kamf, ueauth.FC_FOR_ALGORITHM_KEY_DERIVATION, P0, L0, P1, L1)
	if err != nil {
		return err
	}
	copy(knasEnc[:], kenc[16:32])

	// Integrity Key
	P0 = []byte{security.NNASIntAlg}
	L0 = ueauth.KDFLen(P0)
	P1 = []byte{integrityAlg}
	L1 = ueauth.KDFLen(P1)

	kint, err := ueauth.GetKDFValue(kamf, ueauth.FC_FOR_ALGORITHM_KEY_DERIVATION, P0, L0, P1, L1)
	if err != nil {
		return err
	}
	copy(knasInt[:], kint[16:32])

	return nil
}

func SelectAlgorithms(securityCapability *nasType.UESecCapability) (intergritygAlgorithm uint8, cipheringAlgorithm uint8) {
	// set the algorithms of integrity
	if securityCapability.IA05G {
		intergritygAlgorithm = uint8(security.AlgIntegrity128NIA0)
	} else if securityCapability.IA1_128_5G {
		intergritygAlgorithm = uint8(security.AlgIntegrity128NIA1)
	} else if securityCapability.IA2_128_5G {
		intergritygAlgorithm = uint8(security.AlgIntegrity128NIA2)
	} else if securityCapability.IA3_128_5G {
		intergritygAlgorithm = uint8(security.AlgIntegrity128NIA3)
	}

	// set the algorithms of ciphering
	if securityCapability.EA05G {
		cipheringAlgorithm = uint8(security.AlgCiphering128NEA0)
	} else if securityCapability.EA1_128_5G {
		cipheringAlgorithm = uint8(security.AlgCiphering128NEA1)
	} else if securityCapability.EA2_128_5G {
		cipheringAlgorithm = uint8(security.AlgCiphering128NEA2)
	} else if securityCapability.EA3_128_5G {
		cipheringAlgorithm = uint8(security.AlgCiphering128NEA3)
	}

	return intergritygAlgorithm, cipheringAlgorithm
}
