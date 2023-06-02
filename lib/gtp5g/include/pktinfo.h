#ifndef __PKTINFO_H__
#define __PKTINFO_H__

#include <linux/skbuff.h>
#include <linux/net.h>

#include "qer.h"

struct gtp5g_pktinfo {
    struct sock                   *sk;
    struct iphdr                  *iph;
    struct flowi4                 fl4;
    struct rtable                 *rt;
    struct outer_header_creation  *hdr_creation;
    u8                            qfi;
    struct net_device             *dev;
    __be16                        gtph_port;
};

struct gtp5g_emark_pktinfo {
    u32 teid;
    u32 peer_addr;
    u32 local_addr;
    u32 role_addr;
    
    struct sock         *sk;
    struct flowi4       fl4;
    struct rtable       *rt;
    struct net_device   *dev;
    __be16              gtph_port;
};

extern u64 ip4_rm_header(struct sk_buff *skb, unsigned int hdrlen);
extern struct rtable *ip4_find_route(struct sk_buff *, struct iphdr *,
        struct sock *, struct net_device *,
        __be32, __be32, struct flowi4 *);
extern void gtp5g_fwd_emark_skb_ipv4(struct sk_buff *,
        struct net_device *, struct gtp5g_emark_pktinfo *);
extern int ip_xmit(struct sk_buff *, struct sock *, struct net_device *);
extern void gtp5g_xmit_skb_ipv4(struct sk_buff *, struct gtp5g_pktinfo *);

extern void gtp5g_set_pktinfo_ipv4(struct gtp5g_pktinfo *,
        struct sock *, struct iphdr *,
        struct outer_header_creation *,
        u8, struct rtable *, struct flowi4 *,
        struct net_device *);
extern void gtp5g_push_header(struct sk_buff *, struct gtp5g_pktinfo *);

#endif // __PKTINFO_H__
