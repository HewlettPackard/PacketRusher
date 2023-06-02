#ifndef __GTP5G_GENL_H__
#define __GTP5G_GENL_H__

#include <net/genetlink.h>

enum gtp5g_cmd {
    GTP5G_CMD_UNSPEC = 0,

    GTP5G_CMD_ADD_PDR,
    GTP5G_CMD_ADD_FAR,
    GTP5G_CMD_ADD_QER,

    GTP5G_CMD_DEL_PDR,
    GTP5G_CMD_DEL_FAR,
    GTP5G_CMD_DEL_QER,

    GTP5G_CMD_GET_PDR,
    GTP5G_CMD_GET_FAR,
    GTP5G_CMD_GET_QER,

    GTP5G_CMD_ADD_URR,
    GTP5G_CMD_ADD_BAR,
    GTP5G_CMD_DEL_URR,
    GTP5G_CMD_DEL_BAR,
    GTP5G_CMD_GET_URR,
    GTP5G_CMD_GET_BAR,

    /* Add newly supported feature ON ABOVE
     * for compatability with older version of
     * free5GC's UPF or libgtp5gnl
     * */

    GTP5G_CMD_GET_VERSION,

    GTP5G_CMD_GET_REPORT,

    GTP5G_CMD_BUFFER_GTPU,

    GTP5G_CMD_GET_MULTI_REPORTS,

    __GTP5G_CMD_MAX,
};
#define GTP5G_CMD_MAX (__GTP5G_CMD_MAX - 1)

/* This const value need to bigger than the Layer 1 attr size,
 * like GTP5G_PDR_ATTR_MAX and GTP5G_FAR_ATTR_MAX
 */
#define GTP5G_ATTR_MAX 0x10

enum gtp5g_device_attrs {
    GTP5G_LINK = 1,
    GTP5G_NET_NS_FD,
};

enum gtp5g_multicast_groups {
	GTP5G_GENL_MCGRP,
};

extern struct genl_family gtp5g_genl_family;

#endif // __GTP5G_GENL_H__
