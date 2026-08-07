#include <linux/version.h>
#include <linux/socket.h>
#include <linux/rculist.h>
#include <linux/udp.h>
#include <linux/gtp.h>

#include <net/ip.h>
#include <net/udp.h>
#include <net/udp_tunnel.h>

#include "dev.h"
#include "link.h"
#include "encap.h"
#include "gtp.h"
#include "pdr.h"
#include "far.h"
#include "qer.h"
#include "urr.h"
#include "report.h"
#include "util.h"

#include "genl.h"
#include "genl_report.h"

#include "log.h"
#include "api_version.h"
#include "pktinfo.h"

#if LINUX_VERSION_CODE >= KERNEL_VERSION(6, 18, 0)
#include <net/inet_dscp.h>
#endif

/* used to compatible with api with/without seid */
#define MSG_KOV_LEN 4

enum msg_type {
    TYPE_BUFFER = 1,
    TYPE_URR_REPORT,
    TYPE_BAR_INFO,
};

static void gtp5g_encap_disable_locked(struct sock *);
static int gtp5g_encap_recv(struct sock *, struct sk_buff *);
static int gtp1u_udp_encap_recv(struct gtp5g_dev *, struct sk_buff *);
static int gtp5g_rx(struct pdr *, struct sk_buff *, unsigned int, unsigned int, u64 *);
static int gtp5g_fwd_skb_encap(struct sk_buff *, struct net_device *,
        unsigned int, struct pdr *, struct far *, u64 *);
static int netlink_send(struct pdr *, struct far *, struct sk_buff *, struct net *, struct usage_report *, u32);
static int unix_sock_send(struct pdr *, struct far *, void *, u32, u32);
static int gtp5g_fwd_skb_ipv4(struct sk_buff *, 
    struct net_device *, struct gtp5g_pktinfo *, 
    struct pdr *, struct far *);

/* When gtp5g newlink, establish the udp tunnel used in N3 interface */
struct sock *gtp5g_encap_enable(int fd, int type, struct gtp5g_dev *gtp){
    struct udp_tunnel_sock_cfg tuncfg = {NULL};
    struct socket *sock;
    struct sock *sk;
    int err;

    GTP5G_LOG(NULL, "enable gtp5g for the fd(%d) type(%d)\n", fd, type);

    sock = sockfd_lookup(fd, &err);
    if (!sock) {
        GTP5G_ERR(NULL, "Failed to find the socket fd(%d)\n", fd);
        return NULL;
    }

    if (sock->sk->sk_protocol != IPPROTO_UDP) {
        GTP5G_ERR(NULL, "socket fd(%d) is not a UDP\n", fd);
        sk = ERR_PTR(-EINVAL);
        goto out_sock;
    }

    lock_sock(sock->sk);
    if (sock->sk->sk_user_data) {
        GTP5G_ERR(NULL, "Failed to set sk_user_datat of socket fd(%d)\n", fd);
        sk = ERR_PTR(-EBUSY);
        goto out_sock;
    }

    sk = sock->sk;
    sock_hold(sk);

    tuncfg.sk_user_data = gtp;
    tuncfg.encap_type = type;
    tuncfg.encap_rcv = gtp5g_encap_recv;
    tuncfg.encap_destroy = gtp5g_encap_disable_locked;

    setup_udp_tunnel_sock(sock_net(sock->sk), sock, &tuncfg);

out_sock:
    release_sock(sock->sk);
    sockfd_put(sock);
    return sk;
}


void gtp5g_encap_disable(struct sock *sk)
{
    struct gtp5g_dev *gtp;

    if (!sk) {
        return;
    }

    lock_sock(sk);
    gtp = sk->sk_user_data;
    if (gtp) {
        gtp->sk1u = NULL;
        udp_sk(sk)->encap_type = 0;
        // sk_user_data was protected by RCU
        rcu_assign_sk_user_data(sk, NULL);
        sock_put(sk);
    }
    release_sock(sk);
}

static void gtp5g_encap_disable_locked(struct sock *sk)
{
    rtnl_lock();
    gtp5g_encap_disable(sk);
    rtnl_unlock();
}

/**
 * Entry function for Uplink packets
 * */
static int gtp5g_encap_recv(struct sock *sk, struct sk_buff *skb)
{
    struct gtp5g_dev *gtp;
    int ret = 0;

    gtp = rcu_dereference_sk_user_data(sk);
    if (!gtp) {
        return 1;
    }

    switch (udp_sk(sk)->encap_type) {
    case UDP_ENCAP_GTP1U:
        ret = gtp1u_udp_encap_recv(gtp, skb);
        break;
    default:
        ret = -1; // Should not happen
    }

    switch (ret) {
    case PKT_TO_APP: // packet that gtp5g cannot handle
        GTP5G_TRC(gtp->dev, "Pass up to the process\n");
        break;
    case PKT_FORWARDED:
        break;
    case PKT_DROPPED:
        GTP5G_TRC(gtp->dev, "GTP packet has been dropped\n");
        kfree_skb(skb);
        ret = 0;
        break;
    case PKT_DROPPED_AND_FREED:
        GTP5G_TRC(gtp->dev, "GTP packet has been dropped and already freed\n");
        ret = 0;
        break;
    default:
        GTP5G_ERR(gtp->dev, "Unhandled return value from gtp1u_udp_encap_recv\n");
    }

    return ret;
}

static int gtp1c_handle_echo_req(struct sk_buff *skb, struct gtp5g_dev *gtp)
{
    struct gtpv1_hdr *req_gtp1;
    struct gtp1_hdr_opt *req_gtpOptHdr;

    struct gtpv1_echo_resp *gtp_pkt;

    struct rtable *rt;
    struct flowi4 fl4;
    struct iphdr *iph;
    struct udphdr *udph;

    __u8   flags = 0;
    __be16 seq_number = 0;

    req_gtp1 = (struct gtpv1_hdr *)(skb->data + sizeof(struct udphdr));

    flags = req_gtp1->flags;
    if (flags & GTPV1_HDR_FLG_SEQ) {
         req_gtpOptHdr = (struct gtp1_hdr_opt *)(skb->data + sizeof(struct udphdr) 
                                                            + sizeof(struct gtpv1_hdr));
         seq_number = req_gtpOptHdr->seq_number;
    } else {
        GTP5G_ERR(gtp->dev, "GTP echo request shall bring sequence number\n");
        seq_number = 0;
    }

    pskb_pull(skb, skb->len);          

    gtp_pkt = skb_push(skb, sizeof(struct gtpv1_echo_resp));
    if (!gtp_pkt) {
        GTP5G_ERR(gtp->dev, "can not construct GTP Echo Response\n");
        return PKT_DROPPED;
    }
    memset(gtp_pkt, 0, sizeof(struct gtpv1_echo_resp));

    /* gtp header*/
    gtp_pkt->gtpv1_h.flags = GTPV1 | GTPV1_HDR_FLG_SEQ;
    gtp_pkt->gtpv1_h.type = GTPV1_MSG_TYPE_ECHO_RSP;
    gtp_pkt->gtpv1_h.length =
        htons(sizeof(struct gtpv1_echo_resp) - sizeof(struct gtpv1_hdr));
    gtp_pkt->gtpv1_h.tid = 0;

    /* gtp opt header*/
    gtp_pkt->gtpv1_opt_h.seq_number = seq_number;

    /* gtp recovery*/
    gtp_pkt->recov.type_num = GTPV1_IE_RECOVERY;
    gtp_pkt->recov.cnt = 0;

    iph = ip_hdr(skb);
    udph = udp_hdr(skb);
  
    rt = ip4_find_route(skb, iph, gtp->sk1u, gtp->dev, 
        iph->daddr ,
        iph->saddr, 
        &fl4);
    if (IS_ERR(rt)) {
        GTP5G_ERR(gtp->dev, "no route for GTP echo response from %pI4\n", 
        &iph->saddr);
        return PKT_DROPPED;
    }

    udp_tunnel_xmit_skb(rt, gtp->sk1u, skb,
                    fl4.saddr, fl4.daddr,
                    iph->tos,
                    ip4_dst_hoplimit(&rt->dst),
                    0,
                    udph->dest, udph->source,
                    !net_eq(sock_net(gtp->sk1u),
                        dev_net(gtp->dev)),
#if LINUX_VERSION_CODE >= KERNEL_VERSION(6,17,0)
                    false,
                    0);
#else
                    false);
#endif

    return PKT_FORWARDED;
}

static int gtp1u_udp_encap_recv(struct gtp5g_dev *gtp, struct sk_buff *skb)
{
    unsigned int hdrlen = sizeof(struct udphdr) + sizeof(struct gtpv1_hdr);
    struct gtpv1_hdr *gtpv1;
    struct pdr *pdr;
    unsigned int pull_len = hdrlen;
    u8 gtp_type;
    u32 teid;
    int rt = 0;
    u64 rxVol = skb->len - sizeof(struct udphdr); // exclude UDP header of GTP packet
    /* Captured by the forwarding path while the skb is still owned by us.
     * skb->len must NOT be read after gtp5g_rx() returns: the skb has by then
     * been handed to ip_xmit()/netif_rx() or freed, so reading it is a
     * use-after-free that corrupts the tx byte counters.
     * */
    u64 txVol = 0;

    if (!pskb_may_pull(skb, pull_len)) {
        GTP5G_ERR(gtp->dev, "Failed to pull skb length %#x\n", pull_len);
        rt = PKT_DROPPED;
        goto end;
    }

    gtpv1 = (struct gtpv1_hdr *)(skb->data + sizeof(struct udphdr));
    if ((gtpv1->flags >> 5) != GTP_V1) {
        GTP5G_ERR(gtp->dev, "GTP version is not v1: %#x\n",
            gtpv1->flags);
        rt = PKT_TO_APP;
        goto end;
    }

    gtp_type = gtpv1->type;
    teid = gtpv1->tid;
    if (gtp_type == GTPV1_MSG_TYPE_ECHO_REQ) {
        GTP5G_INF(gtp->dev, "GTP-C message type is GTP echo request: %#x\n",
            gtp_type);

        rt = gtp1c_handle_echo_req(skb, gtp);
        goto end;
    }

    if (gtp_type != GTPV1_MSG_TYPE_TPDU && gtp_type != GTPV1_MSG_TYPE_EMARK) {
        GTP5G_ERR(gtp->dev, "GTP-U message type is not a TPDU or End Marker: %#x\n",
            gtp_type);
        rt = PKT_TO_APP;
        goto end;
    }

    /** TS 29.281 Chapter 5.1 and Figure 5.1-1
     * GTP-U header at least 8 byte
     *
     * This field shall be present if and only if any one or more of the S, PN and E flags are set.
     * This field means seq number (2 Octect), N-PDU number (1 Octet) and  Next ext hdr type (1 Octet).
     *
     * TODO: Validate the Reserved flag set or not, if it is set then protocol error
     */
    if (gtpv1->flags & GTPV1_HDR_FLG_MASK) {
        u8 *ext_hdr = NULL;
        hdrlen += sizeof(struct gtp1_hdr_opt);
        pull_len = hdrlen;
        if (!pskb_may_pull(skb, pull_len)) {
            GTP5G_ERR(gtp->dev, "Failed to pull skb length %#x\n", pull_len);
            rt = PKT_DROPPED;
            goto end;
        }

        /** TS 29.281 Chapter 5.2 and Figure 5.2.1-1
         * The length of the Extension header shall be defined in a variable length of 4 octets,
         * i.e. m+1 = n*4 octets, where n is a positive integer.
         */
        while (*(ext_hdr = (u8 *)(skb->data + hdrlen - 1))) {
            u8 ext_hdr_type = *ext_hdr;
            unsigned int extlen;
            pull_len = hdrlen + 1; // 1 byte for the length of extension hdr
            if (!pskb_may_pull(skb, pull_len)) {
                GTP5G_ERR(gtp->dev, "Failed to pull skb length %#x\n", pull_len);
                rt = PKT_DROPPED;
                goto end;
            }
            extlen = (*((u8 *)(skb->data + hdrlen))) * 4; // total length of extension hdr
            if (extlen == 0) {
                GTP5G_ERR(gtp->dev, "Invalid extention header length\n");
                rt = PKT_DROPPED;
                goto end;
            }
            hdrlen += extlen;
            pull_len = hdrlen;
            if (!pskb_may_pull(skb, pull_len)) {
                GTP5G_ERR(gtp->dev, "Failed to pull skb length %#x\n", pull_len);
                rt = PKT_DROPPED;
                goto end;
            }
            switch (ext_hdr_type) {
                case GTPV1_NEXT_EXT_HDR_TYPE_85:
                {
                    // ext_pdu_sess_ctr_t *etype85 = (ext_pdu_sess_ctr_t *) (skb->data + hdrlen);
                    // pdu_sess_ctr_t *pdu_sess_info = &etype85->pdu_sess_ctr;

                    // Commented the below code due to support N9 packet downlink
                    // if (pdu_sess_info->type_spare == PDU_SESSION_INFO_TYPE0)
                    //     return -1;

                    //TODO: validate pdu_sess_ctr
                    break;
                }
            }
        }
    }

    pdr = pdr_find_by_gtp1u(gtp, skb, hdrlen, teid, gtp_type);
    if (!pdr) {
        GTP5G_ERR(gtp->dev, "No PDR match this skb : teid[%x]\n", ntohl(teid));
        rt = PKT_DROPPED;
        goto end;
    }

    rt = gtp5g_rx(pdr, skb, hdrlen, gtp->role, &txVol);

end:
    if (pdr && pdr->pdi) {
        update_usage_statistic(gtp, rxVol, txVol, rt, pdr->pdi->srcIntf);
    } else {
        update_usage_statistic(gtp, rxVol, txVol, rt, SRC_INTF_ACCESS);
    }

    return rt;
}

static int gtp5g_drop_skb_encap(struct sk_buff *skb, struct net_device *dev, 
    struct pdr *pdr)
{
    struct gtpv1_hdr *gtp1 = (struct gtpv1_hdr *)(skb->data + sizeof(struct udphdr));
    if (gtp1->type == GTPV1_MSG_TYPE_TPDU) {
        pdr->ul_drop_cnt++;
        GTP5G_INF(NULL, "PDR (%u) UL_DROP_CNT (%llu)", pdr->id, pdr->ul_drop_cnt);
    }
    
    // release skb in outer function
    return PKT_DROPPED;
}

static int gtp5g_buf_skb_encap(struct sk_buff *skb, struct net_device *dev,
    unsigned int hdrlen, struct pdr *pdr, struct far *far, u64 *txVol)
{
    struct gtpv1_hdr *gtp1 = (struct gtpv1_hdr *)(skb->data + sizeof(struct udphdr));
    if (gtp1->type == GTPV1_MSG_TYPE_TPDU) {
        // Get rid of the GTP-U + UDP headers.
        if (iptunnel_pull_header(skb,
                hdrlen,
                skb->protocol,
                !net_eq(sock_net(pdr->sk), dev_net(dev)))) {
            GTP5G_ERR(dev, "Failed to pull GTP-U and UDP headers\n");
            return PKT_DROPPED;
        }
        // Increment the sequence number for the packet
        far->seq_number++;

        if (pdr_addr_is_netlink(pdr)) {
            if (netlink_send(pdr, far, skb, dev_net(dev), NULL, 0) < 0) {
                GTP5G_ERR(dev, "Failed to send skb to netlink socket PDR(%u)", pdr->id);
                ++pdr->ul_drop_cnt;
            }
        } else {
            if (unix_sock_send(pdr, far, skb->data, skb_headlen(skb), 0) < 0) {
                GTP5G_ERR(dev, "Failed to send skb to unix domain socket PDR(%u)", pdr->id);
                ++pdr->ul_drop_cnt;
            }
        }
    }
    /* Last point at which the skb is still valid: neither netlink_send() nor
     * unix_sock_send() takes ownership, they only copy out of it. This matches
     * what the caller used to read at its end label, for both the TPDU case
     * (headers already pulled) and the non-TPDU case (untouched).
     * */
    *txVol = skb->len;
    dev_kfree_skb(skb);
    return PKT_FORWARDED;
}

/* Function netlink_{...} are used to handle buffering */
// Send PDR ID, FAR action and buffered packet to user space
static int netlink_send(struct pdr *pdr, struct far *far, struct sk_buff *skb_in, struct net *net, struct usage_report *reports, u32 report_num)
{
    struct sk_buff *skb;
    static atomic_t seq_counter;
    u32 seq;
    void *header;
    struct nlattr *attr;
    int i, err;
    struct nlattr *nest_msg_type;

    if (reports != NULL) {
        skb = genlmsg_new(NLMSG_GOODSIZE, GFP_ATOMIC);
    } else {
        skb = genlmsg_new(
            nla_total_size_64bit(8) + // SEID
                nla_total_size(2) + // PDR ID
                nla_total_size(2) + // Action
                nla_total_size(2) + // Sequence Num
                nla_total_size(skb_in->len), // Buff Pkt Size
            GFP_ATOMIC);
    }

    if (!skb)
        return -ENOMEM;

    seq = atomic_inc_return(&seq_counter);
    header = genlmsg_put(skb, 0, seq, &gtp5g_genl_family, 0, GTP5G_CMD_BUFFER_GTPU);
    if (!header)
    {
        nlmsg_free(skb);
        return -ENOMEM;
    }

    if (reports != NULL) {
        nest_msg_type = nla_nest_start(skb, GTP5G_REPORT);

        for (i = 0; i < report_num; i++) {
            gtp5g_genl_fill_ur(skb, &reports[i]);
        }
        
        nla_nest_end(skb, nest_msg_type);
    } else {
        nest_msg_type = nla_nest_start(skb, GTP5G_BUFFER);

        err = nla_put_u16(skb, GTP5G_BUFFER_ID, pdr->id);
        if (err != 0) {
            nlmsg_free(skb);
            return err;
        }

        err = nla_put_u64_64bit(skb, GTP5G_BUFFER_SEID, pdr->seid, GTP5G_BUFFER_PAD);
        if (err != 0) {
            nlmsg_free(skb);
            return err;
        }

        err = nla_put_u16(skb, GTP5G_BUFFER_ACTION, far->action);
        if (err != 0) {
            nlmsg_free(skb);
            return err;
        }

        err = nla_put_u16(skb, GTP5G_BUFFER_SEQ_NUMBER, far->seq_number);
        if (err != 0) {
            nlmsg_free(skb);
            return err;
        }

        attr = nla_reserve(skb, GTP5G_BUFFER_PACKET, skb_in->len);
        if (!attr) {
            nlmsg_free(skb);
            return -EINVAL;
        }
        skb_copy_bits(skb_in, 0, nla_data(attr), skb_in->len);

        nla_nest_end(skb, nest_msg_type);
    }

    genlmsg_end(skb, header);
    genlmsg_multicast_netns(&gtp5g_genl_family, net, skb, 0, GTP5G_GENL_MCGRP, GFP_ATOMIC);
    return 0;
}

/* Function unix_sock_{...} are used to handle buffering */
// Send PDR ID, FAR action and buffered packet to user space
static int unix_sock_send(struct pdr *pdr, struct far *far, void *buf, u32 len, u32 report_num)
{
    struct msghdr msg;
    struct kvec *kov;
    struct socket *sock = pdr->sock_for_buf;
    int msg_kovlen;
    int total_kov_len = 0;
    int i, rt;
    u8  type_hdr[1] = {TYPE_BUFFER};
    u64 self_seid_hdr[1] = {pdr->seid};
    u16 self_hdr[2] = {pdr->id, far->action};
    u32 self_num_hdr[1] = {report_num};

    if (pdr_addr_is_netlink(pdr)) {
        return -EINVAL;
    }

    if (!sock) {
        GTP5G_ERR(NULL, "Failed Socket is NULL\n");
        return -EINVAL;
    }

    memset(&msg, 0, sizeof(msg));
    if (report_num > 0) {
        type_hdr[0] = TYPE_URR_REPORT;
    }

    msg_kovlen = MSG_KOV_LEN;
    kov = kmalloc_array(msg_kovlen, sizeof(struct kvec),
        GFP_KERNEL);
    if (kov == NULL)
        return -ENOMEM;

    memset(kov, 0, sizeof(struct kvec) * msg_kovlen);

    kov[0].iov_base = type_hdr;
    kov[0].iov_len = sizeof(type_hdr);
    kov[1].iov_base = self_seid_hdr;
    kov[1].iov_len = sizeof(self_seid_hdr);

    //buf is report in this case
    if (report_num > 0) {
        kov[2].iov_base = self_num_hdr;
        kov[2].iov_len = sizeof(self_num_hdr);
    }
    else {
        kov[2].iov_base = self_hdr;
        kov[2].iov_len = sizeof(self_hdr);
    }

    kov[3].iov_base = buf;
    kov[3].iov_len = len;

    for (i = 0; i < msg_kovlen; i++)
        total_kov_len += kov[i].iov_len;

    msg.msg_name = 0;
    msg.msg_namelen = 0;
    iov_iter_kvec(&msg.msg_iter, WRITE, kov, msg_kovlen, total_kov_len);
    msg.msg_control = NULL;
    msg.msg_controllen = 0;
    msg.msg_flags = MSG_DONTWAIT;

    rt = sock_sendmsg(sock, &msg);
    kfree(kov);

    return rt;
}

void update_counter(struct VolumeMeasurement *volmeasure, u64 vol, bool uplink, bool mnop){
    if (!volmeasure) {
        return;
    }

    if (vol == 0) {
        return;
    }

    if (mnop) {
        if (uplink) {
            volmeasure->uplinkPktNum++;
        } else {
            volmeasure->downlinkPktNum++;
        }
        volmeasure->totalPktNum = volmeasure->uplinkPktNum + volmeasure->downlinkPktNum;
    }

    if (uplink) {
        volmeasure->uplinkVolume += vol;
    } else {
        volmeasure->downlinkVolume += vol;
    }

    volmeasure->totalVolume = volmeasure->uplinkVolume + volmeasure->downlinkVolume;
}

bool check_counter(struct VolumeMeasurement *volmeasure, struct Volume *volume){
    if (!volmeasure) {
        return false;
    }

    if (!volume) {
        return false;
    }

    if (volume->totalVolume && (volmeasure->totalVolume >= volume->totalVolume) && (volume->flag & URR_VOLUME_TOVOL)) {
        return true;
    } else if (volume->uplinkVolume && (volmeasure->uplinkVolume >= volume->uplinkVolume) && (volume->flag & URR_VOLUME_ULVOL)) {
        return true;
    } else if (volume->downlinkVolume && (volmeasure->downlinkVolume >= volume->downlinkVolume) && (volume->flag & URR_VOLUME_DLVOL)) {
        return true;
    }

    return false;
}

static struct VolumeMeasurement *get_urr_counter_by_trigger(struct urr *urr, u32 trigger) {
    if (!urr)
        return NULL;

    if (trigger & URR_RPT_TRIGGER_VOLTH) {
        return &urr->vol_th;
    }
    else if (trigger & URR_RPT_TRIGGER_VOLQU) {
        return &urr->vol_qu;
    }
    return NULL;
}

static inline void update_period_vol_counter(struct urr *urr, u64 vol, bool uplink, bool mnop) {
    struct VolumeMeasurement *urr_counter = NULL;

    u64_stats_update_begin(&urr->period_vol_counter_sync);
    urr_counter = get_period_vol_counter(urr, urr->use_vol2);
    update_counter(urr_counter, vol, uplink, mnop);
    u64_stats_update_end(&urr->period_vol_counter_sync);
}

int update_urr_counter_and_send_report(struct pdr *pdr, struct far *far, u64 vol, u64 vol_mbqe) {
    struct gtp5g_dev *gtp = netdev_priv(pdr->dev);
    int i;
    int ret = 1;
    u64 volume;
    u64 len;
    u32 report_num = 0;
    u32 *triggers = NULL;
    struct urr *urr, **urrs = NULL;
    struct usage_report *report = NULL;
    struct VolumeMeasurement *urr_counter = NULL;
    bool mnop;
    struct sk_buff *skb;
    bool uplink = false;

    // Determine if the packet is uplink or downlink
    if (is_uplink(pdr)) {
        uplink = true;
    } else if (is_downlink(pdr)) {
        uplink = false;
    } else {
        GTP5G_ERR(pdr->dev, "PDR(%u) is not uplink or downlink", pdr->id);
        return -1;
    }

    // vol_mbqe(volume of measurement before QoS enforcement) is zero(payload is zero), 
    // no need to add volume and packet count
    if (vol_mbqe == 0) {
        return ret;
    }

    urrs = kzalloc(sizeof(struct urr *) * pdr->urr_num , GFP_ATOMIC);
    triggers = kzalloc(sizeof(u32) * pdr->urr_num , GFP_ATOMIC);
    if (!urrs || !triggers) {
        ret = -1;
        goto err1;
    }

    for (i = 0; i < pdr->urr_num; i++) {
        urr = find_urr_by_id(gtp, pdr->seid,  pdr->urr_ids[i]);
        if (!urr)
           continue;

        mnop = false;
        if (urr->info & URR_INFO_MNOP)
            mnop = true;

        if (!(urr->info & URR_INFO_INAM)) {
            if (urr->method & URR_METHOD_VOLUM) {
                if (urr->trigger == 0) {
                    GTP5G_ERR(pdr->dev, "no supported trigger(%u) in URR(%u) and related to PDR(%u)",
                        urr->trigger, urr->id, pdr->id);
                    ret = 0;
                    goto err1;
                }

                if ((urr->trigger & URR_RPT_TRIGGER_START) && uplink) {
                    triggers[report_num] = USAR_TRIGGER_START;
                    urrs[report_num++] = urr;
                    urr_quota_exhaust_action(urr, gtp);
                    GTP5G_TRC(NULL, "URR (%u) Start of Service Data Flow, stop measure until recieve quota", urr->id);
                    continue;
                }

                if (urr->info & URR_INFO_MBQE) {
                    volume = vol_mbqe;
                } else {
                    volume = vol;
                }

                // Calculate Volume measurement for each trigger
                if (urr->trigger & URR_RPT_TRIGGER_VOLTH) {
                    update_counter(&urr->vol_th, volume, uplink, mnop);
                    if (check_counter(&urr->vol_th, &urr->volumethreshold)) {
                        triggers[report_num] = USAR_TRIGGER_VOLTH;
                        urrs[report_num++] = urr;
                    }
                }
                if (urr->period != 0) {
                    update_period_vol_counter(urr, volume, uplink, mnop);
                }
                if (urr->trigger & URR_RPT_TRIGGER_VOLQU) {
                    update_counter(&urr->vol_qu, volume, uplink, mnop);
                    if (check_counter(&urr->vol_qu, &urr->volumequota)) {
                        triggers[report_num] = USAR_TRIGGER_VOLQU;
                        urrs[report_num++] = urr;
                        urr_quota_exhaust_action(urr, gtp);
                        GTP5G_INF(NULL, "URR (%u) Quota Exhaust, stop measure", urr->id);
                    }
                }
            }
        } else {
            GTP5G_TRC(pdr->dev, "URR stop measurement");
        }
    }
    if (report_num > 0) {
        len = sizeof(*report) * report_num;

        report = kzalloc(len, GFP_ATOMIC);
        if (!report) {
            ret = -1;
            goto err1;
        }

        for (i = 0; i < report_num; i++) {
            // TODO: FAR ID for Quota Action IE for indicating the action while no quota is granted
            if (triggers[i] == USAR_TRIGGER_START){
                ret = DONT_SEND_UL_PACKET;
            }
            urr_counter = get_urr_counter_by_trigger(urrs[i], triggers[i]);
            convert_urr_to_report(urrs[i], urr_counter, &report[i]);

            report[i].trigger = triggers[i];
        }

        if (pdr_addr_is_netlink(pdr)) {
            if (netlink_send(pdr, far, skb, dev_net(pdr->dev), report, report_num) < 0) {
                GTP5G_ERR(pdr->dev, "Failed to send report to netlink socket PDR(%u)", pdr->id);
                ret = -1;
                goto err1;
            }
        } else {
            if (unix_sock_send(pdr, far, report, len, report_num) < 0) {
                GTP5G_ERR(pdr->dev, "Failed to send report to unix domain socket PDR(%u)", pdr->id);
                ret = -1;
                goto err1;
            }
        }
    }

err1:
    if (report) {
        kfree(report);
    }
    if (urrs) {
        kfree(urrs);
    }
    if (triggers) {
        kfree(triggers);
    }
    return ret;
}

static int gtp5g_rx(struct pdr *pdr, struct sk_buff *skb,
    unsigned int hdrlen, unsigned int role, u64 *txVol)
{
    int rt = -1;
    struct far *far = rcu_dereference(pdr->far);

    if (!far) {
        GTP5G_ERR(pdr->dev, "FAR not exists for PDR(%u)\n", pdr->id);
        goto out;
    }

    // TODO: not reading the value of outer_header_removal now,
    // just check if it is assigned.
    if (pdr->outer_header_removal) {
        // One and only one of the DROP, FORW and BUFF flags shall be set to 1.
        // The NOCP flag may only be set if the BUFF flag is set.
        // The DUPL flag may be set with any of the DROP, FORW, BUFF and NOCP flags.
        switch(far->action & FAR_ACTION_MASK) {
        case FAR_ACTION_DROP:
            rt = gtp5g_drop_skb_encap(skb, pdr->dev, pdr);
            break;
        case FAR_ACTION_FORW:
            if (pdr->ul_dl_gate & QER_UL_GATE_CLOSE) {
                GTP5G_TRC(pdr->dev, "QER UL gate is closed, drop the packet");
                return PKT_DROPPED;
            }
            rt = gtp5g_fwd_skb_encap(skb, pdr->dev, hdrlen, pdr, far, txVol);
            break;
        case FAR_ACTION_BUFF:
            rt = gtp5g_buf_skb_encap(skb, pdr->dev, hdrlen, pdr, far, txVol);
            break;
        default:
            GTP5G_ERR(pdr->dev, "Unhandled apply action(%u) in FAR(%u) and related to PDR(%u)\n",
                far->action, far->id, pdr->id);
        }
        goto out;
    } 

    // TODO: this action is not supported
    GTP5G_ERR(pdr->dev, "Uplink: PDR(%u) didn't has a OHR information "
        "(which routed to the gtp interface and matches a PDR)\n", pdr->id);

out:
    return rt;
}

static int gtp5g_fwd_skb_encap(struct sk_buff *skb, struct net_device *dev,
    unsigned int hdrlen, struct pdr *pdr, struct far *far, u64 *txVol)
{
    struct forwarding_parameter *fwd_param = rcu_dereference(far->fwd_param);
    struct outer_header_creation *hdr_creation;
    struct forwarding_policy *fwd_policy;
    struct gtpv1_hdr *gtp1 = (struct gtpv1_hdr *)(skb->data + sizeof(struct udphdr));
    struct iphdr *iph;
    struct udphdr *uh;
    struct pcpu_sw_netstats *stats;
    int ret;
    u64 volume, volume_mbqe = 0;

    TrafficPolicer* tp = NULL;
    Color color = Green;
    struct qer __rcu *qer_with_rate = NULL;
    
    if (!pdr) {
        GTP5G_ERR(dev, "PDR is NULL\n");
        return PKT_DROPPED;
    }

    if (gtp1->type == GTPV1_MSG_TYPE_TPDU)
        volume_mbqe = ip4_rm_header(skb, hdrlen);

    qer_with_rate = rcu_dereference(pdr->qer_with_rate);
    if (qer_with_rate != NULL){
        if (is_uplink(pdr)) {
            tp = qer_with_rate->ul_policer;
        } else if (is_downlink(pdr)) {
            tp = qer_with_rate->dl_policer;
        }
    }
    if (get_qos_enable() && tp != NULL) {
        color = policePacket(tp, volume_mbqe);
    }
    if (color == Red) {
        volume = 0;
    } else {
        volume = volume_mbqe;
    }

    if (fwd_param) {
        if ((fwd_policy = fwd_param->fwd_policy))
            skb->mark = fwd_policy->mark;

        if ((hdr_creation = fwd_param->hdr_creation)) {
            // Just modify the teid and packet dest ip
            gtp1->tid = hdr_creation->teid;

            skb_push(skb, 20); // L3 Header Length
            iph = ip_hdr(skb);

            if (!pdr->pdi->f_teid) {
                GTP5G_ERR(dev, "Failed to hdr removal + creation "
                    "due to pdr->pdi->f_teid not exist\n");
                return PKT_DROPPED;
            }

            iph->saddr = pdr->pdi->f_teid->gtpu_addr_ipv4.s_addr;
            iph->daddr = hdr_creation->peer_addr_ipv4.s_addr;
            if (hdr_creation->tosTc) {
                iph->tos = hdr_creation->tosTc;
            }
            iph->check = 0;

            uh = udp_hdr(skb);
            uh->check = 0;

            // Reset checksum state after modifying headers
            skb->ip_summed = CHECKSUM_NONE;
            skb->csum_valid = 0;

            if (pdr->urr_num != 0) {
                ret = update_urr_counter_and_send_report(pdr, far, volume, volume_mbqe);
                if (ret < 0) {
                    if (ret == DONT_SEND_UL_PACKET) {
                        GTP5G_INF(pdr->dev, "Should not foward the first uplink packet");
                        return PKT_DROPPED;
                    } else {
                        GTP5G_ERR(pdr->dev, "Fail to send Usage Report");
                    }
                }
            }

            if (color == Red) {
                GTP5G_TRC(pdr->dev, "Drop red packet");
                return PKT_DROPPED;
            }
            /* Record the on-the-wire length while we still own the skb;
             * ip_xmit() consumes it.
             * */
            *txVol = skb->len;
            if (ip_xmit(skb, pdr->sk, dev) < 0) {
                GTP5G_ERR(dev, "Failed to transmit skb through ip_xmit\n");
                *txVol = 0;
                return PKT_DROPPED;
            }

            return PKT_FORWARDED;
        }
    }

    if (gtp1->type != GTPV1_MSG_TYPE_TPDU) {
        GTP5G_WAR(dev, "Uplink: GTPv1 msg type is not TPDU\n");
        return PKT_DROPPED;
    }

    // Get rid of the GTP-U + UDP headers.
    if (iptunnel_pull_header(skb,
            hdrlen, 
            skb->protocol,
            !net_eq(sock_net(pdr->sk), 
            dev_net(dev)))) {
        GTP5G_ERR(dev, "Failed to pull GTP-U and UDP headers\n");
        return PKT_DROPPED;
    }

    /* Now that the UDP and the GTP header have been removed, set up the
     * new network header. This is required by the upper layer to
     * calculate the transport header.
     * */
    skb_reset_network_header(skb);

    skb->dev = dev;

    stats = this_cpu_ptr(skb->dev->tstats);
    u64_stats_update_begin(&stats->syncp);
#if LINUX_VERSION_CODE >= KERNEL_VERSION(6, 0, 0)
    u64_stats_inc(&stats->rx_packets);
    u64_stats_add(&stats->rx_bytes, skb->len);
#else
    stats->rx_packets++;
    stats->rx_bytes += skb->len;
#endif
    u64_stats_update_end(&stats->syncp);

    pdr->ul_pkt_cnt++;
    pdr->ul_byte_cnt += skb->len; /* length without GTP header */
    GTP5G_INF(NULL, "PDR (%u) UL_PKT_CNT (%llu) UL_BYTE_CNT (%llu)", pdr->id, pdr->ul_pkt_cnt, pdr->ul_byte_cnt);    
 
    if (pdr->urr_num != 0) {
        if (update_urr_counter_and_send_report(pdr, far, volume, volume_mbqe) < 0)
            GTP5G_ERR(pdr->dev, "Fail to send Usage Report");
    }
    
    if (color == Red) {
        GTP5G_TRC(pdr->dev, "Drop red packet");
        return PKT_DROPPED;
    }
    /* Record the length while we still own the skb; netif_rx() consumes it. */
    *txVol = skb->len;
    ret = netif_rx(skb);
    if (ret != NET_RX_SUCCESS) {
        GTP5G_ERR(dev, "Uplink: Packet got dropped\n");
        *txVol = 0;
        return PKT_DROPPED_AND_FREED;
    }

    return PKT_FORWARDED;
}

static int gtp5g_drop_skb_ipv4(struct sk_buff *skb, struct net_device *dev,
    struct pdr *pdr)
{
    ++pdr->dl_drop_cnt;
    GTP5G_INF(NULL, "PDR (%u) DL_DROP_CNT (%llu)", pdr->id, pdr->dl_drop_cnt);
    return PKT_DROPPED;
}

static struct rtable *find_ip4_route(struct flowi4 *fl4,
    const struct sock *sk,
    __be32 daddr, __be32 saddr)
{
    memset(fl4, 0, sizeof(*fl4));
    fl4->flowi4_oif = sk->sk_bound_dev_if;
    fl4->daddr = daddr;
    fl4->saddr = saddr;
#if LINUX_VERSION_CODE >= KERNEL_VERSION(6, 18, 0)
    fl4->flowi4_dscp = inet_sk_dscp(inet_sk(sk));
    fl4->flowi4_scope = ip_sock_rt_scope(sk);
#else
    fl4->flowi4_tos = RT_TOS(inet_sk(sk)->tos);
#endif
    fl4->flowi4_proto = sk->sk_protocol;
    return ip_route_output_key(sock_net(sk), fl4);
}

static int gtp5g_fwd_skb_ipv4(struct sk_buff *skb, 
    struct net_device *dev, struct gtp5g_pktinfo *pktinfo, 
    struct pdr *pdr, struct far *far)
{
    struct rtable *rt;
    struct flowi4 fl4;
    struct iphdr *iph = ip_hdr(skb);
    struct outer_header_creation *hdr_creation;
    u64 volume, volume_mbqe = 0;
    struct forwarding_parameter *fwd_param;
    u8 pdu_type = PDU_SESSION_INFO_TYPE0;

    TrafficPolicer* tp = NULL;
    Color color = Green;
    struct qer __rcu *qer_with_rate = NULL;
    
    if (!far) {
        GTP5G_ERR(dev, "Unknown RAN address\n");
        goto err;
    }

    fwd_param = rcu_dereference(far->fwd_param);
    if (!(fwd_param &&
        fwd_param->hdr_creation)) {
        GTP5G_ERR(dev, "Unknown RAN address\n");
        goto err;
    }

    hdr_creation = fwd_param->hdr_creation;
    rt = find_ip4_route(&fl4,
        pdr->sk,
        hdr_creation->peer_addr_ipv4.s_addr,
        pdr->role_addr_ipv4.s_addr);
    if (IS_ERR(rt))
        goto err;

    if (is_uplink(pdr)) {
        pdu_type = PDU_SESSION_INFO_TYPE1;
    }

    gtp5g_set_pktinfo_ipv4(pktinfo,
            pdr->sk, 
            iph, 
            hdr_creation,
            pdr->qfi,
            pdu_type,
            far->seq_number,
            rt, 
            &fl4, 
            dev);

    far->seq_number++;
    pdr->dl_pkt_cnt++;
    pdr->dl_byte_cnt += skb->len;
    GTP5G_INF(NULL, "PDR (%u) DL_PKT_CNT (%llu) DL_BYTE_CNT (%llu)", pdr->id, pdr->dl_pkt_cnt, pdr->dl_byte_cnt);

    volume_mbqe = ip4_rm_header(skb, 0);

    qer_with_rate = rcu_dereference(pdr->qer_with_rate);
    if (qer_with_rate != NULL){
        if (is_uplink(pdr)) {
            tp = qer_with_rate->ul_policer;
        } else if (is_downlink(pdr)) {
            tp = qer_with_rate->dl_policer;
        }
    }
    if (get_qos_enable() && tp != NULL) {
        color = policePacket(tp, volume_mbqe);
    }
    if (color == Red) {
        volume = 0;
    } else {
        volume = volume_mbqe;
    }

    gtp5g_push_header(skb, pktinfo);

    if (pdr->urr_num != 0) {
        if (update_urr_counter_and_send_report(pdr, far, volume, volume_mbqe) < 0)
            GTP5G_ERR(pdr->dev, "Fail to send Usage Report");
    }
    if (color == Red) {
        GTP5G_TRC(pdr->dev, "Drop red packet");
        return PKT_DROPPED;
    }
    return PKT_FORWARDED;
err:
    return -EBADMSG;
}

static int gtp5g_buf_skb_ipv4(struct sk_buff *skb, struct net_device *dev,
    struct pdr *pdr, struct far *far)
{
    // Increment the sequence number for the packet
    far->seq_number++;
    if (pdr_addr_is_netlink(pdr)) {
        if (netlink_send(pdr, far, skb, dev_net(dev), NULL, 0) < 0) {
            GTP5G_ERR(dev, "Failed to send skb to netlink socket PDR(%u)", pdr->id);
            ++pdr->dl_drop_cnt;
        }
    } else {
        // TODO: handle nonlinear part
        if (unix_sock_send(pdr, far, skb->data, skb_headlen(skb), 0) < 0) {
            GTP5G_ERR(dev, "Failed to send skb to unix domain socket PDR(%u)", pdr->id);
            ++pdr->dl_drop_cnt;
        }
    }

    dev_kfree_skb(skb);
    return PKT_TO_APP;
}

int gtp5g_handle_skb_ipv4(struct sk_buff *skb, struct net_device *dev,
    struct gtp5g_pktinfo *pktinfo)
{
    struct gtp5g_dev *gtp = netdev_priv(dev);
    struct pdr *pdr;
    struct far *far;
    //struct gtp5g_qer *qer;
    struct iphdr *iph;
    struct qer __rcu *qer_with_rate = NULL;
    u32 mark;

    /* Read the IP destination address and resolve the PDR.
     * Prepend PDR header with TEI/TID from PDR.
     */
    iph = ip_hdr(skb);

    mark = get_skb_routing_mark(skb);
    if (is_framed_route_mark(mark)) {
        /* Try to find PDR by framed route using prefix length from mark */
        pdr = pdr_find_by_framed_route(gtp, skb, 0, iph->daddr,
                                       get_framed_route_prefix(mark));
    } else {
        // Note: The role is used here to get the pdr hashtable key - ueIP.
        // It will improve pdr lookup speed.
        if (gtp->role == GTP5G_ROLE_UPF) {
            pdr = pdr_find_by_ipv4(gtp, skb, 0, iph->daddr);
        } else {
            pdr = pdr_find_by_ipv4(gtp, skb, 0, iph->saddr);
        }
    }

    if (!pdr) {
        GTP5G_INF(dev, "no PDR found for %pI4, skip\n", &iph->daddr);
        return -ENOENT;
    }

    /* TODO: QoS rule have to apply before apply FAR 
     * */
    //qer = rcu_dereference(pdr->qer);
    //if (qer) {
    //    GTP5G_ERR(dev, "%s:%d QER Rule found, id(%#x) qfi(%#x) TODO\n", 
    //            __func__, __LINE__, qer->id, qer->qfi);
    //}

    qer_with_rate = rcu_dereference(pdr->qer_with_rate);
    far = rcu_dereference(pdr->far);
    if (far) {
        // One and only one of the DROP, FORW and BUFF flags shall be set to 1.
        // The NOCP flag may only be set if the BUFF flag is set.
        // The DUPL flag may be set with any of the DROP, FORW, BUFF and NOCP flags.
        switch (far->action & FAR_ACTION_MASK) {
        case FAR_ACTION_DROP:
            return gtp5g_drop_skb_ipv4(skb, dev, pdr);
        case FAR_ACTION_FORW:
            if (pdr->ul_dl_gate & QER_DL_GATE_CLOSE) {
                GTP5G_TRC(pdr->dev, "QER DL gate is closed, drop the packet");
                return PKT_DROPPED;
            }
            return gtp5g_fwd_skb_ipv4(skb, dev, pktinfo, pdr, far);
        case FAR_ACTION_BUFF:
            return gtp5g_buf_skb_ipv4(skb, dev, pdr, far);
        default:
            GTP5G_ERR(dev, "Unspec apply action(%u) in FAR(%u) and related to PDR(%u)",
                far->action, far->id, pdr->id);
        }
    }

    return -ENOENT;
}
