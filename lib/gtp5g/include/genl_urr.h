#ifndef __GENL_URR_H__
#define __GENL_URR_H__

#include "genl.h"

enum gtp5g_urr_attrs {
    GTP5G_URR_ID = 3,
    GTP5G_URR_MEASUREMENT_METHOD,
    GTP5G_URR_REPORTING_TRIGGER,
    GTP5G_URR_MEASUREMENT_PERIOD,
    GTP5G_URR_MEASUREMENT_INFO,
    GTP5G_URR_SEID,

    GTP5G_URR_VOLUME_THRESHOLD,
    GTP5G_URR_VOLUME_QUOTA,
    GTP5G_URR_MULTI_SEID_URRID,
    GTP5G_URR_NUM,

    /* Not IEs in 3GPP Spec, for other purpose */
    GTP5G_URR_RELATED_TO_PDR,

    __GTP5G_URR_ATTR_MAX,
};
#define GTP5G_URR_ATTR_MAX (__GTP5G_URR_ATTR_MAX - 1)

/* Nest in GTP5G_URR_VOL_THRESHOLD */
enum gtp5g_urr_volume_threshold_attrs {
    GTP5G_URR_VOLUME_THRESHOLD_FLAG = 1,

    GTP5G_URR_VOLUME_THRESHOLD_TOVOL,
    GTP5G_URR_VOLUME_THRESHOLD_UVOL,
    GTP5G_URR_VOLUME_THRESHOLD_DVOL,

    __GTP5G_URR_VOLUME_THRESHOLD_ATTR_MAX,
};
#define GTP5G_URR_VOLUME_THRESHOLD_ATTR_MAX (__GTP5G_URR_VOLUME_THRESHOLD_ATTR_MAX - 1)

/* Nest in GTP5G_URR_VOL_QUOTA */
enum gtp5g_urr_volume_quota_attrs {
    GTP5G_URR_VOLUME_QUOTA_FLAG = 1,

    GTP5G_URR_VOLUME_QUOTA_TOVOL,
    GTP5G_URR_VOLUME_QUOTA_UVOL,
    GTP5G_URR_VOLUME_QUOTA_DVOL,

    __GTP5G_URR_VOLUME_QUOTA_ATTR_MAX,
};
#define GTP5G_URR_VOLUME_QUOTA_ATTR_MAX (__GTP5G_URR_VOLUME_QUOTA_ATTR_MAX - 1)

struct seid_urr {
    u64 seid;
    u32 urrid;
};

extern int gtp5g_genl_add_urr(struct sk_buff *, struct genl_info *);
extern int gtp5g_genl_del_urr(struct sk_buff *, struct genl_info *);
extern int gtp5g_genl_get_urr(struct sk_buff *, struct genl_info *);
extern int gtp5g_genl_dump_urr(struct sk_buff *, struct netlink_callback *);

#endif // __GENL_URR_H__
