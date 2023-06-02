#include <linux/netdevice.h>
#include <net/udp_tunnel.h>
#include <linux/version.h>

#include "dev.h"
#include "genl.h"
#include "encap.h"
#include "pdr.h"
#include "far.h"
#include "qer.h"
#include "bar.h"
#include "urr.h"
#include "pktinfo.h"

struct gtp5g_dev *gtp5g_find_dev(struct net *src_net, int ifindex, int netnsfd)
{
    struct gtp5g_dev *gtp = NULL;
    struct net_device *dev;
    struct net *net;

    /* Examine the link attributes and figure out which network namespace
     * we are talking about.
     */
    if (netnsfd == -1)
        net = get_net(src_net);
    else
        net = get_net_ns_by_fd(netnsfd);

    if (IS_ERR(net))
        return NULL;

    /* Check if there's an existing gtp5g device to configure */
    dev = dev_get_by_index_rcu(net, ifindex);
    if (dev && dev->netdev_ops == &gtp5g_netdev_ops)
        gtp = netdev_priv(dev);
    else
        gtp = NULL;

    put_net(net);

    return gtp;
}

static int gtp5g_dev_init(struct net_device *dev)
{
    struct gtp5g_dev *gtp = netdev_priv(dev);

    gtp->dev = dev;

    dev->tstats = netdev_alloc_pcpu_stats(struct pcpu_sw_netstats);
    if (!dev->tstats) {
        return -ENOMEM;
    }

    return 0;
}

static void gtp5g_dev_uninit(struct net_device *dev)
{
    struct gtp5g_dev *gtp = netdev_priv(dev);

    gtp5g_encap_disable(gtp->sk1u);
    free_percpu(dev->tstats);
}

/**
 * Entry function for Downlink packets
 * */
static netdev_tx_t gtp5g_dev_xmit(struct sk_buff *skb, struct net_device *dev)
{
    unsigned int proto = ntohs(skb->protocol);
    struct gtp5g_pktinfo pktinfo;
    int ret = 0;

    /* Ensure there is sufficient headroom */
    if (skb_cow_head(skb, dev->needed_headroom)) {
        goto tx_err;
    }

    skb_reset_inner_headers(skb);

    /* PDR lookups in gtp5g_build_skb_*() need rcu read-side lock. 
     * */
    rcu_read_lock();
    switch (proto) {
    case ETH_P_IP:
        ret = gtp5g_handle_skb_ipv4(skb, dev, &pktinfo);
        break;
    default:
        ret = -EOPNOTSUPP;
    }
    rcu_read_unlock();

    if (ret < 0)
        goto tx_err;

    if (ret == FAR_ACTION_FORW)
        gtp5g_xmit_skb_ipv4(skb, &pktinfo);

    return NETDEV_TX_OK;

tx_err:
    dev->stats.tx_errors++;
    dev_kfree_skb(skb);
    return NETDEV_TX_OK;
}

const struct net_device_ops gtp5g_netdev_ops = {
    .ndo_init           = gtp5g_dev_init,
    .ndo_uninit         = gtp5g_dev_uninit,
    .ndo_start_xmit     = gtp5g_dev_xmit,
#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 11, 0)
    .ndo_get_stats64    = dev_get_tstats64,
#else
    .ndo_get_stats64    = ip_tunnel_get_stats64,
#endif
};

int dev_hashtable_new(struct gtp5g_dev *gtp, int hsize)
{
    int i;

    gtp->addr_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
        GFP_KERNEL);
    if (gtp->addr_hash == NULL)
        return -ENOMEM;

    gtp->i_teid_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
        GFP_KERNEL);
    if (gtp->i_teid_hash == NULL)
        goto err1;

    gtp->pdr_id_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
        GFP_KERNEL);
    if (gtp->pdr_id_hash == NULL)
        goto err2;

    gtp->far_id_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
        GFP_KERNEL);
    if (gtp->far_id_hash == NULL)
        goto err3;

    gtp->qer_id_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
        GFP_KERNEL);
    if (gtp->qer_id_hash == NULL)
        goto err4;

    gtp->bar_id_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
            GFP_KERNEL);
    if (!gtp->bar_id_hash)
        goto err5;

    gtp->urr_id_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
            GFP_KERNEL);
    if (!gtp->urr_id_hash)
        goto err6;

    gtp->related_far_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
        GFP_KERNEL);
    if (gtp->related_far_hash == NULL)
        goto err7;

    gtp->related_qer_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
        GFP_KERNEL);
    if (gtp->related_qer_hash == NULL)
        goto err8;

    gtp->related_bar_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
            GFP_KERNEL);
    if (!gtp->related_bar_hash)
        goto err9;

    gtp->related_urr_hash = kmalloc_array(hsize, sizeof(struct hlist_head),
            GFP_KERNEL);
    if (!gtp->related_urr_hash)
        goto err10;

    gtp->hash_size = hsize;

    for (i = 0; i < hsize; i++) {
        INIT_HLIST_HEAD(&gtp->addr_hash[i]);
        INIT_HLIST_HEAD(&gtp->i_teid_hash[i]);
        INIT_HLIST_HEAD(&gtp->pdr_id_hash[i]);
        INIT_HLIST_HEAD(&gtp->far_id_hash[i]);
        INIT_HLIST_HEAD(&gtp->qer_id_hash[i]);
        INIT_HLIST_HEAD(&gtp->bar_id_hash[i]);
        INIT_HLIST_HEAD(&gtp->urr_id_hash[i]);
        INIT_HLIST_HEAD(&gtp->related_far_hash[i]);
        INIT_HLIST_HEAD(&gtp->related_qer_hash[i]);
        INIT_HLIST_HEAD(&gtp->related_bar_hash[i]);
        INIT_HLIST_HEAD(&gtp->related_urr_hash[i]);
    }

    return 0;
err10:
    kfree(gtp->related_bar_hash);
err9:
    kfree(gtp->related_qer_hash);
err8:
    kfree(gtp->related_far_hash);
err7:
    kfree(gtp->urr_id_hash);
err6:
    kfree(gtp->bar_id_hash);
err5:
    kfree(gtp->qer_id_hash);
err4:
    kfree(gtp->far_id_hash);
err3:
    kfree(gtp->pdr_id_hash);
err2:
    kfree(gtp->i_teid_hash);
err1:
    kfree(gtp->addr_hash);
    return -ENOMEM;
}

void gtp5g_hashtable_free(struct gtp5g_dev *gtp)
{
    struct pdr *pdr;
    struct far *far;
    struct qer *qer;
    struct bar *bar;
    struct urr *urr;
    int i;

    for (i = 0; i < gtp->hash_size; i++) {
        hlist_for_each_entry_rcu(far, &gtp->far_id_hash[i], hlist_id)
            far_context_delete(far);
        hlist_for_each_entry_rcu(qer, &gtp->qer_id_hash[i], hlist_id)
            qer_context_delete(qer);
        hlist_for_each_entry_rcu(pdr, &gtp->pdr_id_hash[i], hlist_id)
            pdr_context_delete(pdr);
        hlist_for_each_entry_rcu(bar, &gtp->bar_id_hash[i], hlist_id)
            bar_context_delete(bar);
        hlist_for_each_entry_rcu(urr, &gtp->urr_id_hash[i], hlist_id)
            urr_context_delete(urr);
    }

    synchronize_rcu();
    kfree(gtp->addr_hash);
    kfree(gtp->i_teid_hash);
    kfree(gtp->pdr_id_hash);
    kfree(gtp->far_id_hash);
    kfree(gtp->qer_id_hash);
    kfree(gtp->bar_id_hash);
    kfree(gtp->urr_id_hash);
    kfree(gtp->related_far_hash);
    kfree(gtp->related_qer_hash);
    kfree(gtp->related_bar_hash);
    kfree(gtp->related_urr_hash);
}