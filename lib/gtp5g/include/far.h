#ifndef __FAR_H__
#define __FAR_H__

#include <linux/kernel.h>
#include <linux/rculist.h>
#include <linux/net.h>
#include <net/ip.h>

#include "pktinfo.h"
#include "bar.h"
#include "pdr.h"

#define SEID_U32ID_HEX_STR_LEN 24

#define FAR_ACTION_UPSPEC 0x00
#define FAR_ACTION_DROP   0x01
#define FAR_ACTION_FORW   0x02
#define FAR_ACTION_BUFF   0x04
#define FAR_ACTION_MASK   0x07
#define FAR_ACTION_NOCP   0x08
#define FAR_ACTION_DUPL   0x10

#define FAR_ACTION_U8  1
#define FAR_ACTION_U16 2

struct outer_header_creation {
    u16 description;
    u32 teid;
    struct in_addr peer_addr_ipv4;
    u16 port;
};

struct forwarding_policy {
    int len;
    char identifier[0xff + 1];
    /* Exact value to handle forwarding policy */
    u32 mark;
};

struct forwarding_parameter {
    struct outer_header_creation *hdr_creation;
    struct forwarding_policy *fwd_policy;
};

struct far {
    struct hlist_node hlist_id;
    struct hlist_node hlist_related_bar;
    u64 seid;
    u32 id;
    u16 action;
    struct forwarding_parameter __rcu *fwd_param;
    u8 *bar_id;
    struct bar *bar;
    struct net_device *dev;
    struct rcu_head rcu_head;
};

extern void far_context_delete(struct far *);
extern struct far *find_far_by_id(struct gtp5g_dev *, u64, u32);
extern void far_update(struct far *, struct gtp5g_dev *, u8 *,
        struct gtp5g_emark_pktinfo *);
extern void far_append(u64, u32, struct far *, struct gtp5g_dev *);
extern int far_get_pdr_ids(u16 *, int, struct far *, struct gtp5g_dev *);
extern int far_set_pdr(struct pdr *, struct gtp5g_dev *);
extern void del_related_far_hash(struct gtp5g_dev *, struct pdr *);

#endif // __FAR_H__