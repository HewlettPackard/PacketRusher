#ifndef __GTP5G_NET_H__
#define __GTP5G_NET_H__

#include <linux/rculist.h>
#include <net/net_namespace.h>

struct gtp5g_net {
    struct list_head gtp5g_dev_list;
};

extern struct pernet_operations gtp5g_net_ops;

#define GTP5G_NET_ID() (*gtp5g_net_ops.id)

#endif // __GTP5G_NET_H__