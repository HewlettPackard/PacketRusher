#include <net/genetlink.h>
#include <net/sock.h>
#include <linux/rculist.h>
#include <net/netns/generic.h>

#include "dev.h"
#include "genl.h"
#include "report.h"
#include "genl_report.h"
#include "genl_urr.h"
#include "hash.h"

#include "net.h"
#include "log.h"
#include "pdr.h"
#include "urr.h"

static int gtp5g_genl_fill_volume_measurement(struct sk_buff *, struct VolumeMeasurement);
static int parse_seid_urr(struct seid_urr *, struct nlattr *);

int gtp5g_genl_get_usage_report(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct urr *urr;
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

    rcu_read_unlock();
    return genlmsg_unicast(genl_info_net(info), skb_ack, info->snd_portid); 
}

int gtp5g_genl_get_multi_usage_reports(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct urr *urr = NULL;
    int ifindex;
    int netnsfd;
    u32 urr_num, report_num = 0, i = 0;
    struct sk_buff *skb_ack = NULL;
    int err = 0;
    struct nlattr *hdr = nlmsg_attrdata(info->nlhdr, 0);
    int remaining = nlmsg_attrlen(info->nlhdr, 0);
    struct seid_urr **seid_urrs;
    struct usage_report **reports = NULL;

    if (!info->attrs[GTP5G_LINK])
        return -EINVAL;
    ifindex = nla_get_u32(info->attrs[GTP5G_LINK]);

    if (info->attrs[GTP5G_NET_NS_FD])
        netnsfd = nla_get_u32(info->attrs[GTP5G_NET_NS_FD]);
    else
        netnsfd = -1;

    if (!info->attrs[GTP5G_URR_NUM])
        return -EINVAL;
    urr_num = nla_get_u32(info->attrs[GTP5G_URR_NUM]);

    rcu_read_lock();

    seid_urrs = kzalloc(sizeof(struct seid_urr *) * urr_num , GFP_KERNEL);
    hdr = nla_next(hdr, &remaining);
    while (nla_ok(hdr,remaining)) {
        switch (nla_type(hdr)) {
        case GTP5G_URR_MULTI_SEID_URRID:
            seid_urrs[i] = kzalloc(sizeof(struct seid_urr), GFP_KERNEL);
            err = parse_seid_urr(seid_urrs[i++], hdr);
            if (err) {
                goto fail;
            }
            break;
        }
        hdr = nla_next(hdr, &remaining);
    }

    gtp = gtp5g_find_dev(sock_net(skb->sk), ifindex, netnsfd);
    if (!gtp) {
        err = -ENODEV;
        goto fail;
    }

    reports = kzalloc(sizeof(struct usage_report *) * urr_num , GFP_KERNEL);
    if (!reports) {
        err =  -ENOMEM;
        goto fail;
    }

    for (i = 0; i < urr_num; i++) {
        urr = find_urr_by_id(gtp, seid_urrs[i]->seid, seid_urrs[i]->urrid);
        if (!urr) {
            err =  -ENOENT;
            goto fail;
        }

        reports[i] = kzalloc(sizeof(struct usage_report), GFP_KERNEL);
        if (!reports[i]) {
            err =  -ENOMEM;
            goto fail;
        }

        convert_urr_to_report(urr, reports[report_num++]);
    }

    skb_ack = genlmsg_new(NLMSG_GOODSIZE, GFP_ATOMIC);
    if (!skb_ack) {
        err =  -ENOMEM;
        goto fail;
    }

    err = gtp5g_genl_fill_multi_usage_reports(skb_ack,
            NETLINK_CB(skb).portid,
            info->snd_seq,
            info->nlhdr->nlmsg_type,
            reports, report_num);

fail:
    if (reports) {
        for (i = 0; i < report_num; i++) {
            if (reports[i])
                kfree(reports[i]);
        }
        kfree(reports);
    }
    if (seid_urrs) {
        for (i = 0; i < urr_num; i++) {
            if (seid_urrs[i])
                kfree(seid_urrs[i]);
        }
        kfree(seid_urrs);
    }

    if (err) {
        if (skb_ack)
            kfree_skb(skb_ack);
        rcu_read_unlock();
        return err;
    }

    rcu_read_unlock();
    return genlmsg_unicast(genl_info_net(info), skb_ack, info->snd_portid);
}

static int gtp5g_genl_fill_volume_measurement(struct sk_buff *skb, struct VolumeMeasurement bytes)
{
    struct nlattr *nest_volume_measurement;

    nest_volume_measurement = nla_nest_start(skb, GTP5G_UR_VOLUME_MEASUREMENT);
    if (!nest_volume_measurement)
        return -EMSGSIZE;

    if (nla_put_u64_64bit(skb, GTP5G_UR_VOLUME_MEASUREMENT_TOVOL, bytes.totalVolume , 0))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_UR_VOLUME_MEASUREMENT_UVOL, bytes.uplinkVolume, 0))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_UR_VOLUME_MEASUREMENT_DVOL, bytes.downlinkVolume, 0))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_UR_VOLUME_MEASUREMENT_TOPACKET, bytes.totalPktNum, 0))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_UR_VOLUME_MEASUREMENT_UPACKET, bytes.uplinkPktNum, 0))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_UR_VOLUME_MEASUREMENT_DPACKET, bytes.downlinkPktNum, 0))
        return -EMSGSIZE;

    nla_nest_end(skb, nest_volume_measurement);

    memset(&bytes, 0, sizeof(struct VolumeMeasurement));

    return 0;
}

int gtp5g_genl_fill_ur(struct sk_buff *skb, struct usage_report *report)
{
    struct nlattr *nest_usage_report;

    nest_usage_report = nla_nest_start(skb, GTP5G_UR);
    if (!nest_usage_report)
        return -EMSGSIZE;

    if (nla_put_u32(skb, GTP5G_UR_URRID, report->urrid))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_UR_START_TIME, report->start_time, 0))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_UR_END_TIME, report->end_time, 0))
        return -EMSGSIZE;
    if (gtp5g_genl_fill_volume_measurement(skb, report->volmeasurement))
        return -EMSGSIZE;
    if (nla_put_u64_64bit(skb, GTP5G_UR_SEID, report->seid, 0))
        return -EMSGSIZE;

    if (report->trigger) {
        if (nla_put_be32(skb, GTP5G_UR_USAGE_REPORT_TRIGGER, report->trigger)) 
            return -EMSGSIZE;
    }

    nla_nest_end(skb, nest_usage_report);

    return 0;
}

static int parse_seid_urr(struct seid_urr *seid_urr, struct nlattr *a)
{
    struct nlattr *attrs[GTP5G_URR_ATTR_MAX + 1];
    int err;

    if (!seid_urr) {
        return -ENOMEM;
    }

    err = nla_parse_nested(attrs, GTP5G_URR_ATTR_MAX, a, NULL, NULL);
    if (err){
        return err;
    }

    if (attrs[GTP5G_URR_ID]) {
        seid_urr->urrid = nla_get_u32(attrs[GTP5G_URR_ID]);
    }

    if (attrs[GTP5G_URR_SEID]) {
        seid_urr->seid = nla_get_u64(attrs[GTP5G_URR_SEID]);
    }

    return 0;
}

int gtp5g_genl_fill_usage_report(struct sk_buff *skb, u32 snd_portid, u32 snd_seq,
    u32 type, struct usage_report *report)
{
    void *genlh;
    genlh = genlmsg_put(skb, snd_portid, snd_seq,
            &gtp5g_genl_family, 0, type);
    if (!genlh)
        goto genlmsg_fail;

    gtp5g_genl_fill_ur(skb, report);

    genlmsg_end(skb, genlh);

    return 0;

genlmsg_fail:
    genlmsg_cancel(skb, genlh);
    return -EMSGSIZE;
}

int gtp5g_genl_fill_multi_usage_reports(struct sk_buff *skb, u32 snd_portid, u32 snd_seq,
    u32 type, struct usage_report **reports, u32 report_num)
{
    int i;
    void *genlh;
    genlh = genlmsg_put(skb, snd_portid, snd_seq,
            &gtp5g_genl_family, 0, type);
    if (!genlh)
        goto genlmsg_fail;

    for (i = 0; i < report_num; i++) {
        gtp5g_genl_fill_ur(skb, reports[i]);
    }

    genlmsg_end(skb, genlh);

    return 0;

genlmsg_fail:
    genlmsg_cancel(skb, genlh);
    return -EMSGSIZE;
}

void convert_urr_to_report(struct urr *urr, struct usage_report *report)
{
    urr->end_time = ktime_get_real();
    *report = (struct usage_report ) {
                urr->id,
                0,
                urr->bytes,
                0,
                urr->start_time,
                urr->end_time,
                urr->seid
        };

    memset(&urr->bytes, 0, sizeof(struct VolumeMeasurement));

    urr->start_time = ktime_get_real();
}