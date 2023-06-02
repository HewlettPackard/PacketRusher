#ifndef __GENL_QER_H__
#define __GENL_QER_H__

#include "genl.h"

enum gtp5g_qer_attrs {
    /* gtp5g_device_attrs in this part */

    GTP5G_QER_ID = 3,
    GTP5G_QER_GATE,
    GTP5G_QER_MBR,
    GTP5G_QER_GBR,
    GTP5G_QER_CORR_ID,
    GTP5G_QER_RQI,
    GTP5G_QER_QFI,
    GTP5G_QER_PPI,
    GTP5G_QER_RCSR,

    /* Not IEs in 3GPP Spec, for other purpose */
    GTP5G_QER_RELATED_TO_PDR,

    GTP5G_QER_SEID,
    __GTP5G_QER_ATTR_MAX,
};
#define GTP5G_QER_ATTR_MAX (__GTP5G_QER_ATTR_MAX - 1)

/* Nest in GTP5G_QER_MBR */
enum gtp5g_mbr_attrs {
    GTP5G_QER_MBR_UL_HIGH32 = 1,
    GTP5G_QER_MBR_UL_LOW8,
    GTP5G_QER_MBR_DL_HIGH32,
    GTP5G_QER_MBR_DL_LOW8,

    __GTP5G_QER_MBR_ATTR_MAX,
};
#define GTP5G_QER_MBR_ATTR_MAX (__GTP5G_QER_MBR_ATTR_MAX - 1)

/* Nest in GTP5G_QER_QBR */
enum gtp5g_qer_gbr_attrs {
    GTP5G_QER_GBR_UL_HIGH32 = 1,
    GTP5G_QER_GBR_UL_LOW8,
    GTP5G_QER_GBR_DL_HIGH32,
    GTP5G_QER_GBR_DL_LOW8,

    __GTP5G_QER_GBR_ATTR_MAX,
};
#define GTP5G_QER_GBR_ATTR_MAX (__GTP5G_QER_GBR_ATTR_MAX - 1)

extern int gtp5g_genl_add_qer(struct sk_buff *, struct genl_info *);
extern int gtp5g_genl_del_qer(struct sk_buff *, struct genl_info *);
extern int gtp5g_genl_get_qer(struct sk_buff *, struct genl_info *);
extern int gtp5g_genl_dump_qer(struct sk_buff *, struct netlink_callback *);

#endif // __GENL_QER_H__