/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2025 Free Mobile SAS
 */
package sidf

import (
	"crypto/ecdh"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

type HomeNetworkPublicKey struct {
	ProtectionScheme string
	PublicKey        *ecdh.PublicKey
	PublicKeyID      string
}

func profileAEncrypt(msin string, hnPubkey *ecdh.PublicKey) (string, error) {
	// Profile A curve
	x25519Curve := ecdh.X25519()

	// The UE generates an ephemeral key to transmit its SUPI to network
	ephemeralPriv, err := x25519Curve.GenerateKey(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate ephemeral X25519 key: %w", err)
	}
	ephemeralPub := ephemeralPriv.PublicKey().Bytes()

	// ECDH between UE's ephemeral key and Home Network Public Key
	sharedKey, err := ephemeralPriv.ECDH(hnPubkey)
	if err != nil {
		return "", fmt.Errorf("failed to compute ECDH: %w", err)
	}

	plainBCD, err := hex.DecodeString(Tbcd(msin))
	if err != nil {
		return "", err
	}

	kdfKey := AnsiX963KDF(sharedKey, ephemeralPub, ProfileAEncKeyLen, ProfileAMacKeyLen, ProfileAHashLen)
	encKey := kdfKey[:ProfileAEncKeyLen]
	iv := kdfKey[ProfileAEncKeyLen : ProfileAEncKeyLen+ProfileAIcbLen]
	macKey := kdfKey[len(kdfKey)-ProfileAMacKeyLen:]

	cipherText, err := Aes128ctr(plainBCD, encKey, iv)
	if err != nil {
		return "", err
	}

	mac, err := HmacSha256(cipherText, macKey, ProfileAMacLen)
	if err != nil {
		return "", err
	}

	// UE's ephemeral public key || ciphered(MSIN || iv) || MAC.
	out := append(ephemeralPub, cipherText...)
	out = append(out, mac...)

	return hex.EncodeToString(out), nil
}

func profileBEncrypt(msin string, hnPubkey *ecdh.PublicKey) (string, error) {
	// Profile B curve
	p256Curve := ecdh.P256()

	// The UE generates an ephemeral key to transmit its SUPI to network
	ephemeralPriv, err := p256Curve.GenerateKey(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate ephemeral P256 key: %w", err)
	}

	// ECDH between UE's ephemeral key and Home Network Public Key
	sharedKey, err := ephemeralPriv.ECDH(hnPubkey)
	if err != nil {
		return "", fmt.Errorf("failed to compute ECDH: %w", err)
	}

	// For the KDF we need the ephemeral public key in compressed form
	x, y := elliptic.Unmarshal(elliptic.P256(), ephemeralPriv.PublicKey().Bytes())
	if x == nil || y == nil {
		return "", errors.New("failed to unmarshal ephemeral public key")
	}
	ephemeralPubCompressed := elliptic.MarshalCompressed(elliptic.P256(), x, y)

	plainBCD, err := hex.DecodeString(Tbcd(msin))
	if err != nil {
		return "", err
	}

	kdfKey := AnsiX963KDF(sharedKey, ephemeralPubCompressed, ProfileBEncKeyLen, ProfileBMacKeyLen, ProfileBHashLen)
	encKey := kdfKey[:ProfileBEncKeyLen]
	iv := kdfKey[ProfileBEncKeyLen : ProfileBEncKeyLen+ProfileBIcbLen]
	macKey := kdfKey[len(kdfKey)-ProfileBMacKeyLen:]

	cipherText, err := Aes128ctr(plainBCD, encKey, iv)
	if err != nil {
		return "", err
	}

	mac, err := HmacSha256(cipherText, macKey, ProfileBMacLen)
	if err != nil {
		return "", err
	}

	// ephemeral public key || ciphertext || MAC
	out := append(ephemeralPubCompressed, cipherText...)
	out = append(out, mac...)

	return hex.EncodeToString(out), nil
}

func CipherSuci(msin, mcc, mnc string, routingIndicator string, profile HomeNetworkPublicKey) (*Suci, error) {
	if len(msin)+len(mcc)+len(mnc) < 14 {
		return nil, errors.New("supi length must be 15")
	}

	var schemeOutput string
	var err error

	switch profile.ProtectionScheme {
	case NullScheme:
		schemeOutput = msin
	case ProfileAScheme:
		schemeOutput, err = profileAEncrypt(msin, profile.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("profile A encryption failed: %w", err)
		}
	case ProfileBScheme:
		schemeOutput, err = profileBEncrypt(msin, profile.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("profile B encryption failed: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported protection scheme: %s", profile.ProtectionScheme)
	}

	// suci-<supi_type>-<MCC>-<MNC>-<routing_indicator>-<protection_scheme>-<public_key_id>-<scheme_output>
	suci := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%s-%s",
		PrefixSUCI,
		SupiTypeIMSI,
		mcc,
		mnc,
		routingIndicator,
		profile.ProtectionScheme,
		profile.PublicKeyID,
		schemeOutput,
	)
	return ParseSuci(suci), nil
}
