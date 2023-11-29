/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package tools

import (
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/rand"
	"time"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/milenage"
	"github.com/free5gc/util/ueauth"
	log "github.com/sirupsen/logrus"
)

// might move to nasMsgHandler.authenticationRequest if not used in other requests
func AuthProcedure(authSub models.AuthenticationSubscription, servingNetworkName string) (
	response *models.AuthenticationInfoResult, problemDetails *models.ProblemDetails,
) {
	response = &models.AuthenticationInfoResult{}

	rand.Seed(time.Now().UnixNano())
	RAND := make([]byte, 16)
	_, err := cryptoRand.Read(RAND)

	opc, err := hex.DecodeString(authSub.Opc.OpcValue)
	if err != nil {
		log.Error("[AMF] err while decoding opcStr: ", err)
	}
	k, err := hex.DecodeString(authSub.PermanentKey.PermanentKeyValue)
	if err != nil {
		log.Error("[AMF] err while decoding kStr: ", err)
	}

	// TODO: Improve?
	sqnString := "000001c7feae"
	sqn, err := hex.DecodeString(sqnString)
	if err != nil {
		log.Error("[AMF] err while decoding sqnStr: ", err)
	}
	amf, err := hex.DecodeString(authSub.AuthenticationManagementField)
	if err != nil {
		log.Error("[AMF] err while decoding amfStr: ", err)
	}

	// Run milenage
	macA, macS := make([]byte, 8), make([]byte, 8)
	CK, IK := make([]byte, 16), make([]byte, 16)
	RES := make([]byte, 8)
	AK, AKstar := make([]byte, 6), make([]byte, 6)

	// Generate macA, macS
	err = milenage.F1(opc, k, RAND, sqn, amf, macA, macS)
	if err != nil {
		log.Error("milenage F1 err:", err)
	}

	// Generate RES, CK, IK, AK, AKstar
	// RES == XRES (expected RES) for server
	err = milenage.F2345(opc, k, RAND, RES, CK, IK, AK, AKstar)
	if err != nil {
		log.Error("milenage F2345 err:", err)
	}

	// Generate AUTN
	SQNxorAK := make([]byte, 6)
	for i := 0; i < len(sqn); i++ {
		SQNxorAK[i] = sqn[i] ^ AK[i]
	}
	AUTN := append(append(SQNxorAK, amf...), macA...)

	var av models.AuthenticationVector

	response.AuthType = models.AuthType__5_G_AKA

	// derive XRES*
	key := append(CK, IK...)
	FC := ueauth.FC_FOR_RES_STAR_XRES_STAR_DERIVATION
	P0 := []byte(servingNetworkName)
	P1 := RAND
	P2 := RES

	kdfValForXresStar, err := ueauth.GetKDFValue(
		key, FC, P0, ueauth.KDFLen(P0), P1, ueauth.KDFLen(P1), P2, ueauth.KDFLen(P2))
	if err != nil {
		log.Errorf("Get kdfValForXresStar err: %+v", err)
	}
	xresStar := kdfValForXresStar[len(kdfValForXresStar)/2:]

	// derive Kausf
	FC = ueauth.FC_FOR_KAUSF_DERIVATION
	P0 = []byte(servingNetworkName)
	P1 = SQNxorAK
	kdfValForKausf, err := ueauth.GetKDFValue(key, FC, P0, ueauth.KDFLen(P0), P1, ueauth.KDFLen(P1))

	// strKausf := hex.EncodeToString(kdfValForKausf)
	// log.Info("[AMF] Kausf: " + strKausf)
	if err != nil {
		log.Errorf("Get kdfValForKausf err: %+v", err)
	}

	// Fill in rand, xresStar, autn, kausf
	av.Rand = hex.EncodeToString(RAND)
	av.XresStar = hex.EncodeToString(xresStar)
	av.Autn = hex.EncodeToString(AUTN)
	av.Kausf = hex.EncodeToString(kdfValForKausf)
	av.AvType = models.AvType__5_G_HE_AKA

	response.AuthenticationVector = &av
	return response, nil
}

func DeriveHXRES(auth *models.AuthenticationInfoResult, servingNetworkName string) (models.UeAuthenticationCtx, string, error) {
	authCtx := models.UeAuthenticationCtx{}
	// Derive HXRES* from XRES*
	concat := auth.AuthenticationVector.Rand + auth.AuthenticationVector.XresStar
	var hxresStarBytes []byte
	if bytes, err := hex.DecodeString(concat); err != nil {
		return authCtx, "", errors.New("[AMF] Error while decoding hxresStar")
	} else {
		hxresStarBytes = bytes
	}
	hxresStarAll := sha256.Sum256(hxresStarBytes)
	hxresStar := hex.EncodeToString(hxresStarAll[16:]) // last 128 bits

	// Derive Kseaf from Kausf
	Kausf := auth.AuthenticationVector.Kausf
	var KausfDecode []byte
	ausfDecode, err := hex.DecodeString(Kausf)
	if err != nil {
		return authCtx, "", errors.New("[AMF] decode Kausf failed: %+v" + err.Error())
	}
	KausfDecode = ausfDecode

	P0 := []byte(servingNetworkName)
	Kseaf, err := ueauth.GetKDFValue(KausfDecode, ueauth.FC_FOR_KSEAF_DERIVATION, P0, ueauth.KDFLen(P0))

	var av5gAka models.Av5gAka
	av5gAka.Rand = auth.AuthenticationVector.Rand
	av5gAka.Autn = auth.AuthenticationVector.Autn
	av5gAka.HxresStar = hxresStar

	authCtx.Var5gAuthData = av5gAka
	authCtx.AuthType = auth.AuthType
	return authCtx, hex.EncodeToString(Kseaf), nil
}
