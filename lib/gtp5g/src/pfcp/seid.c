#include <linux/string.h>
#include <linux/slab.h>

#include "seid.h"

#define SEID_HEX_STR_LEN 16
#define U32_ID_HEX_STR_LEN 8

void seid_and_u32id_to_hex_str(u64 seid_int, u32 id, char *seid_u32id_hexstr)
{  
    char seid_hexstr[SEID_HEX_STR_LEN];
    char id_hexstr[U32_ID_HEX_STR_LEN];

    snprintf(seid_hexstr, SEID_HEX_STR_LEN, "%llx", seid_int);
    snprintf(id_hexstr, U32_ID_HEX_STR_LEN, "%x", id);
    strcpy(seid_u32id_hexstr, seid_hexstr);
    strcat(seid_u32id_hexstr, id_hexstr);
}