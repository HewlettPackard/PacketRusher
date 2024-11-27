#include <linux/version.h>
#include <linux/udp.h>
#include <linux/skbuff.h>
#include <linux/inetdevice.h>
#include <net/ip.h>
#include <net/icmp.h>
#include <net/udp_tunnel.h>
#include <net/route.h>

#include "gtp.h"
#include "far.h"
#include "qer.h"
#include "pktinfo.h"
#include "log.h"

u64 network_and_transport_header_len(struct sk_buff *skb) {
    u64 hdrlen;
    struct iphdr *iph;
    struct tcphdr *tcp;
    
    iph = (struct iphdr *)skb->data;
    hdrlen = iph->ihl * 4;

    switch (iph->protocol) {
        case IPPROTO_TCP:
            // tcp = (struct tcphdr *)(skb_transport_header(skb) + (iph->ihl << 2));
            skb->len -= iph->ihl * 4;
            skb->data += iph->ihl * 4;

            tcp =  (struct tcphdr *)skb->data;
            hdrlen += tcp->doff * 4;
            break;
        case IPPROTO_UDP:
            hdrlen +=  8; // udp header len = 8B
            break;
        default:
            break;
    }

    return hdrlen;
}

u64 ip4_rm_header(struct sk_buff *skb, unsigned int hdrlen) {
    struct sk_buff *skb_copy, tmp;
    u64 volume;

    // To make sure cacaluting the len of skb will not move the data & len value 
    // of the original skb
    tmp = *skb;
    skb_copy = &tmp;

    volume = skb->len;
    if (hdrlen > 0) {
        // packets with gtp header
        volume -= hdrlen;
        skb_copy->len -= hdrlen;
        skb_copy->data += hdrlen;
    }

    // packets without gtp header
    volume -= network_and_transport_header_len(skb_copy);
    return volume;
}

struct rtable *ip4_find_route(struct sk_buff *skb, struct iphdr *iph,
    struct sock *sk, struct net_device *gtp_dev, 
    __be32 saddr, __be32 daddr, struct flowi4 *fl4)
{
    struct rtable *rt;
    __be16 df;
    int mtu;

    memset(fl4, 0, sizeof(*fl4));
    fl4->flowi4_oif = sk->sk_bound_dev_if;
    fl4->daddr = daddr;
    fl4->saddr = (saddr ? saddr : inet_sk(sk)->inet_saddr);
    fl4->flowi4_tos = RT_TOS(inet_sk(sk)->tos) | sock_flag(sk, SOCK_LOCALROUTE);
    fl4->flowi4_proto = sk->sk_protocol;

    rt = ip_route_output_key(dev_net(gtp_dev), fl4);
    if (IS_ERR(rt)) {
        GTP5G_ERR(gtp_dev, "no route to %pI4\n", &iph->daddr);
        gtp_dev->stats.tx_carrier_errors++;
        goto err;
    }

    if (rt->dst.dev == gtp_dev) {
        GTP5G_ERR(gtp_dev, "circular route to %pI4\n", &iph->daddr);
        gtp_dev->stats.collisions++;
        goto err_rt;
    }

    skb_dst_drop(skb);

    /* This is similar to tnl_update_pmtu(). */
    df = iph->frag_off;
    if (df) {
        mtu = dst_mtu(&rt->dst) - gtp_dev->hard_header_len -
            sizeof(struct iphdr) - sizeof(struct udphdr);
        // GTPv1
        mtu -= sizeof(struct gtpv1_hdr);
    }
    else {
        mtu = dst_mtu(&rt->dst);
    }

#if LINUX_VERSION_CODE >= KERNEL_VERSION(5, 4, 8) || defined(RHEL8)
       rt->dst.ops->update_pmtu(&rt->dst, NULL, skb, mtu, false);
#else
       rt->dst.ops->update_pmtu(&rt->dst, NULL, skb, mtu);
#endif

    if (!skb_is_gso(skb) && (iph->frag_off & htons(IP_DF)) &&
        mtu < ntohs(iph->tot_len)) {
        GTP5G_ERR(gtp_dev, "packet too big, fragmentation needed\n");
        memset(IPCB(skb), 0, sizeof(*IPCB(skb)));
        icmp_send(skb, ICMP_DEST_UNREACH, ICMP_FRAG_NEEDED,
              htonl(mtu));
        goto err_rt;
    }

    return rt;
err_rt:
    ip_rt_put(rt);
err:
    return ERR_PTR(-ENOENT);
}

struct rtable *ip4_find_route_simple(struct sk_buff *skb,
    struct sock *sk, struct net_device *gtp_dev, 
    __be32 saddr, __be32 daddr, struct flowi4 *fl4)
{
    struct rtable *rt;

    memset(fl4, 0, sizeof(*fl4));
    fl4->flowi4_oif = sk->sk_bound_dev_if;
    fl4->daddr = daddr;
    fl4->saddr = (saddr ? saddr : inet_sk(sk)->inet_saddr);
    fl4->flowi4_tos = RT_TOS(inet_sk(sk)->tos) | sock_flag(sk, SOCK_LOCALROUTE);
    fl4->flowi4_proto = sk->sk_protocol;

    rt = ip_route_output_key(dev_net(gtp_dev), fl4);
    if (IS_ERR(rt)) {
        GTP5G_ERR(gtp_dev, "no route from %#x to %#x\n", saddr, daddr);
        gtp_dev->stats.tx_carrier_errors++;
        goto err;
    }

    if (rt->dst.dev == gtp_dev) {
        GTP5G_ERR(gtp_dev, "Packet colllisions from %#x to %#x\n", 
            saddr, daddr);
        gtp_dev->stats.collisions++;
        goto err_rt;
    }

    skb_dst_drop(skb);

    return rt;

err_rt:
    ip_rt_put(rt);
err:
    return ERR_PTR(-ENOENT);
}

int ip_xmit(struct sk_buff *skb, struct sock *sk, struct net_device *gtp_dev) 
{
    struct iphdr *iph = ip_hdr(skb);
    struct flowi4 fl4;
    struct rtable *rt;
    __be32 src;

    rt = ip4_find_route_simple(skb, sk, gtp_dev, 0, iph->daddr, &fl4);
    if (IS_ERR(rt)) {
        GTP5G_ERR(gtp_dev, "Failed to find route\n");
        return -EBADMSG;
    }

    skb_dst_set(skb, &rt->dst);
    /*
        fill in correct source address of the outgoing interface.
        Support multiple IP address configured on outgoing interface.
     */
    src = inet_select_addr(rt->dst.dev,
                    rt_nexthop(rt, iph->daddr),
                    RT_SCOPE_UNIVERSE);
    if (src != 0) {
        iph->saddr = src;
    }

    if (ip_local_out(dev_net(gtp_dev), sk, skb) < 0) {
        GTP5G_ERR(gtp_dev, "Failed to send skb to ip layer\n");
        return -1;
    }
    return 0;
}

void gtp5g_fwd_emark_skb_ipv4(struct sk_buff *skb,
    struct net_device *dev, struct gtp5g_emark_pktinfo *epkt_info) 
{
    struct rtable *rt;
    struct flowi4 fl4;
    struct gtpv1_hdr *gtp1;

    /* Reset all headers */
    skb_reset_transport_header(skb);
    skb_reset_network_header(skb);
    skb_reset_mac_header(skb);

    /* Fill GTP-U Header */
    gtp1 = skb_push(skb, sizeof(*gtp1));
    gtp1->flags = GTPV1; /* v1, GTP-non-prime. */
    gtp1->type = GTPV1_MSG_TYPE_EMARK;
    gtp1->tid = epkt_info->teid;

    rt = ip4_find_route_simple(skb, epkt_info->sk, dev, 
        epkt_info->role_addr /* Src Addr */ ,
        epkt_info->peer_addr /* Dst Addr*/, 
        &fl4);
    if (IS_ERR(rt)) {
        GTP5G_ERR(dev, "Failed to send GTP-U end-marker due to routing\n");
        dev_kfree_skb(skb);
        return;
    }

    udp_tunnel_xmit_skb(rt, 
        epkt_info->sk, 
        skb,
        fl4.saddr, 
        fl4.daddr,
        0,
        ip4_dst_hoplimit(&rt->dst),
        0,
        epkt_info->gtph_port, 
        epkt_info->gtph_port,
        true, 
        true);
}

void gtp5g_xmit_skb_ipv4(struct sk_buff *skb, struct gtp5g_pktinfo *pktinfo)
{
    u8 tos = 0;
    if (pktinfo->hdr_creation == NULL) {
        tos = pktinfo->iph->tos;
    } else {
        tos = pktinfo->hdr_creation->tosTc;
    }
    udp_tunnel_xmit_skb(pktinfo->rt, 
        pktinfo->sk,
        skb,
        pktinfo->fl4.saddr,
        pktinfo->fl4.daddr,
        tos,
        ip4_dst_hoplimit(&pktinfo->rt->dst),
        0,
        pktinfo->gtph_port, 
        pktinfo->gtph_port,
        true, 
        true);
}

inline void gtp5g_set_pktinfo_ipv4(struct gtp5g_pktinfo *pktinfo,
    struct sock *sk, struct iphdr *iph, struct outer_header_creation *hdr_creation,
    u8 qfi, u8 pdu_type, u16 seq_number, struct rtable *rt, struct flowi4 *fl4,
    struct net_device *dev)
{
    pktinfo->sk = sk;
    pktinfo->iph = iph;
    pktinfo->hdr_creation = hdr_creation;
    pktinfo->qfi = qfi;
    pktinfo->pdu_type = pdu_type;
    pktinfo->seq_number = seq_number;
    pktinfo->rt = rt;
    pktinfo->fl4 = *fl4;
    pktinfo->dev = dev;
}

void gtp5g_push_header(struct sk_buff *skb, struct gtp5g_pktinfo *pktinfo)
{
    int payload_len = skb->len;
    struct gtpv1_hdr *gtp1;
    gtpv1_hdr_opt_t *gtp1opt;
    ext_pdu_sess_ctr_t *ext_pdu_sess;
    u16 seq_number = 0;
    u8 next_ehdr_type = 0;

    int ext_flag = 0;
    int opt_flag = 0;
    int seq_flag = get_seq_enable();

    GTP5G_TRC(NULL, "SKBLen(%u) GTP-U V1(%zu) Opt(%zu) PDU(%zu)\n",
        payload_len, sizeof(*gtp1), sizeof(*gtp1opt), sizeof(*ext_pdu_sess));

    pktinfo->gtph_port = pktinfo->hdr_creation->port;

    /* Suppport for extension header, sequence number and N-PDU.
     * Update the length field if any of them is available.
     */
    if (pktinfo->qfi > 0) {
        ext_flag = 1; 

        /* Push PDU Session container information */
        ext_pdu_sess = skb_push(skb, sizeof(*ext_pdu_sess));
        /* Multiple of 4 (TODO include PPI) */
        ext_pdu_sess->length = 1;

        if (pktinfo->pdu_type == PDU_SESSION_INFO_TYPE1) { // UL
            ext_pdu_sess->pdu_sess_ctr.type_spare = PDU_SESSION_INFO_TYPE1;
            ext_pdu_sess->pdu_sess_ctr.u.ul.spare_qfi = pktinfo->qfi;
        } else { // DL
            ext_pdu_sess->pdu_sess_ctr.type_spare = PDU_SESSION_INFO_TYPE0;
            ext_pdu_sess->pdu_sess_ctr.u.dl.ppp_rqi_qfi = pktinfo->qfi;
        }

        //TODO: PPI
        ext_pdu_sess->next_ehdr_type = 0; /* No more extension Header */
        
        opt_flag = 1;
        next_ehdr_type = 0x85; /* PDU Session Container */
    }

    if (seq_flag){
        opt_flag = 1;
        seq_number = htons(pktinfo->seq_number);
    }

    if (opt_flag) {
        /* Push optional header information */
        gtp1opt = skb_push(skb, sizeof(*gtp1opt));
        gtp1opt->seq_number = seq_number;
        gtp1opt->NPDU = 0;
        gtp1opt->next_ehdr_type = next_ehdr_type;
        // Increment the GTP-U payload length by size of optional headers length
        payload_len += (sizeof(*gtp1opt) + sizeof(*ext_pdu_sess));
    }

    /* Bits 8  7  6  5  4  3  2  1
     *    +--+--+--+--+--+--+--+--+
     *    |version |PT| 0| E| S|PN|
     *    +--+--+--+--+--+--+--+--+
     *      0  0  1  1  0  0  0  0
     */
    gtp1 = skb_push(skb, sizeof(*gtp1));
    gtp1->flags = GTPV1; /* v1, GTP-non-prime. */
    if (ext_flag) 
        gtp1->flags |= GTPV1_HDR_FLG_EXTHDR; /* v1, Extension header enabled */ 
    if (seq_flag)
        gtp1->flags |= GTPV1_HDR_FLG_SEQ;
    gtp1->type = GTPV1_MSG_TYPE_TPDU;
    gtp1->tid = pktinfo->hdr_creation->teid;
    gtp1->length = htons(payload_len);       /* Excluded the header length of gtpv1 */

    GTP5G_TRC(NULL, "QER Found GTP-U Flg(%u) GTPU-L(%u) SkbLen(%u)\n", 
        gtp1->flags, ntohs(gtp1->length), skb->len);
}
