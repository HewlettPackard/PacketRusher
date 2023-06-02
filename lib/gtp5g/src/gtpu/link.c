#include <net/rtnetlink.h>
#include <net/ip.h>
#include <net/udp.h>
#include <net/netns/generic.h>

#include "dev.h"
#include "link.h"
#include "net.h"
#include "encap.h"
#include "gtp.h"
#include "log.h"
#include "proc.h"

const struct nla_policy gtp5g_policy[IFLA_GTP5G_MAX + 1] = {
    [IFLA_GTP5G_FD1]             = { .type = NLA_U32 },
    [IFLA_GTP5G_PDR_HASHSIZE]    = { .type = NLA_U32 },
    [IFLA_GTP5G_ROLE]            = { .type = NLA_U32 },
};

static void gtp5g_link_setup(struct net_device *dev)
{
    dev->netdev_ops = &gtp5g_netdev_ops;
    dev->needs_free_netdev = true;

    dev->hard_header_len = 0;
    dev->addr_len = 0;
    dev->mtu = ETH_DATA_LEN -
        (sizeof(struct iphdr) +
         sizeof(struct udphdr) +
         sizeof(struct gtpv1_hdr));

    /* Zero header length. */
    dev->type = ARPHRD_NONE;
    dev->flags = IFF_POINTOPOINT | IFF_NOARP | IFF_MULTICAST;

    dev->priv_flags |= IFF_NO_QUEUE;
    dev->features |= NETIF_F_LLTX;
    netif_keep_dst(dev);

    /* TODO: Modify the headroom size based on
     * what are the extension header going to support
     * */
    dev->needed_headroom = LL_MAX_HEADER +
        sizeof(struct iphdr) +
        sizeof(struct udphdr) +
        sizeof(struct gtpv1_hdr);
}

static int gtp5g_validate(struct nlattr *tb[], struct nlattr *data[],
    struct netlink_ext_ack *extack)
{
    if (!data)
        return -EINVAL;

    return 0;
}

static int gtp5g_newlink(struct net *src_net, struct net_device *dev,
    struct nlattr *tb[], struct nlattr *data[],
    struct netlink_ext_ack *extack)
{
    struct gtp5g_dev *gtp;
    struct gtp5g_net *gn;
    struct sock *sk;
    unsigned int role = GTP5G_ROLE_UPF;
    u32 fd1;
    int hashsize, err;

    gtp = netdev_priv(dev);

    if (!data[IFLA_GTP5G_FD1]) {
        GTP5G_ERR(NULL, "Failed to create a new link\n");
        return -EINVAL;
    }
    fd1 = nla_get_u32(data[IFLA_GTP5G_FD1]);
    sk = gtp5g_encap_enable(fd1, UDP_ENCAP_GTP1U, gtp);
    if (IS_ERR(sk))
        return PTR_ERR(sk);
    gtp->sk1u = sk;
    
    if (data[IFLA_GTP5G_ROLE]) {
        role = nla_get_u32(data[IFLA_GTP5G_ROLE]);
        if (role > GTP5G_ROLE_RAN) {
            if (sk)
                gtp5g_encap_disable(sk);
            return -EINVAL;
        }
    }
    gtp->role = role;
    
    if (!data[IFLA_GTP5G_PDR_HASHSIZE])
        hashsize = 1024;
    else
        hashsize = nla_get_u32(data[IFLA_GTP5G_PDR_HASHSIZE]);

    err = dev_hashtable_new(gtp, hashsize);
    if (err < 0) {
        gtp5g_encap_disable(gtp->sk1u);
        GTP5G_ERR(dev, "Failed to create a hash table\n");
        goto out_encap;
    }

    err = register_netdevice(dev);
    if (err < 0) {
        netdev_dbg(dev, "failed to register new netdev %d\n", err);
        gtp5g_hashtable_free(gtp);
        gtp5g_encap_disable(gtp->sk1u);
        goto out_hashtable;
    }

    gn = net_generic(dev_net(dev), GTP5G_NET_ID());
    list_add_rcu(&gtp->list, &gn->gtp5g_dev_list);
    list_add_rcu(&gtp->proc_list, get_proc_gtp5g_dev_list_head());

    GTP5G_LOG(dev, "Registered a new 5G GTP interface\n");
    return 0;
out_hashtable:
    gtp5g_hashtable_free(gtp);
out_encap:
    gtp5g_encap_disable(gtp->sk1u);
    return err;
}

static void gtp5g_dellink(struct net_device *dev, struct list_head *head)
{
    struct gtp5g_dev *gtp = netdev_priv(dev);

    gtp5g_hashtable_free(gtp);
    list_del_rcu(&gtp->list);
    list_del_rcu(&gtp->proc_list);
    unregister_netdevice_queue(dev, head);

    GTP5G_LOG(dev, "De-registered 5G GTP interface\n");
}

static size_t gtp5g_get_size(const struct net_device *dev)
{
    /* IFLA_UPF_PDR_HASHSIZE */
    return nla_total_size(sizeof(__u32));
}

static int gtp5g_fill_info(struct sk_buff *skb, const struct net_device *dev)
{
    struct gtp5g_dev *gtp = netdev_priv(dev);

    if (nla_put_u32(skb, IFLA_GTP5G_PDR_HASHSIZE, gtp->hash_size))
        goto nla_put_failure;

    return 0;
nla_put_failure:
    return -EMSGSIZE;
}

struct rtnl_link_ops gtp5g_link_ops __read_mostly = {
    .kind         = "gtp5g",
    .maxtype      = IFLA_GTP5G_MAX,
    .policy       = gtp5g_policy,
    .priv_size    = sizeof(struct gtp5g_dev),
    .setup        = gtp5g_link_setup,
    .validate     = gtp5g_validate,
    .newlink      = gtp5g_newlink,
    .dellink      = gtp5g_dellink,
    .get_size     = gtp5g_get_size,
    .fill_info    = gtp5g_fill_info,
};

void gtp5g_link_all_del(struct list_head *dev_list)
{
    struct gtp5g_dev *gtp;
    LIST_HEAD(list);

    rtnl_lock();
    list_for_each_entry(gtp, dev_list, list)
        gtp5g_dellink(gtp->dev, &list);

    unregister_netdevice_many(&list);
    rtnl_unlock();
}