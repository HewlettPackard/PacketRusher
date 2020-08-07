package milenage

import (
	"fmt"
	"my5G-RANTester/lib/aes"
	"reflect"
	"strconv"
)

func rtLength(keybits int) int {
	return (keybits)/8 + 28
}

func aes128EncryptBlock(key, in, out []uint8) int {
	const keyBits int = 128

	rk := make([]uint32, rtLength(128))
	var nrounds = aes.AesSetupEnc(rk, key, keyBits)
	//fmt.Printf("nrounds: %d\n", nrounds)

	aes.AesEncrypt(rk, nrounds, in, out)
	return 0

}

/*
int aes_128_encrypt_block(const c_uint8_t *key,
const c_uint8_t *in, c_uint8_t *out)
{
const int key_bits = 128;
unsigned int rk[RKLENGTH(128)];
int nrounds;

nrounds = aes_setup_enc(rk, key, key_bits);
aes_encrypt(rk, nrounds, in, out);

return 0;
}*/

/**
 * milenage_f1 - Milenage f1 and f1* algorithms
 * @opc: OPc = 128-bit value derived from OP and K
 * @k: K = 128-bit subscriber key
 * @_rand: RAND = 128-bit random challenge
 * @sqn: SQN = 48-bit sequence number
 * @amf: AMF = 16-bit authentication management field
 * @mac_a: Buffer for MAC-A = 64-bit network authentication code, or %NULL
 * @mac_s: Buffer for MAC-S = 64-bit resync authentication code, or %NULL
 * Returns: 0 on success, -1 on failure
 */
func milenageF1(opc, k, _rand, sqn, amf, mac_a, mac_s []uint8) int {

	tmp1, tmp2, tmp3 := make([]uint8, 16), make([]uint8, 16), make([]uint8, 16)
	// var tmp1, tmp2, tmp3 [16]uint8

	rijndaelInput := make([]uint8, 16)

	/* tmp1 = TEMP = E_K(RAND XOR OP_C) */
	for i := 0; i < 16; i++ {
		rijndaelInput[i] = _rand[i] ^ opc[i]
	}
	// RijndaelEncrypt( OP, op_c );
	if aes128EncryptBlock(k, rijndaelInput, tmp1) != 0 {
		return -1
	}

	// fmt.Printf("tmp1: %x\n", tmp1)

	/* tmp2 = IN1 = SQN || AMF || SQN || AMF */
	copy(tmp2[0:], sqn[0:6])
	copy(tmp2[6:], amf[0:2])
	copy(tmp2[8:], tmp2[0:8])
	/*
		os_memcpy(tmp2, sqn, 6);
		os_memcpy(tmp2 + 6, amf, 2);
		os_memcpy(tmp2 + 8, tmp2, 8);
	*/

	/* OUT1 = E_K(TEMP XOR rot(IN1 XOR OP_C, r1) XOR c1) XOR OP_C */

	/* rotate (tmp2 XOR OP_C) by r1 (= 0x40 = 8 bytes) */
	for i := 0; i < 16; i++ {
		tmp3[(i+8)%16] = tmp2[i] ^ opc[i]
	}

	// fmt.Printf("tmp3: %x\n", tmp3)

	/* XOR with TEMP = E_K(RAND XOR OP_C) */
	for i := 0; i < 16; i++ {
		tmp3[i] ^= tmp1[i]
	}
	// fmt.Printf("tmp3 XOR with TEMP: %x\n", tmp3)

	/* XOR with c1 (= ..00, i.e., NOP) */
	/* f1 || f1* = E_K(tmp3) XOR OP_c */
	if aes128EncryptBlock(k, tmp3, tmp1) != 0 {
		return -1
	}
	// fmt.Printf("XOR with c1 (: %x\n", tmp1)

	for i := 0; i < 16; i++ {
		tmp1[i] ^= opc[i]
	}
	// fmt.Printf("tmp1[i] ^= opc[i] %x\n", tmp1)
	if mac_a != nil {
		copy(mac_a[0:], tmp1[0:8])
	}

	if mac_s != nil {
		copy(mac_s[0:], tmp1[8:16])
	}

	return 0
}

/**
 * milenage_f2345 - Milenage f2, f3, f4, f5, f5* algorithms
 * @opc: OPc = 128-bit value derived from OP and K
 * @k: K = 128-bit subscriber key
 * @_rand: RAND = 128-bit random challenge
 * @res: Buffer for RES = 64-bit signed response (f2), or %NULL
 * @ck: Buffer for CK = 128-bit confidentiality key (f3), or %NULL
 * @ik: Buffer for IK = 128-bit integrity key (f4), or %NULL
 * @ak: Buffer for AK = 48-bit anonymity key (f5), or %NULL
 * @akstar: Buffer for AK = 48-bit anonymity key (f5*), or %NULL
 * Returns: 0 on success, -1 on failure
 */
func milenageF2345(opc, k, _rand, res, ck, ik, ak, akstar []uint8) int {
	tmp1, tmp2, tmp3 := make([]uint8, 16), make([]uint8, 16), make([]uint8, 16)
	//c_uint8_t tmp1[16], tmp2[16], tmp3[16];

	/* tmp2 = TEMP = E_K(RAND XOR OP_C) */
	for i := 0; i < 16; i++ {
		tmp1[i] = _rand[i] ^ opc[i]
	}

	if aes128EncryptBlock(k, tmp1, tmp2) != 0 {
		return -1
	}
	/*
		for (i = 0; i < 16; i++)
			tmp1[i] = _rand[i] ^ opc[i];
		if (aes_128_encrypt_block(k, tmp1, tmp2))
		return -1;
	*/

	/* OUT2 = E_K(rot(TEMP XOR OP_C, r2) XOR c2) XOR OP_C */
	/* OUT3 = E_K(rot(TEMP XOR OP_C, r3) XOR c3) XOR OP_C */
	/* OUT4 = E_K(rot(TEMP XOR OP_C, r4) XOR c4) XOR OP_C */
	/* OUT5 = E_K(rot(TEMP XOR OP_C, r5) XOR c5) XOR OP_C */

	/* f2 and f5 */
	/* rotate by r2 (= 0, i.e., NOP) */
	for i := 0; i < 16; i++ {
		tmp1[i] = tmp2[i] ^ opc[i]
	}
	tmp1[15] ^= 1 // XOR c2 (= ..01)
	/*
		for (i = 0; i < 16; i++)
			tmp1[i] = tmp2[i] ^ opc[i];
		tmp1[15] ^= 1; // XOR c2 (= ..01)
	*/

	/* f5 || f2 = E_K(tmp1) XOR OP_c */
	if aes128EncryptBlock(k, tmp1, tmp3) != 0 {
		return -1
	}

	for i := 0; i < 16; i++ {
		tmp3[i] ^= opc[i]
	}

	if res != nil {
		copy(res[0:], tmp3[8:16]) // f2
	}

	if ak != nil {
		copy(ak[0:], tmp3[0:6]) // f5
	}
	/*
		if (aes_128_encrypt_block(k, tmp1, tmp3))
			return -1;
		for (i = 0; i < 16; i++)
			tmp3[i] ^= opc[i];
		if (res)
			os_memcpy(res, tmp3 + 8, 8); // f2
		if (ak)
			os_memcpy(ak, tmp3, 6); // f5
	*/

	/* f3 */
	if ck != nil {
		// rotate by r3 = 0x20 = 4 bytes
		for i := 0; i < 16; i++ {
			tmp1[(i+12)%16] = tmp2[i] ^ opc[i]
		}
		tmp1[15] ^= 2 // XOR c3 (= ..02)

		if aes128EncryptBlock(k, tmp1, ck) != 0 {
			return -1
		}

		for i := 0; i < 16; i++ {
			ck[i] ^= opc[i]
		}
	}
	/*
		if (ck) {
			// rotate by r3 = 0x20 = 4 bytes
			for (i = 0; i < 16; i++)
				tmp1[(i + 12) % 16] = tmp2[i] ^ opc[i];
			tmp1[15] ^= 2; // XOR c3 (= ..02)
			if (aes_128_encrypt_block(k, tmp1, ck))
				return -1;
			for (i = 0; i < 16; i++)
				ck[i] ^= opc[i];
		}
	*/

	/* f4 */
	if ik != nil {
		//rotate by r4 = 0x40 = 8 bytes
		for i := 0; i < 16; i++ {
			tmp1[(i+8)%16] = tmp2[i] ^ opc[i]
		}
		tmp1[15] ^= 4 // XOR c4 (= ..04)

		if aes128EncryptBlock(k, tmp1, ik) != 0 {
			return -1
		}

		for i := 0; i < 16; i++ {
			ik[i] ^= opc[i]
		}
	}
	/*
		if (ik) {
			//rotate by r4 = 0x40 = 8 bytes
			for (i = 0; i < 16; i++)
				tmp1[(i + 8) % 16] = tmp2[i] ^ opc[i];
			tmp1[15] ^= 4; // XOR c4 (= ..04)
			if (aes_128_encrypt_block(k, tmp1, ik))
				return -1;
			for (i = 0; i < 16; i++)
				ik[i] ^= opc[i];
		}
	*/

	/* f5* */
	if akstar != nil {
		// rotate by r5 = 0x60 = 12 bytes
		for i := 0; i < 16; i++ {
			tmp1[(i+4)%16] = tmp2[i] ^ opc[i]
		}
		tmp1[15] ^= 8 // XOR c5 (= ..08)

		if aes128EncryptBlock(k, tmp1, tmp1) != 0 {
			return -1
		}

		for i := 0; i < 6; i++ {
			akstar[i] = tmp1[i] ^ opc[i]
		}
	}
	/*
		if (akstar) {
			// rotate by r5 = 0x60 = 12 bytes
			for (i = 0; i < 16; i++)
				tmp1[(i + 4) % 16] = tmp2[i] ^ opc[i];
			tmp1[15] ^= 8; // XOR c5 (= ..08)
			if (aes_128_encrypt_block(k, tmp1, tmp1))
				return -1;
			for (i = 0; i < 6; i++)
				akstar[i] = tmp1[i] ^ opc[i];
		}
	*/

	return 0
}

func MilenageGenerate(opc, amf, k, sqn, _rand, autn, ik, ck, ak, res []uint8, res_len *uint) {
	// var i int
	mac_a := make([]uint8, 8)

	// fmt.Println(i)
	// fmt.Println(mac_a)

	if (*res_len) < 8 {
		*res_len = 0
		return
	}

	if milenageF1(opc, k, _rand, sqn, amf, mac_a, nil) != 0 || milenageF2345(opc, k, _rand, res, ck, ik, ak, nil) != 0 {
		*res_len = 0
		return
	}

	*res_len = 8

	/* AUTN = (SQN ^ AK) || AMF || MAC */
	for i := 0; i < 6; i++ {
		autn[i] = sqn[i] ^ ak[i]
		copy(autn[6:], amf[0:2])
		copy(autn[8:], mac_a[0:8])
	}
	/*
		for (i = 0; i < 6; i++)
		autn[i] = sqn[i] ^ ak[i];
		os_memcpy(autn + 6, amf, 2);
		os_memcpy(autn + 8, mac_a, 8);
	*/
}

/**
 * milenage_auts - Milenage AUTS validation
 * @opc: OPc = 128-bit operator variant algorithm configuration field (encr.)
 * @k: K = 128-bit subscriber key
 * @_rand: RAND = 128-bit random challenge
 * @auts: AUTS = 112-bit authentication token from client
 * @sqn: Buffer for SQN = 48-bit sequence number
 * Returns: 0 = success (sqn filled), -1 on failure
 */
//int milenage_auts(const c_uint8_t *opc, const c_uint8_t *k, const c_uint8_t *_rand, const c_uint8_t *auts, c_uint8_t *sqn)
func Milenage_auts(opc, k, _rand, auts, sqn []uint8) int {
	amf := []uint8{0x00, 0x00} // TS 33.102 v7.0.0, 6.3.3
	ak := make([]uint8, 6)
	mac_s := make([]uint8, 8)
	/*
	   c_uint8_t amf[2] = { 0x00, 0x00 }; // TS 33.102 v7.0.0, 6.3.3
	   c_uint8_t ak[6], mac_s[8];
	   int i;
	*/

	if milenageF2345(opc, k, _rand, nil, nil, nil, nil, ak) != 0 {
		return -1
	}

	for i := 0; i < 6; i++ {
		sqn[i] = auts[i] ^ ak[i]
	}

	if milenageF1(opc, k, _rand, sqn, amf, nil, mac_s) != 0 || !reflect.DeepEqual(mac_s, auts[6:14]) {
		return -1
	}

	/*
		   if (milenage_f2345(opc, k, _rand, NULL, NULL, NULL, NULL, ak))
			   return -1;
		   for (i = 0; i < 6; i++)
			   sqn[i] = auts[i] ^ ak[i];
		   if (milenage_f1(opc, k, _rand, sqn, amf, NULL, mac_s) ||
			   os_memcmp_const(mac_s, auts + 6, 8) != 0)
			   return -1;
	*/
	return 0
}

/**
 * gsm_milenage - Generate GSM-Milenage (3GPP TS 55.205) authentication triplet
 * @opc: OPc = 128-bit operator variant algorithm configuration field (encr.)
 * @k: K = 128-bit subscriber key
 * @_rand: RAND = 128-bit random challenge
 * @sres: Buffer for SRES = 32-bit SRES
 * @kc: Buffer for Kc = 64-bit Kc
 * Returns: 0 on success, -1 on failure
 */
func Gsm_milenage(opc, k, _rand, sres, kc []uint8) int {
	res, ck, ik := make([]uint8, 8), make([]uint8, 16), make([]uint8, 16)

	if milenageF2345(opc, k, _rand, res, ck, ik, nil, nil) != 0 {
		return -1
	}
	/*
		if (milenage_f2345(opc, k, _rand, res, ck, ik, NULL, NULL))
			return -1;
	*/

	for i := 0; i < 8; i++ {
		kc[i] = ck[i] ^ ck[i+8] ^ ik[i] ^ ik[i+8]
	}
	/*
		for (i = 0; i < 8; i++)
			kc[i] = ck[i] ^ ck[i + 8] ^ ik[i] ^ ik[i + 8];
	*/

	// if GSM_MILENAGE_ALT_SRES
	//copy(sres, res[0:4])

	// if not GSM_MILENAGE_ALT_SRES
	for i := 0; i < 4; i++ {
		sres[i] = res[i] ^ res[i+4]
	}
	/*
		#ifdef GSM_MILENAGE_ALT_SRES
			os_memcpy(sres, res, 4);
		#else // GSM_MILENAGE_ALT_SRES
		for (i = 0; i < 4; i++)
			sres[i] = res[i] ^ res[i + 4];
		#endif // GSM_MILENAGE_ALT_SRES
	*/
	return 0
}

/**
 * milenage_generate - Generate AKA AUTN,IK,CK,RES
 * @opc: OPc = 128-bit operator variant algorithm configuration field (encr.)
 * @k: K = 128-bit subscriber key
 * @sqn: SQN = 48-bit sequence number
 * @_rand: RAND = 128-bit random challenge
 * @autn: AUTN = 128-bit authentication token
 * @ik: Buffer for IK = 128-bit integrity key (f4), or %NULL
 * @ck: Buffer for CK = 128-bit confidentiality key (f3), or %NULL
 * @res: Buffer for RES = 64-bit signed response (f2), or %NULL
 * @res_len: Variable that will be set to RES length
 * @auts: 112-bit buffer for AUTS
 * Returns: 0 on success, -1 on failure, or -2 on synchronization failure
 */
func Milenage_check(opc, k, sqn, _rand, autn, ik, ck, res []uint8, res_len *uint, auts []uint8) int {

	mac_a, ak, rx_sqn := make([]uint8, 8), make([]uint8, 6), make([]uint8, 6)
	var amf []uint8

	// fmt.Println(mac_a, amf)

	/* TODO
	d_trace(1, "Milenage: AUTN\n"); d_trace_hex(1, autn, 16);
	d_trace(1, "Milenage: RAND\n"); d_trace_hex(1, _rand, 16);
	*/

	if milenageF2345(opc, k, _rand, res, ck, ik, ak, nil) != 0 {
		return -1
	}
	/*
		if (milenage_f2345(opc, k, _rand, res, ck, ik, ak, NULL))
			return -1;
	*/

	*res_len = 8
	/* TODO
	d_trace(1, "Milenage: RES\n"); d_trace_hex(1, res, *res_len);
	d_trace(1, "Milenage: CK\n"); d_trace_hex(1, ck, 16);
	d_trace(1, "Milenage: IK\n"); d_trace_hex(1, ik, 16);
	d_trace(1, "Milenage: AK\n"); d_trace_hex(1, ak, 6);
	*/

	/* AUTN = (SQN ^ AK) || AMF || MAC */
	for i := 0; i < 6; i++ {
		rx_sqn[i] = autn[i] ^ ak[i]
	}
	/*
		for (i = 0; i < 6; i++)
			rx_sqn[i] = autn[i] ^ ak[i];
	*/

	//TODO d_trace(1, "Milenage: SQN\n"); d_trace_hex(1, rx_sqn, 6);

	if os_memcmp(rx_sqn, sqn, 6) <= 0 {
		auts_amf := []uint8{0x00, 0x00} // TS 33.102 v7.0.0, 6.3.3

		if milenageF2345(opc, k, _rand, nil, nil, nil, nil, ak) != 0 {
			return -1
		}

		//TODO d_trace(1, "Milenage: AK*\n"); d_trace_hex(1, ak, 6);

		for i := 0; i < 6; i++ {
			auts[i] = sqn[i] ^ ak[i]
		}

		if milenageF1(opc, k, _rand, sqn, auts_amf, nil, auts[6:]) != 0 {
			return -1
		}

		// TODO d_trace(1, "Milenage: AUTS*\n"); d_trace_hex(1, auts, 14);

		return -2
	}
	/*
		if (os_memcmp(rx_sqn, sqn, 6) <= 0) {
			c_uint8_t auts_amf[2] = { 0x00, 0x00 }; // TS 33.102 v7.0.0, 6.3.3
			if (milenage_f2345(opc, k, _rand, NULL, NULL, NULL, NULL, ak))
			return -1;
			d_trace(1, "Milenage: AK*\n"); d_trace_hex(1, ak, 6);



			for (i = 0; i < 6; i++)
				auts[i] = sqn[i] ^ ak[i];
			if (milenage_f1(opc, k, _rand, sqn, auts_amf, NULL, auts + 6))
				return -1;
			d_trace(1, "Milenage: AUTS*\n"); d_trace_hex(1, auts, 14);
		return -2;
		}
	*/

	amf = autn[6:]
	//TODO d_trace(1, "Milenage: AMF\n"); d_trace_hex(1, amf, 2);

	if milenageF1(opc, k, _rand, rx_sqn, amf, mac_a, nil) != 0 {
		return -1
	}
	//TODO d_trace(1, "Milenage: MAC_A\n"); d_trace_hex(1, mac_a, 8);

	if os_memcmp(mac_a, autn[8:], 8) != 0 {
		//TODO d_trace(1, "Milenage: MAC mismatch\n");
		//TODO d_trace(1, "Milenage: Received MAC_A\n"); d_trace_hex(1, autn + 8, 8);

		return -1
	}
	/*
		amf = autn + 6;
		d_trace(1, "Milenage: AMF\n"); d_trace_hex(1, amf, 2);
		if (milenage_f1(opc, k, _rand, rx_sqn, amf, mac_a, NULL))
			return -1;

		d_trace(1, "Milenage: MAC_A\n"); d_trace_hex(1, mac_a, 8);

		if (os_memcmp_const(mac_a, autn + 8, 8) != 0) {
			d_trace(1, "Milenage: MAC mismatch\n");
			d_trace(1, "Milenage: Received MAC_A\n"); d_trace_hex(1, autn + 8, 8);
			return -1;
		}
	*/

	return 0
}

func milenage_opc(k, op, opc []uint8) {

	// RijndaelEncrypt( OP, op_c );
	aes128EncryptBlock(k, op, opc)

	for i := 0; i < 16; i++ {
		opc[i] ^= op[i]
	}
}

// implementation of os_memcmp
func os_memcmp(a, b []uint8, num int) int {
	for i := 0; i < num; i++ {
		if a[i] < b[i] {
			return -i
		}

		if a[i] > b[i] {
			return i
		}
	}

	return 0
}

func F1_Test(opc, k, _rand, sqn, amf, mac_a, mac_s []uint8) int {

	res := milenageF1(opc, k, _rand, sqn, amf, mac_a, mac_s)

	return res
}

func F2345_Test(opc, k, _rand, res, ck, ik, ak, akstar []uint8) int {

	flag := milenageF2345(opc, k, _rand, res, ck, ik, ak, akstar)

	return flag
}

func GenerateOPC(k, op, opc []uint8) {
	milenage_opc(k, op, opc)
}

func InsertData(op, k, _rand, sqn, amf []uint8, OP, K, RAND, SQN, AMF string) {

	var res uint64
	var err error

	// load op
	// fmt.Print("OP: ")
	for i := 0; i < 16; i++ {
		res, err = strconv.ParseUint(OP[i*2:i*2+2], 16, 8)

		if err == nil {
			op[i] = uint8(res)
			// fmt.Printf("%02x ", op[i])
		}
	}

	// fmt.Println()
	fmt.Printf("OP: %x\n", op)

	// load k
	// fmt.Print("K: ")
	for i := 0; i < 16; i++ {
		res, err = strconv.ParseUint(K[i*2:i*2+2], 16, 8)

		if err == nil {
			k[i] = uint8(res)
			// fmt.Printf("%02x ", k[i])
		}
	}

	// fmt.Println()
	fmt.Printf("K: %x\n", k)

	// load _rand
	// fmt.Print("RAND: ")
	for i := 0; i < 16; i++ {
		res, err = strconv.ParseUint(RAND[i*2:i*2+2], 16, 8)

		if err == nil {
			_rand[i] = uint8(res)
			// fmt.Printf("%02x ", _rand[i])
		}
	}

	// fmt.Println()
	// load sqn
	// fmt.Print("SQN: ")
	for i := 0; i < 6; i++ {
		res, err = strconv.ParseUint(SQN[i*2:i*2+2], 16, 8)

		if err == nil {
			sqn[i] = uint8(res)
			// fmt.Printf("%02x ", sqn[i])
		}
	}

	// fmt.Println()
	fmt.Printf("SQN: %x\n", sqn)

	// load amf
	// fmt.Print("AMF: ")
	for i := 0; i < 2; i++ {
		res, err = strconv.ParseUint(AMF[i*2:i*2+2], 16, 8)

		if err == nil {
			amf[i] = uint8(res)
			// fmt.Printf("%02x ", amf[i])
		}
	}

	// fmt.Println()
	fmt.Printf("AMF: %x\n", amf)
	fmt.Println()

}
