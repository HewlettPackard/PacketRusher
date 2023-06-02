#ifndef __GTP5G_GTP_H__
#define __GTP5G_GTP_H__

#include <linux/skbuff.h>

#define GTPV1 0x30

/* gtpv1_hdr flags */
#define GTPV1_HDR_FLG_NPDU    0x01
#define GTPV1_HDR_FLG_SEQ     0x02
#define GTPV1_HDR_FLG_EXTHDR  0x04
#define GTPV1_HDR_FLG_MASK    0x07

/* According to 3GPP TS 29.060. */
struct gtpv1_hdr {
    __u8    flags;
    __u8    type;
    __be16  length;
    __be32  tid;
} __attribute__((packed));

typedef struct gtp1_hdr_opt {
    __be16  seq_number;
    __u8    NPDU;
    __u8    next_ehdr_type;
} __attribute__((packed)) gtpv1_hdr_opt_t;

struct recovery {
    __u8  type_num;
    __u8  cnt;
}__attribute__((packed));

struct gtpv1_echo_resp {
    struct  gtpv1_hdr    gtpv1_h;
    struct  gtp1_hdr_opt gtpv1_opt_h;
    struct  recovery     recov;
} __attribute__((packed));

/** 3GPP TS 29.281
 * From Figure 5.2.1-2 Definition of Extension Header Type
 */
#define GTPV1_NEXT_EXT_HDR_TYPE_00      0x00 /* No More extension */
#define GTPV1_NEXT_EXT_HDR_TYPE_03      0x03 /* Long PDCP PDU Number */
#define GTPV1_NEXT_EXT_HDR_TYPE_20      0x20 /* Service Class Indicator */
#define GTPV1_NEXT_EXT_HDR_TYPE_40      0x40 /* UDP Port */
#define GTPV1_NEXT_EXT_HDR_TYPE_81      0x81 /* RAN Container */
#define GTPV1_NEXT_EXT_HDR_TYPE_82      0x82 /* Long PDCP PDU Number */
#define GTPV1_NEXT_EXT_HDR_TYPE_83      0x83 /* Xw RAN Container */
#define GTPV1_NEXT_EXT_HDR_TYPE_84      0x84 /* NR RAN Container */
#define GTPV1_NEXT_EXT_HDR_TYPE_85      0x85 /* PDU Session Container */
#define GTPV1_NEXT_EXT_HDR_TYPE_C0      0xc0 /* PDCP PDU Number */

#define GTP1U_PORT  2152

#define GTPV1_MSG_TYPE_ECHO_REQ  1
#define GTPV1_MSG_TYPE_ECHO_RSP  2
#define GTPV1_MSG_TYPE_EMARK     254
#define GTPV1_MSG_TYPE_TPDU      255

#define GTPV1_IE_RECOVERY  14

typedef struct ul_pdu_sess_info {
        __u8    spare_qfi;                      /* Spare(2b) + qfi(6b)*/
} __attribute__((packed)) ul_pdu_sess_info_t;

typedef struct dl_pdu_sess_info {
        __u8    ppp_rqi_qfi;            /* ppp(1b) + rqi(1b) + qfi(6) */
} __attribute__((packed)) dl_pdu_sess_info_t;

typedef struct dl_pdu_sess_info_ppi {
        __u8    ppi_spare;                      /* ppi(3b) + spare(5b) */
        __u8    padding[3];
} __attribute__((packed)) dl_pdu_sess_info_ppi_t;

typedef struct pdu_sess_ctr {
    __u8 type_spare;                        /* type(4b) + spare(4b) */
#define PDU_SESSION_INFO_TYPE0  0x00
#define PDU_SESSION_INFO_TYPE1  0x10
    union {
        ul_pdu_sess_info_t ul;
        dl_pdu_sess_info_t dl;
    } u;
    //dl_pdu_sess_info_ppi_t dl_ppi[0];
} __attribute__((packed)) pdu_sess_ctr_t;

typedef struct gtp1_hdr_ext_pdu_sess_ctr {
    __u8            length;
    pdu_sess_ctr_t  pdu_sess_ctr;
    __u8            next_ehdr_type;
} __attribute__((packed)) ext_pdu_sess_ctr_t;

#endif // __GTP5G_GTP_H__