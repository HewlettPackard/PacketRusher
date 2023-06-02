#ifndef __GENL_REPORT_H__
#define __GENL_REPORT_H__

#include "genl.h"
#include "urr.h"

/* Nest in GTP5G_UR_VOLUME_MEASUREMENT */
enum gtp5g_usage_report_volume_measurement_attrs {
    GTP5G_UR_VOLUME_MEASUREMENT_FLAGS = 1,

    GTP5G_UR_VOLUME_MEASUREMENT_TOVOL,
    GTP5G_UR_VOLUME_MEASUREMENT_UVOL,
    GTP5G_UR_VOLUME_MEASUREMENT_DVOL,

    GTP5G_UR_VOLUME_MEASUREMENT_TOPACKET,
    GTP5G_UR_VOLUME_MEASUREMENT_UPACKET,
    GTP5G_UR_VOLUME_MEASUREMENT_DPACKET,

    __GTP5G_UR_VOLUME_MEASUREMENT_ATTR_MAX,
};
#define GTP5G_UR_VOLUME_MEASUREMENT_ATTR_MAX (__GTP5G_UR_VOLUME_MEASUREMENT_ATTR_MAX - 1)

enum gtp5g_usage_report_attrs {
    GTP5G_UR_URRID = 3,
    GTP5G_UR_USAGE_REPORT_TRIGGER,
    GTP5G_UR_URSEQN,
    GTP5G_UR_VOLUME_MEASUREMENT,
    GTP5G_UR_QUERY_URR_REFERENCE,
    GTP5G_UR_START_TIME,
    GTP5G_UR_END_TIME,
    GTP5G_UR_SEID,

    __GTP5G_UR_ATTR_MAX,
};
#define GTP5G_UR_ATTR_MAX (__GTP5G_UR_ATTR_MAX - 1)

enum gtp5g_multi_usage_report_attrs {
    GTP5G_UR = 5,

    __GTP5G_URS_ATTR_MAX,
};

extern int gtp5g_genl_get_usage_report(struct sk_buff *, struct genl_info *);
extern int gtp5g_genl_get_multi_usage_reports(struct sk_buff *, struct genl_info *);

extern int gtp5g_genl_fill_ur(struct sk_buff *, struct usage_report *);
extern int gtp5g_genl_fill_usage_report(struct sk_buff *, u32, u32, u32, struct usage_report *);
extern int gtp5g_genl_fill_multi_usage_reports(struct sk_buff *, u32, u32, u32, struct usage_report **, u32);

extern void convert_urr_to_report(struct urr *, struct usage_report *);
#endif // __GENL_URR_H__
