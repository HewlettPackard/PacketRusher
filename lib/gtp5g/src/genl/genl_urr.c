#include <net/genetlink.h>
#include <net/sock.h>
#include <linux/rculist.h>
#include <net/netns/generic.h>

#include "common.h"
#include "dev.h"
#include "genl.h"
#include "urr.h"
#include "genl_urr.h"
#include "genl_report.h"

#include "net.h"
#include "log.h"

static int urr_fill(struct urr *, struct gtp5g_dev *, struct genl_info *);
static int gtp5g_genl_fill_urr(struct sk_buff *, u32, u32, u32, struct urr *);
static int parse_volumethreshold(struct urr *urr, struct nlattr *a);
static int parse_volumeqouta(struct urr *urr, struct nlattr *a);

int gtp5g_genl_add_urr(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct urr *urr;
    int ifindex;
    int netnsfd;
    u64 seid;
    u32 urr_id;
    int err = 0;

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
        err = -ENODEV;
        goto end;
    }

    if (info->attrs[GTP5G_URR_SEID]) {
        seid = nla_get_u64(info->attrs[GTP5G_URR_SEID]);
    } else {
        seid = 0;
    }

    if (!info->attrs[GTP5G_URR_ID]) {
        err = -ENODEV;
        goto end;
    }
    urr_id = nla_get_u32(info->attrs[GTP5G_URR_ID]);

    urr = find_urr_by_id(gtp, seid, urr_id);
    if (urr) {
        if (info->nlhdr->nlmsg_flags & NLM_F_EXCL) {
            err = -EEXIST;
            goto end;
        }
        if (!(info->nlhdr->nlmsg_flags & NLM_F_REPLACE)) {
            err = -EOPNOTSUPP;
            goto end;
        }
        err = urr_fill(urr, gtp, info);
        if (err) {
            urr_context_delete(urr);
            goto end;
        }

        goto end;
    }

    if (info->nlhdr->nlmsg_flags & NLM_F_REPLACE) {
        err = -ENOENT;
        goto end;
    }

    if (info->nlhdr->nlmsg_flags & NLM_F_APPEND) {
        err = -EOPNOTSUPP;
        goto end;
    }

    urr = kzalloc(sizeof(*urr), GFP_ATOMIC);
    if (!urr) {
        err = -ENOMEM;
        goto end;
    }

    urr->dev = gtp->dev;
    urr->start_time = ktime_get_real();

    err = urr_fill(urr, gtp, info);
    if (err) {
        urr_context_delete(urr);
        goto end;
    }

    urr_append(seid, urr_id, urr, gtp);

end:
    rcu_read_unlock();
    rtnl_unlock();
    return err;
}

int gtp5g_genl_del_urr(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct urr *urr = NULL;
    int ifindex;
    int netnsfd;
    u64 seid;
    u32 urr_id;
    struct sk_buff *skb_ack = NULL;
    int err = 0;
    struct usage_report *report = NULL;

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
        err = -ENODEV;
        goto fail;
    }

    if (info->attrs[GTP5G_URR_SEID]) {
        seid = nla_get_u64(info->attrs[GTP5G_URR_SEID]);
    } else {
        seid = 0;
    }

    if (!info->attrs[GTP5G_URR_ID]) {
        err = -ENODEV;
        goto fail;
    }
    urr_id = nla_get_u32(info->attrs[GTP5G_URR_ID]);

    urr = find_urr_by_id(gtp, seid, urr_id);
    if (!urr) {
        err = -ENOENT;
        goto fail;
    }

    skb_ack = genlmsg_new(NLMSG_GOODSIZE, GFP_ATOMIC);
    if (!skb_ack) {
        err = -ENOMEM;
        goto fail;
    }

    report = kzalloc(sizeof(struct usage_report), GFP_KERNEL);
    if (!report) {
        err = -ENOMEM;
        goto fail;
    }

    convert_urr_to_report(urr, report);

    err = gtp5g_genl_fill_usage_report(skb_ack,
            NETLINK_CB(skb).portid,
            info->snd_seq,
            info->nlhdr->nlmsg_type,
            report);

    kfree(report);

fail:
    if (err) {
        if (skb_ack) {
            kfree_skb(skb_ack);
        }
        rcu_read_unlock();
        return err;
    }

    urr_context_delete(urr);
    rcu_read_unlock();
    return genlmsg_unicast(genl_info_net(info), skb_ack, info->snd_portid);
}

int gtp5g_genl_get_urr(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct urr *urr;
    int ifindex;
    int netnsfd;
    u64 seid;
    u32 urr_id;
    struct sk_buff *skb_ack = NULL;
    int err = 0;

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
        err = -ENODEV;
        goto fail;
    }

    if (info->attrs[GTP5G_URR_SEID]) {
        seid = nla_get_u64(info->attrs[GTP5G_URR_SEID]);
    } else {
        seid = 0;
    }

    if (!info->attrs[GTP5G_URR_ID]) {
        err = -ENODEV;
        goto fail;
    }
    urr_id = nla_get_u32(info->attrs[GTP5G_URR_ID]);

    urr = find_urr_by_id(gtp, seid, urr_id);
    if (!urr) {
        err = -ENOENT;
        goto fail;
    }

    skb_ack = genlmsg_new(NLMSG_GOODSIZE, GFP_ATOMIC);
    if (!skb_ack) {
        err = -ENOMEM;
        goto fail;
    }

    err = gtp5g_genl_fill_urr(skb_ack,
            NETLINK_CB(skb).portid,
            info->snd_seq,
            info->nlhdr->nlmsg_type,
            urr);

fail:
    if (err) {
        if (skb_ack) {
            kfree_skb(skb_ack);
        }
        rcu_read_unlock();
        return err;
    }

    rcu_read_unlock();
    return genlmsg_unicast(genl_info_net(info), skb_ack, info->snd_portid); 
}

int gtp5g_genl_dump_urr(struct sk_buff *skb, struct netlink_callback *cb)
{
    /* netlink_callback->args
     * args[0] : index of gtp5g dev id
     * args[1] : index of gtp5g hash entry id in dev
     * args[2] : index of gtp5g urr id
     * args[5] : set non-zero means it is finished
     */
    struct gtp5g_dev *gtp;
    struct gtp5g_dev *last_gtp = (struct gtp5g_dev *)cb->args[0];
    struct net *net = sock_net(skb->sk);
    struct gtp5g_net *gn = net_generic(net, GTP5G_NET_ID());
    int i;
    int last_hash_entry_id = cb->args[1];
    int ret;
    u32 urr_id = cb->args[2];
    struct urr *urr;

    if (cb->args[5])
        return 0;

    list_for_each_entry_rcu(gtp, &gn->gtp5g_dev_list, list) {
        if (last_gtp && last_gtp != gtp)
            continue;
        else
            last_gtp = NULL;

        for (i = last_hash_entry_id; i < gtp->hash_size; i++) {
            hlist_for_each_entry_rcu(urr, &gtp->urr_id_hash[i], hlist_id) {
                if (urr_id && urr_id != urr->id)
                    continue;
                urr_id = 0;
                ret = gtp5g_genl_fill_urr(skb,
                        NETLINK_CB(cb->skb).portid,
                        cb->nlh->nlmsg_seq,
                        cb->nlh->nlmsg_type,
                        urr);
                if (ret) {
                    cb->args[0] = (unsigned long)gtp;
                    cb->args[1] = i;
                    cb->args[2] = urr->id;
                    goto out;
                }
            }
        }
    }

    cb->args[5] = 1;
out:
    return skb->len;
}


static int urr_fill(struct urr *urr, struct gtp5g_dev *gtp, struct genl_info *info)
{
    urr->id = nla_get_u32(info->attrs[GTP5G_URR_ID]);

    if (info->attrs[GTP5G_URR_SEID])
        urr->seid = nla_get_u64(info->attrs[GTP5G_URR_SEID]);
    else
        urr->seid = 0;

    if (info->attrs[GTP5G_URR_MEASUREMENT_METHOD])
        urr->method = nla_get_u8(info->attrs[GTP5G_URR_MEASUREMENT_METHOD]);

    if (info->attrs[GTP5G_URR_REPORTING_TRIGGER]) {
        urr->trigger = nla_get_u32(info->attrs[GTP5G_URR_REPORTING_TRIGGER]);
        if (urr->trigger == URR_RPT_TRIGGER_START) {
            // Clean bytes to make sure the bytes are counted after the start of service data flow
            // TODO: Should send the previous stroed bytes to CP first
            memset(&urr->bytes, 0, sizeof(struct VolumeMeasurement));
            memset(&urr->bytes2, 0, sizeof(struct VolumeMeasurement));
        }
    }

    if (info->attrs[GTP5G_URR_MEASUREMENT_PERIOD])
        urr->period = nla_get_u32(info->attrs[GTP5G_URR_MEASUREMENT_PERIOD]);

    if (info->attrs[GTP5G_URR_MEASUREMENT_INFO])
        urr->info = nla_get_u8(info->attrs[GTP5G_URR_MEASUREMENT_INFO]);

    if (info->attrs[GTP5G_URR_VOLUME_THRESHOLD])
        parse_volumethreshold(urr, info->attrs[GTP5G_URR_VOLUME_THRESHOLD]);

    if (info->attrs[GTP5G_URR_VOLUME_QUOTA]) {
        parse_volumeqouta(urr, info->attrs[GTP5G_URR_VOLUME_QUOTA]);
        memset(&urr->consumed, 0, sizeof(struct VolumeMeasurement));

        if (urr->volumequota.totalVolume == 0 && urr->trigger == URR_RPT_TRIGGER_VOLQU) {
            urr_quota_exhaust_action(urr, gtp);
            GTP5G_INF(NULL, "URR (%u) Receive zero quota, stop measure", urr->id);
        } else if (urr->quota_exhausted) {
            urr_reverse_quota_exhaust_action(urr, gtp);
            GTP5G_INF(NULL, "URR (%u) Receive New quota, continue measure", urr->id);
        }
    }

    /* Update PDRs which has not linked to this URR */
    urr_update(urr, gtp);

    return 0;
}

static int parse_volumethreshold(struct urr *urr, struct nlattr *a)
{
    struct nlattr *attrs[GTP5G_URR_ATTR_MAX + 1];
    int err;

    err = nla_parse_nested(attrs, GTP5G_URR_VOLUME_THRESHOLD_ATTR_MAX, a, NULL, NULL);
    if (err)
        return err;

    if (attrs[GTP5G_URR_VOLUME_THRESHOLD_FLAG]) {
        urr->volumethreshold.flag = nla_get_u8(attrs[GTP5G_URR_VOLUME_THRESHOLD_FLAG]);
    }

    if (attrs[GTP5G_URR_VOLUME_THRESHOLD_TOVOL]) {
        urr->volumethreshold.totalVolume = nla_get_u64(attrs[GTP5G_URR_VOLUME_THRESHOLD_TOVOL]);
    }

    if (attrs[GTP5G_URR_VOLUME_THRESHOLD_UVOL]) {
        urr->volumethreshold.uplinkVolume = nla_get_u64(attrs[GTP5G_URR_VOLUME_THRESHOLD_UVOL]);
    }

    if (attrs[GTP5G_URR_VOLUME_THRESHOLD_DVOL]) {
        urr->volumethreshold.downlinkVolume = nla_get_u64(attrs[GTP5G_URR_VOLUME_THRESHOLD_DVOL]);
    }

    return 0;
}

static int parse_volumeqouta(struct urr *urr, struct nlattr *a)
{
    struct nlattr *attrs[GTP5G_URR_ATTR_MAX + 1];
    int err;

    err = nla_parse_nested(attrs, GTP5G_URR_VOLUME_QUOTA_ATTR_MAX, a, NULL, NULL);
    if (err)
        return err;

    if (attrs[GTP5G_URR_VOLUME_QUOTA_FLAG]) {
        urr->volumequota.flag = nla_get_u8(attrs[GTP5G_URR_VOLUME_QUOTA_FLAG]);
    }

    if (attrs[GTP5G_URR_VOLUME_QUOTA_TOVOL]) {
        urr->volumequota.totalVolume = nla_get_u64(attrs[GTP5G_URR_VOLUME_QUOTA_TOVOL]);
    }

    if (attrs[GTP5G_URR_VOLUME_QUOTA_UVOL]) {
        urr->volumequota.uplinkVolume = nla_get_u64(attrs[GTP5G_URR_VOLUME_QUOTA_UVOL]);
    }

    if (attrs[GTP5G_URR_VOLUME_QUOTA_DVOL]) {
        urr->volumequota.downlinkVolume = nla_get_u64(attrs[GTP5G_URR_VOLUME_QUOTA_DVOL]);
    }

    return 0;
}

static int gtp5g_genl_fill_volume_threshold(struct sk_buff *skb, struct Volume volumethreshold)
{
    struct nlattr *nest_volume_threshold;

    nest_volume_threshold = nla_nest_start(skb, GTP5G_URR_VOLUME_THRESHOLD);
    if (!nest_volume_threshold)
        return -EMSGSIZE;

    if (nla_put_u8(skb, GTP5G_URR_VOLUME_THRESHOLD_FLAG, volumethreshold.flag))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_URR_VOLUME_THRESHOLD_TOVOL, volumethreshold.totalVolume, 0))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_URR_VOLUME_THRESHOLD_UVOL, volumethreshold.uplinkVolume, 0))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_URR_VOLUME_THRESHOLD_DVOL, volumethreshold.downlinkVolume, 0))
        return -EMSGSIZE;

    nla_nest_end(skb, nest_volume_threshold);
    return 0;
}

static int gtp5g_genl_fill_volume_quota(struct sk_buff *skb, struct Volume volumequota)
{
    struct nlattr *nest_volume_quota;

    nest_volume_quota = nla_nest_start(skb, GTP5G_URR_VOLUME_QUOTA);
    if (!nest_volume_quota)
        return -EMSGSIZE;

    if (nla_put_u8(skb, GTP5G_URR_VOLUME_QUOTA_FLAG, volumequota.flag))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_URR_VOLUME_QUOTA_TOVOL, volumequota.totalVolume, 0))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_URR_VOLUME_QUOTA_UVOL, volumequota.uplinkVolume, 0))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_URR_VOLUME_QUOTA_DVOL, volumequota.downlinkVolume, 0))
        return -EMSGSIZE;

    nla_nest_end(skb, nest_volume_quota);
    return 0;
}

static int gtp5g_genl_fill_urr(struct sk_buff *skb, u32 snd_portid, u32 snd_seq,
        u32 type, struct urr *urr)
{
    struct gtp5g_dev *gtp = netdev_priv(urr->dev);
    void *genlh;
    u16 *ids = NULL;
    int n;

    genlh = genlmsg_put(skb, snd_portid, snd_seq,
            &gtp5g_genl_family, 0, type);
    if (!genlh)
        goto genlmsg_fail;

    if (nla_put_u32(skb, GTP5G_URR_ID, urr->id))
        goto genlmsg_fail;
    if (nla_put_u8(skb, GTP5G_URR_MEASUREMENT_METHOD, urr->method))
        goto genlmsg_fail;
    if (nla_put_u32(skb, GTP5G_URR_REPORTING_TRIGGER, urr->trigger))
        goto genlmsg_fail;
    if (nla_put_u32(skb, GTP5G_URR_MEASUREMENT_PERIOD, urr->period))
        goto genlmsg_fail;
    if (nla_put_u8(skb, GTP5G_URR_MEASUREMENT_INFO, urr->info))
        goto genlmsg_fail;
    if (urr->seid) {
        if (nla_put_u64_64bit(skb, GTP5G_URR_SEID, urr->seid, 0))
            goto genlmsg_fail;
    }
    if (urr->volumethreshold.flag != 0) {
        if (gtp5g_genl_fill_volume_threshold(skb, urr->volumethreshold))
            goto genlmsg_fail;
    }

    if (urr->volumequota.flag != 0) {
        if (gtp5g_genl_fill_volume_quota(skb, urr->volumequota))
            goto genlmsg_fail;
    }
    ids = kzalloc(MAX_PDR_PER_SESSION * sizeof(u16), GFP_KERNEL);
    if (!ids)
        goto genlmsg_fail;
    n = urr_get_pdr_ids(ids, MAX_PDR_PER_SESSION, urr, gtp);
    if (n) {
        if (nla_put(skb, GTP5G_URR_RELATED_TO_PDR, n * sizeof(u16), ids))
            goto genlmsg_fail;
    }

    kfree(ids);
    genlmsg_end(skb, genlh);
    return 0;

genlmsg_fail:
    if (ids) {
        kfree(ids);
    }
    genlmsg_cancel(skb, genlh);
    return -EMSGSIZE;
}
