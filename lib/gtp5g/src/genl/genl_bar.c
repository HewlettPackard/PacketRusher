#include <net/genetlink.h>
#include <net/sock.h>
#include <linux/rculist.h>
#include <net/netns/generic.h>

#include "dev.h"
#include "genl.h"
#include "bar.h"
#include "genl_bar.h"
#include "net.h"

static int bar_fill(struct bar *, struct gtp5g_dev *, struct genl_info *);
static int gtp5g_genl_fill_bar(struct sk_buff *, u32, u32, u32, struct bar *);

int gtp5g_genl_add_bar(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct bar *bar;
    int ifindex;
    int netnsfd;
    u64 seid;
    u8 bar_id;
    int err;

    if (!info->attrs[GTP5G_LINK])
        return -EINVAL;
    ifindex = nla_get_u32(info->attrs[GTP5G_LINK]);

    if (info->attrs[GTP5G_NET_NS_FD])
        netnsfd = nla_get_u32(info->attrs[GTP5G_NET_NS_FD]);
    else
        netnsfd = -1;

    rtnl_lock();
    rcu_read_lock();

    gtp = gtp5g_find_dev(sock_net(skb->sk), ifindex, netnsfd);
    if (!gtp) {
        rcu_read_unlock();
        rtnl_unlock();
        return -ENODEV;
    }

    if (info->attrs[GTP5G_BAR_SEID]) {
        seid = nla_get_u64(info->attrs[GTP5G_BAR_SEID]);
    } else {
        seid = 0;
    }

    if (info->attrs[GTP5G_BAR_ID]) {
        bar_id = nla_get_u8(info->attrs[GTP5G_BAR_ID]);
    } else {
        rcu_read_unlock();
        rtnl_unlock();
        return -ENODEV;
    }

    bar = find_bar_by_id(gtp, seid, bar_id);
    if (bar) {
        if (info->nlhdr->nlmsg_flags & NLM_F_EXCL) {
            rcu_read_unlock();
            rtnl_unlock();
            return -EEXIST;
        }
        if (!(info->nlhdr->nlmsg_flags & NLM_F_REPLACE)) {
            rcu_read_unlock();
            rtnl_unlock();
            return -EOPNOTSUPP;
        }
        err = bar_fill(bar, gtp, info);
        if (err) {
            bar_context_delete(bar);
            return err;
        }
        return 0;
    }

    if (info->nlhdr->nlmsg_flags & NLM_F_REPLACE) {
        rcu_read_unlock();
        rtnl_unlock();
        return -ENOENT;
    }

    if (info->nlhdr->nlmsg_flags & NLM_F_APPEND) {
        rcu_read_unlock();
        rtnl_unlock();
        return -EOPNOTSUPP;
    }

    bar = kzalloc(sizeof(*bar), GFP_ATOMIC);
    if (!bar) {
        rcu_read_unlock();
        rtnl_unlock();
        return -ENOMEM;
    }

    bar->dev = gtp->dev;

    err = bar_fill(bar, gtp, info);
    if (err) {
        bar_context_delete(bar);
        rcu_read_unlock();
        rtnl_unlock();
        return err;
    }

    bar_append(seid, bar_id, bar, gtp);

    rcu_read_unlock();
    rtnl_unlock();
    return 0;
}

int gtp5g_genl_del_bar(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct bar *bar;
    int ifindex;
    int netnsfd;
    u64 seid;
    u8 bar_id;

    if (!info->attrs[GTP5G_LINK])
        return -EINVAL;
    ifindex = nla_get_u32(info->attrs[GTP5G_LINK]);

    if (info->attrs[GTP5G_NET_NS_FD])
        netnsfd = nla_get_u32(info->attrs[GTP5G_NET_NS_FD]);
    else
        netnsfd = -1;

    rcu_read_lock();

    gtp = gtp5g_find_dev(sock_net(skb->sk), ifindex, netnsfd);
    if (!gtp) {
        rcu_read_unlock();
        return -ENODEV;
    }

    if (info->attrs[GTP5G_BAR_SEID]) {
        seid = nla_get_u64(info->attrs[GTP5G_BAR_SEID]);
    } else {
        seid = 0;
    }

    if (info->attrs[GTP5G_BAR_ID]) {
        bar_id = nla_get_u8(info->attrs[GTP5G_BAR_ID]);
    } else {
        rcu_read_unlock();
        return -ENODEV;
    }

    bar = find_bar_by_id(gtp, seid, bar_id);
    if (!bar) {
        rcu_read_unlock();
        return -ENOENT;
    }

    bar_context_delete(bar);
    rcu_read_unlock();

    return 0;
}

int gtp5g_genl_get_bar(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct bar *bar;
    int ifindex;
    int netnsfd;
    u64 seid;
    u8 bar_id;
    struct sk_buff *skb_ack;
    int err;

    if (!info->attrs[GTP5G_LINK])
        return -EINVAL;
    ifindex = nla_get_u32(info->attrs[GTP5G_LINK]);

    if (info->attrs[GTP5G_NET_NS_FD])
        netnsfd = nla_get_u32(info->attrs[GTP5G_NET_NS_FD]);
    else
        netnsfd = -1;

    rcu_read_lock();

    gtp = gtp5g_find_dev(sock_net(skb->sk), ifindex, netnsfd);
    if (!gtp) {
        rcu_read_unlock();
        return -ENODEV;
    }

    if (info->attrs[GTP5G_BAR_SEID]) {
        seid = nla_get_u64(info->attrs[GTP5G_BAR_SEID]);
    } else {
        seid = 0;
    }

    if (info->attrs[GTP5G_BAR_ID]) {
        bar_id = nla_get_u8(info->attrs[GTP5G_BAR_ID]);
    } else {
        rcu_read_unlock();
        return -ENODEV;
    }

    bar = find_bar_by_id(gtp, seid, bar_id);
    if (!bar) {
        rcu_read_unlock();
        return -ENOENT;
    }

    skb_ack = genlmsg_new(NLMSG_GOODSIZE, GFP_ATOMIC);
    if (!skb_ack) {
        rcu_read_unlock();
        return -ENOMEM;
    }

    err = gtp5g_genl_fill_bar(skb_ack,
            NETLINK_CB(skb).portid,
            info->snd_seq,
            info->nlhdr->nlmsg_type,
            bar);
    if (err) {
        kfree_skb(skb_ack);
        rcu_read_unlock();
        return err;
    }
    rcu_read_unlock();

    return genlmsg_unicast(genl_info_net(info), skb_ack, info->snd_portid);
}

int gtp5g_genl_dump_bar(struct sk_buff *skb, struct netlink_callback *cb)
{
    /* netlink_callback->args
     * args[0] : index of gtp5g dev id
     * args[1] : index of gtp5g hash entry id in dev
     * args[2] : index of gtp5g bar id
     * args[5] : set non-zero means it is finished
     */
    struct gtp5g_dev *gtp;
    struct gtp5g_dev *last_gtp = (struct gtp5g_dev *)cb->args[0];
    struct net *net = sock_net(skb->sk);
    struct gtp5g_net *gn = net_generic(net, GTP5G_NET_ID());
    int i;
    int last_hash_entry_id = cb->args[1];
    int ret;
    u8 bar_id = cb->args[2];
    struct bar *bar;

    if (cb->args[5])
        return 0;

    list_for_each_entry_rcu(gtp, &gn->gtp5g_dev_list, list) {
        if (last_gtp && last_gtp != gtp)
            continue;
        else
            last_gtp = NULL;

        for (i = last_hash_entry_id; i < gtp->hash_size; i++) {
            hlist_for_each_entry_rcu(bar, &gtp->bar_id_hash[i], hlist_id) {
                if (bar_id && bar_id != bar->id)
                    continue;
                bar_id = 0;
                ret = gtp5g_genl_fill_bar(skb,
                        NETLINK_CB(cb->skb).portid,
                        cb->nlh->nlmsg_seq,
                        cb->nlh->nlmsg_type,
                        bar);
                if (ret) {
                    cb->args[0] = (unsigned long)gtp;
                    cb->args[1] = i;
                    cb->args[2] = bar->id;
                    goto out;
                }
            }
        }
    }

    cb->args[5] = 1;
out:
    return skb->len;
}


static int bar_fill(struct bar *bar, struct gtp5g_dev *gtp, struct genl_info *info)
{
    bar->id = nla_get_u8(info->attrs[GTP5G_BAR_ID]);

    if (info->attrs[GTP5G_BAR_SEID])
        bar->seid = nla_get_u64(info->attrs[GTP5G_BAR_SEID]);
    else
        bar->seid = 0;

    if (info->attrs[GTP5G_DOWNLINK_DATA_NOTIFICATION_DELAY])
        bar->delay = nla_get_u8(info->attrs[GTP5G_DOWNLINK_DATA_NOTIFICATION_DELAY]);
    else
        bar->delay = 0;

    if (info->attrs[GTP5G_BUFFERING_PACKETS_COUNT])
        bar->count = nla_get_u16(info->attrs[GTP5G_BUFFERING_PACKETS_COUNT]);
    else
        bar->count = 0;

    /* Update PDRs which has not linked to this BAR */
    bar_update(bar, gtp);
    return 0;
}

static int gtp5g_genl_fill_bar(struct sk_buff *skb, u32 snd_portid, u32 snd_seq,
        u32 type, struct bar *bar)
{
    void *genlh;

    genlh = genlmsg_put(skb, snd_portid, snd_seq,
            &gtp5g_genl_family, 0, type);
    if (!genlh)
        goto genlmsg_fail;

    if (nla_put_u8(skb, GTP5G_BAR_ID, bar->id))
        goto genlmsg_fail;
    if (nla_put_u8(skb, GTP5G_DOWNLINK_DATA_NOTIFICATION_DELAY, bar->delay))
        goto genlmsg_fail;
    if (nla_put_u16(skb, GTP5G_BUFFERING_PACKETS_COUNT, bar->count))
        goto genlmsg_fail;
    if (bar->seid) {
        if (nla_put_u64_64bit(skb, GTP5G_BAR_SEID, bar->seid, 0))
            goto genlmsg_fail;
    }

    genlmsg_end(skb, genlh);
    return 0;
genlmsg_fail:
    genlmsg_cancel(skb, genlh);
    return -EMSGSIZE;
}