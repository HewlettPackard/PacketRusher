#include <linux/rculist.h>
#include <net/net_namespace.h>
#include <net/netns/generic.h>

#include "dev.h"
#include "net.h"
#include "link.h"

static unsigned int gtp5g_net_id __read_mostly;

static int __net_init gtp5g_net_init(struct net *net)
{
    struct gtp5g_net *gn = net_generic(net, gtp5g_net_id);

    INIT_LIST_HEAD(&gn->gtp5g_dev_list);
    return 0;
}

static void __net_exit gtp5g_net_exit(struct net *net)
{
    struct gtp5g_net *gn = net_generic(net, gtp5g_net_id);

    gtp5g_link_all_del(&gn->gtp5g_dev_list);
}

struct pernet_operations gtp5g_net_ops = {
    .init    = gtp5g_net_init,
    .exit    = gtp5g_net_exit,
    .id      = &gtp5g_net_id,
    .size    = sizeof(struct gtp5g_net),
};
