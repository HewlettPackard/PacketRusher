#ifndef __GTP5G_LINK_H__
#define __GTP5G_LINK_H__

#include <net/rtnetlink.h>

enum {
    IFLA_GTP5G_UNSPEC,

    IFLA_GTP5G_FD1,
    IFLA_GTP5G_PDR_HASHSIZE,
    IFLA_GTP5G_ROLE,

    __IFLA_GTP5G_MAX,
};
#define IFLA_GTP5G_MAX (__IFLA_GTP5G_MAX - 1)
/* end of part */

enum ifla_gtp5g_role {
    GTP5G_ROLE_UPF = 0,
    GTP5G_ROLE_RAN,
};

extern struct rtnl_link_ops gtp5g_link_ops;

extern void gtp5g_link_all_del(struct list_head *);

#endif // __GTP5G_LINK_H__