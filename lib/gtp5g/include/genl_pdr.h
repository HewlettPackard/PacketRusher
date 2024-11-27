#ifndef __GENL_PDR_H__
#define __GENL_PDR_H__

#include "genl.h"

enum gtp5g_pdr_attrs {
    /* gtp5g_device_attrs in this part */

    GTP5G_PDR_ID = 3,
    GTP5G_PDR_PRECEDENCE,
    GTP5G_PDR_PDI,
    GTP5G_OUTER_HEADER_REMOVAL,
    GTP5G_PDR_FAR_ID,

    /* Not in 3GPP spec, just used for routing */
    GTP5G_PDR_ROLE_ADDR_IPV4,

    /* Not in 3GPP spec, just used for buffering */
    GTP5G_PDR_UNIX_SOCKET_PATH,

    GTP5G_PDR_QER_ID,

    GTP5G_PDR_SEID,
    GTP5G_PDR_URR_ID,
    /* Add newly supported feature ON ABOVE
     * for compatability with older version of
     * free5GC's UPF or libgtp5gnl
     * */

    __GTP5G_PDR_ATTR_MAX,
};
#define GTP5G_PDR_ATTR_MAX 16

/* Nest in GTP5G_PDR_PDI */
enum gtp5g_pdi_attrs {
    GTP5G_PDI_UNSPEC,
    GTP5G_PDI_UE_ADDR_IPV4,
    GTP5G_PDI_F_TEID,
    GTP5G_PDI_SDF_FILTER,
    GTP5G_PDI_SRC_INTF,

    __GTP5G_PDI_ATTR_MAX,
};
#define GTP5G_PDI_ATTR_MAX 16

/* Nest in GTP5G_PDI_F_TEID */
enum gtp5g_f_teid_attrs {
    GTP5G_F_TEID_UNSPEC,
    GTP5G_F_TEID_I_TEID,
    GTP5G_F_TEID_GTPU_ADDR_IPV4,
    __GTP5G_F_TEID_ATTR_MAX,
};
#define GTP5G_F_TEID_ATTR_MAX 8

/* Nest in GTP5G_PDI_SDF_FILTER */
enum gtp5g_sdf_filter_attrs {
    GTP5G_SDF_FILTER_FLOW_DESCRIPTION = 1,
    GTP5G_SDF_FILTER_TOS_TRAFFIC_CLASS,
    GTP5G_SDF_FILTER_SECURITY_PARAMETER_INDEX,
    GTP5G_SDF_FILTER_FLOW_LABEL,
    GTP5G_SDF_FILTER_SDF_FILTER_ID,

    __GTP5G_SDF_FILTER_ATTR_MAX,
};
#define GTP5G_SDF_FILTER_ATTR_MAX 16

/* Nest in GTP5G_SDF_FILTER_FLOW_DESCRIPTION */
enum gtp5g_flow_description_attrs {
    GTP5G_FLOW_DESCRIPTION_ACTION = 1, // Only "permit"
    GTP5G_FLOW_DESCRIPTION_DIRECTION,
    GTP5G_FLOW_DESCRIPTION_PROTOCOL,
    GTP5G_FLOW_DESCRIPTION_SRC_IPV4,
    GTP5G_FLOW_DESCRIPTION_SRC_MASK,
    GTP5G_FLOW_DESCRIPTION_DEST_IPV4,
    GTP5G_FLOW_DESCRIPTION_DEST_MASK,
    GTP5G_FLOW_DESCRIPTION_SRC_PORT,
    GTP5G_FLOW_DESCRIPTION_DEST_PORT,

    __GTP5G_FLOW_DESCRIPTION_ATTR_MAX,
};
#define GTP5G_FLOW_DESCRIPTION_ATTR_MAX (__GTP5G_FLOW_DESCRIPTION_ATTR_MAX - 1)

enum {
    GTP5G_SDF_FILTER_ACTION_UNSPEC = 0,
    GTP5G_SDF_FILTER_PERMIT,
    __GTP5G_SDF_FILTER_ACTION_MAX,
};
#define GTP5G_SDF_FILTER_ACTION_MAX (__GTP5G_SDF_FILTER_ACTION_MAX - 1)

enum {
    GTP5G_SDF_FILTER_DIRECTION_UNSPEC = 0,
    GTP5G_SDF_FILTER_IN,
    GTP5G_SDF_FILTER_OUT,
    __GTP5G_SDF_FILTER_DIRECTION_MAX,
};
#define GTP5G_SDF_FILTER_DIRECTION_MAX (__GTP5G_SDF_FILTER_DIRECTION_MAX - 1)

int gtp5g_genl_add_pdr(struct sk_buff *, struct genl_info *);
int gtp5g_genl_del_pdr(struct sk_buff *, struct genl_info *);
int gtp5g_genl_get_pdr(struct sk_buff *, struct genl_info *);
int gtp5g_genl_dump_pdr(struct sk_buff *, struct netlink_callback *);

#endif // __GENL_PDR_H__
