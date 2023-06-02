#include <linux/module.h>
#include <linux/etherdevice.h>
#include <net/genetlink.h>

#include "common.h"
#include "dev.h"
#include "genl.h"
#include "genl_far.h"
#include "far.h"
#include "pktinfo.h"
#include "api_version.h"

#include <linux/rculist.h>
#include <net/netns/generic.h>
#include "net.h"

static int header_creation_fill(struct forwarding_parameter *,
                struct nlattr **, u8 *,
                struct gtp5g_emark_pktinfo *,
                uint8_t sendEndmarker);
static int forwarding_parameter_fill(struct forwarding_parameter *,
                struct nlattr **, u8 *,
                struct gtp5g_emark_pktinfo *);
static int far_fill(struct far *, struct gtp5g_dev *, struct genl_info *,
                u8 *, struct gtp5g_emark_pktinfo *);

static int gtp5g_genl_fill_far(struct sk_buff *, u32, u32, u32, struct far *);

int gtp5g_genl_add_far(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct far *far;
    int ifindex;
    int netnsfd;
    u64 seid;
    u32 far_id;
    int err;
    u8 flag;
    struct gtp5g_emark_pktinfo epkt_info;

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

    if (info->attrs[GTP5G_FAR_SEID]) {
        seid = nla_get_u64(info->attrs[GTP5G_FAR_SEID]);
    } else {
        seid = 0;
    }

    if (info->attrs[GTP5G_FAR_ID]) {
        far_id = nla_get_u32(info->attrs[GTP5G_FAR_ID]);
    } else {
        rcu_read_unlock();
        rtnl_unlock();
        return -ENODEV;
    }

    far = find_far_by_id(gtp, seid, far_id);
    if (far) {
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

        flag = 0;
        err = far_fill(far, gtp, info, &flag, &epkt_info);
        if (err) {
            far_context_delete(far);
            rcu_read_unlock();
            rtnl_unlock();
            return err;
        }

        // Send GTP-U End marker to gNB
        if (flag) {
            /* SKB size GTPU(8) + UDP(8) + IP(20) + Eth(14)
             * + 2-Bytes align the IP header
             * */
            struct sk_buff *skb = __netdev_alloc_skb(gtp->dev, 52, GFP_KERNEL);
            if (!skb) {
                rcu_read_unlock();
                rtnl_unlock();
                return 0;
            }
            skb_reserve(skb, 2);
            skb->protocol = eth_type_trans(skb, gtp->dev);
            gtp5g_fwd_emark_skb_ipv4(skb, gtp->dev, &epkt_info);
        }
        rcu_read_unlock();
        rtnl_unlock();
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

    // Check only at the creation part
    if (!info->attrs[GTP5G_FAR_APPLY_ACTION]) {
        rcu_read_unlock();
        rtnl_unlock();
        return -EINVAL;
    }

    far = kzalloc(sizeof(*far), GFP_ATOMIC);
    if (!far) {
        rcu_read_unlock();
        rtnl_unlock();
        return -ENOMEM;
    }
    far->dev = gtp->dev;

    err = far_fill(far, gtp, info, NULL, NULL);
    if (err) {
        far_context_delete(far);
        rcu_read_unlock();
        rtnl_unlock();
        return err;
    }

    far_append(seid, far_id, far, gtp);
 
    rcu_read_unlock();
    rtnl_unlock();
    return 0;
}  

int gtp5g_genl_del_far(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct far *far;
    int ifindex;
    int netnsfd;
    u64 seid;
    u32 far_id;


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

    if (info->attrs[GTP5G_FAR_SEID]) {
        seid = nla_get_u64(info->attrs[GTP5G_FAR_SEID]);
    } else {
        seid = 0;
    }

    if (info->attrs[GTP5G_FAR_ID]) {
        far_id = nla_get_u32(info->attrs[GTP5G_FAR_ID]);
    } else {
        rcu_read_unlock();
        return -ENODEV;
    }

    far = find_far_by_id(gtp, seid, far_id);
    if (!far) {
        rcu_read_unlock();
        return -ENOENT;
    }

    far_context_delete(far);
    rcu_read_unlock();

    return 0;
}   

int gtp5g_genl_get_far(struct sk_buff *skb, struct genl_info *info)
{
    struct gtp5g_dev *gtp;
    struct far *far;
    int ifindex;
    int netnsfd;
    u64 seid;
    u32 far_id;
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

    if (info->attrs[GTP5G_FAR_SEID]) {
        seid = nla_get_u64(info->attrs[GTP5G_FAR_SEID]);
    } else {
        seid = 0;
    }

    if (info->attrs[GTP5G_FAR_ID]) {
        far_id = nla_get_u32(info->attrs[GTP5G_FAR_ID]);
    } else {
        rcu_read_unlock();
        return -ENODEV;
    }

    far = find_far_by_id(gtp, seid, far_id);
    if (!far) {
        rcu_read_unlock();
        return -ENOENT;
    }

    skb_ack = genlmsg_new(NLMSG_GOODSIZE, GFP_ATOMIC);
    if (!skb_ack) {
        rcu_read_unlock();
        return -ENOMEM;
    }

    err = gtp5g_genl_fill_far(skb_ack,
            NETLINK_CB(skb).portid,
            info->snd_seq,
            info->nlhdr->nlmsg_type,
            far);
    if (err) {
        kfree_skb(skb_ack);
        rcu_read_unlock();
        return err;
    }

    rcu_read_unlock();

    return genlmsg_unicast(genl_info_net(info), skb_ack, info->snd_portid);
}

int gtp5g_genl_dump_far(struct sk_buff *skb, struct netlink_callback *cb)
{
    /* netlink_callback->args
     * args[0] : index of gtp5g dev id
     * args[1] : index of gtp5g hash entry id in dev
     * args[2] : index of gtp5g far id
     * args[5] : set non-zero means it is finished
     */
    struct gtp5g_dev *gtp;
    struct gtp5g_dev *last_gtp = (struct gtp5g_dev *)cb->args[0];
    struct net *net = sock_net(skb->sk);
    struct gtp5g_net *gn = net_generic(net, GTP5G_NET_ID());
    int i;
    int last_hash_entry_id = cb->args[1];
    int ret;
    u32 far_id = cb->args[2];
    struct far *far;

    if (cb->args[5])
        return 0;

    list_for_each_entry_rcu(gtp, &gn->gtp5g_dev_list, list) {
        if (last_gtp && last_gtp != gtp)
            continue;
        else
            last_gtp = NULL;

        for (i = last_hash_entry_id; i < gtp->hash_size; i++) {
            hlist_for_each_entry_rcu(far, &gtp->far_id_hash[i], hlist_id) {
                if (far_id && far_id != far->id)
                    continue;
                else
                    far_id = 0;

                ret = gtp5g_genl_fill_far(skb,
                        NETLINK_CB(cb->skb).portid,
                        cb->nlh->nlmsg_seq,
                        cb->nlh->nlmsg_type,
                        far);
                if (ret) {
                    cb->args[0] = (unsigned long)gtp;
                    cb->args[1] = i;
                    cb->args[2] = far->id;
                    goto out;
                }
            }
        }
    }
    cb->args[5] = 1;
out:
    return skb->len;
}


static int header_creation_fill(struct forwarding_parameter *param,
               struct nlattr **attrs, u8 *flag,
               struct gtp5g_emark_pktinfo *epkt_info,
               uint8_t sendEndmarker)
{
    struct outer_header_creation *hdr_creation;

    if (!attrs[GTP5G_OUTER_HEADER_CREATION_DESCRIPTION] ||
            !attrs[GTP5G_OUTER_HEADER_CREATION_O_TEID] ||
            !attrs[GTP5G_OUTER_HEADER_CREATION_PEER_ADDR_IPV4] ||
            !attrs[GTP5G_OUTER_HEADER_CREATION_PORT]) {
        return -EINVAL;
    }

    if (!param->hdr_creation) {
        param->hdr_creation = kzalloc(sizeof(*param->hdr_creation),
                GFP_ATOMIC);
        if (!param->hdr_creation)
            return -ENOMEM;
        hdr_creation = param->hdr_creation;
        hdr_creation->description = nla_get_u16(attrs[GTP5G_OUTER_HEADER_CREATION_DESCRIPTION]);
        hdr_creation->teid = htonl(nla_get_u32(attrs[GTP5G_OUTER_HEADER_CREATION_O_TEID]));
        hdr_creation->peer_addr_ipv4.s_addr = nla_get_be32(attrs[GTP5G_OUTER_HEADER_CREATION_PEER_ADDR_IPV4]);
        hdr_creation->port = htons(nla_get_u16(attrs[GTP5G_OUTER_HEADER_CREATION_PORT]));
    } else {
        u32 old_teid, old_peer_addr;
        u16 old_port;

        hdr_creation = param->hdr_creation;
        old_teid = hdr_creation->teid;
        old_peer_addr = hdr_creation->peer_addr_ipv4.s_addr;
        old_port = hdr_creation->port;
        hdr_creation->description = nla_get_u16(attrs[GTP5G_OUTER_HEADER_CREATION_DESCRIPTION]);
        hdr_creation->teid = htonl(nla_get_u32(attrs[GTP5G_OUTER_HEADER_CREATION_O_TEID]));
        hdr_creation->peer_addr_ipv4.s_addr = nla_get_be32(attrs[GTP5G_OUTER_HEADER_CREATION_PEER_ADDR_IPV4]);
        hdr_creation->port = htons(nla_get_u16(attrs[GTP5G_OUTER_HEADER_CREATION_PORT]));
        /* For Downlink traffic from UPF to gNB
         * In some cases,
         *  1) SMF will send PFCP Msg filled with FAR's TEID and gNB N3 addr as 0
         *  2) Later time, SMF will send PFCP Msg filled with right value in 1)

         *      2.a) We should send the GTP-U EndMarker to gNB
         *      2.b) SHOULD not set the flag as 1
         *  3) Xn Handover in b/w gNB then
         *      3.a) SMF will send modification of PDR, FAR(TEID and GTP-U)
         *      3.b) SHOULD set the flag as 1 and send GTP-U Marker for old gNB

         * */
        /* R15.3 29.281
         * 5.1 General format
         * When setting up a GTP-U tunnel, the GTP-U entity shall not assign th
         e value 'all zeros' to its own TEID.
         * However, for backward compatibility, if a GTP-U entity receives (via
         respective control plane message) a peer's
         * TEID that is set to the value 'all zeros', the GTP-U entity shall ac
         cept this value as valid and send the subsequent
         * G-PDU with the TEID field in the header set to the value 'all zeros'
         .
         * */
        if ((flag != NULL && epkt_info != NULL)) {
            if (sendEndmarker) {
                *flag = 1;
                epkt_info->teid = old_teid;
                epkt_info->peer_addr = old_peer_addr;
                epkt_info->gtph_port = old_port;
            }
        }
    }

    return 0;
}

static int forwarding_parameter_fill(struct forwarding_parameter *param,
               struct nlattr **attrs, u8 *flag,
               struct gtp5g_emark_pktinfo *epkt_info)
{
    struct nlattr *hdr_creation_attrs[GTP5G_OUTER_HEADER_CREATION_ATTR_MAX + 1];
    struct forwarding_policy *fwd_policy;
    uint8_t sendEndmarker = 0;
    int err;

    if (attrs[GTP5G_FORWARDING_PARAMETER_OUTER_HEADER_CREATION]) {
        err = nla_parse_nested(hdr_creation_attrs,
                GTP5G_OUTER_HEADER_CREATION_ATTR_MAX,
                attrs[GTP5G_FORWARDING_PARAMETER_OUTER_HEADER_CREATION],
                NULL,
                NULL);
        if (err)
            return err;

        /*
            TS 29.244 PFCPSMReq-Flags
            SNDEM (Send End Marker Packets): 
                if this bit is set to "1", it indicates that the UP function 
                shall construct and send End Marker packets
        */
        #define SNDEM 0x02
        if (attrs[GTP5G_FORWARDING_PARAMETER_PFCPSM_REQ_FLAGS]) {
            sendEndmarker = nla_get_u8(attrs[GTP5G_FORWARDING_PARAMETER_PFCPSM_REQ_FLAGS]) & SNDEM;       
        }
        err = header_creation_fill(param, hdr_creation_attrs, flag, epkt_info, sendEndmarker);
        if (err)
            return err;
    }

    if (attrs[GTP5G_FORWARDING_PARAMETER_FORWARDING_POLICY]) {
        if (!param->fwd_policy) {
            param->fwd_policy = kzalloc(sizeof(*param->fwd_policy), GFP_ATOMIC);
            if (!param->fwd_policy)
                return -ENOMEM;
        }
        fwd_policy = param->fwd_policy;
        fwd_policy->len = nla_len(attrs[GTP5G_FORWARDING_PARAMETER_FORWARDING_POLICY]);
        if (fwd_policy->len >= sizeof(fwd_policy->identifier))
            return -EINVAL;
        strncpy(fwd_policy->identifier,
                nla_data(attrs[GTP5G_FORWARDING_PARAMETER_FORWARDING_POLICY]), fwd_policy->len);

        /* Exact value to handle forwarding policy */
        if (!(fwd_policy->mark = simple_strtol(fwd_policy->identifier, NULL, 10))) {
            return -EINVAL;
        }
    }

    return 0;
}


static int far_fill(struct far *far, struct gtp5g_dev *gtp, struct genl_info *info,
        u8 *flag, struct gtp5g_emark_pktinfo *epkt_info)
{
    struct nlattr *attrs[GTP5G_FORWARDING_PARAMETER_ATTR_MAX + 1];
    int err;
    struct forwarding_parameter *fwd_param;

    if (!far)
        return -EINVAL;

    far->id = nla_get_u32(info->attrs[GTP5G_FAR_ID]);

    if (info->attrs[GTP5G_FAR_SEID])
        far->seid = nla_get_u64(info->attrs[GTP5G_FAR_SEID]);
    else
        far->seid = 0;

    if (info->attrs[GTP5G_FAR_APPLY_ACTION])
        switch (nla_len(info->attrs[GTP5G_FAR_APPLY_ACTION]))
        {
        case FAR_ACTION_U16:
            set_far_action_u16(true);
            far->action = nla_get_u16(info->attrs[GTP5G_FAR_APPLY_ACTION]);
            break;
        case FAR_ACTION_U8:
            set_far_action_u16(false);
            far->action = nla_get_u8(info->attrs[GTP5G_FAR_APPLY_ACTION]);
            break;
        default:
            break;
        }

    if (info->attrs[GTP5G_FAR_FORWARDING_PARAMETER]) {
        err = nla_parse_nested(attrs,
                GTP5G_FORWARDING_PARAMETER_ATTR_MAX,
                info->attrs[GTP5G_FAR_FORWARDING_PARAMETER],
                NULL,
                NULL);
        if (err)
            return err;
        fwd_param = rcu_dereference(far->fwd_param);
        if (!fwd_param) {
            fwd_param = kzalloc(sizeof(*fwd_param), GFP_ATOMIC);
            if (!fwd_param)
                return -ENOMEM;
        }
        err = forwarding_parameter_fill(fwd_param, attrs, flag, epkt_info);
        rcu_assign_pointer(far->fwd_param, fwd_param);
        if (err)
            return err;
    }

    /* Update PDRs which has not linked to this FAR */
    far_update(far, gtp, flag, epkt_info);

    return 0;
}


static int gtp5g_genl_fill_far(struct sk_buff *skb, u32 snd_portid, u32 snd_seq,
        u32 type, struct far *far)
{
    struct gtp5g_dev *gtp = netdev_priv(far->dev);
    void *genlh;
    struct nlattr *nest_fwd_param;
    struct nlattr *nest_hdr_creation;
    struct forwarding_parameter *fwd_param;
    struct outer_header_creation *hdr_creation;
    struct forwarding_policy *fwd_policy;
    u16 *ids = NULL;
    int n;

    genlh = genlmsg_put(skb, snd_portid, snd_seq, &gtp5g_genl_family, 0, type);
    if (!genlh)
        goto genlmsg_fail;

    if (nla_put_u32(skb, GTP5G_FAR_ID, far->id))
        goto genlmsg_fail;

    if (far_action_is_u16()) {
        if (nla_put_u16(skb, GTP5G_FAR_APPLY_ACTION, far->action))
            goto genlmsg_fail;
    } else {
        if (nla_put_u8(skb, GTP5G_FAR_APPLY_ACTION, far->action))
            goto genlmsg_fail;
    }

    if (far->seid) {
        if (nla_put_u64_64bit(skb, GTP5G_FAR_SEID, far->seid, 0))
            goto genlmsg_fail;
    }
    fwd_param = rcu_dereference(far->fwd_param);
    if (fwd_param) {
        nest_fwd_param = nla_nest_start(skb, GTP5G_FAR_FORWARDING_PARAMETER);
        if (!nest_fwd_param)
            goto genlmsg_fail;

        if (fwd_param->hdr_creation) {
            nest_hdr_creation = nla_nest_start(skb, GTP5G_FORWARDING_PARAMETER_OUTER_HEADER_CREATION);
            if (!nest_hdr_creation)
                goto genlmsg_fail;

            hdr_creation = fwd_param->hdr_creation;
            if (nla_put_u16(skb, GTP5G_OUTER_HEADER_CREATION_DESCRIPTION, hdr_creation->description))
                goto genlmsg_fail;
            if (nla_put_u32(skb, GTP5G_OUTER_HEADER_CREATION_O_TEID, ntohl(hdr_creation->teid)))
                goto genlmsg_fail;
            if (nla_put_be32(skb, GTP5G_OUTER_HEADER_CREATION_PEER_ADDR_IPV4, hdr_creation->peer_addr_ipv4.s_addr))
                goto genlmsg_fail;
            if (nla_put_u16(skb, GTP5G_OUTER_HEADER_CREATION_PORT, ntohs(hdr_creation->port)))
                goto genlmsg_fail;

            nla_nest_end(skb, nest_hdr_creation);
        }

        if ((fwd_policy = fwd_param->fwd_policy))
            if (nla_put(skb, GTP5G_FORWARDING_PARAMETER_FORWARDING_POLICY, fwd_policy->len, fwd_policy->identifier))
                goto genlmsg_fail;

        nla_nest_end(skb, nest_fwd_param);
    }

    ids = kzalloc(MAX_PDR_PER_SESSION * sizeof(u16), GFP_KERNEL);
    if (!ids)
        goto genlmsg_fail;
    n = far_get_pdr_ids(ids, MAX_PDR_PER_SESSION, far, gtp);
    if (n) {
        if (nla_put(skb, GTP5G_FAR_RELATED_TO_PDR, n * sizeof(u16), ids))
            goto genlmsg_fail;
    }

    kfree(ids);
    genlmsg_end(skb, genlh);
    return 0;

genlmsg_fail:
    if (ids)
        kfree(ids);
    genlmsg_cancel(skb, genlh);
    return -EMSGSIZE;
}
