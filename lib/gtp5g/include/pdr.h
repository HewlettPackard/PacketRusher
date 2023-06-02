#ifndef __PDR_H__
#define __PDR_H__

#include <linux/kernel.h>
#include <linux/net.h>
#include <linux/rculist.h>
#include <linux/range.h>
#include <linux/skbuff.h>
#include <linux/un.h>
#include <net/ip.h>

#include "urr.h"

#define SEID_U32ID_HEX_STR_LEN 24

struct local_f_teid {
    u32 teid;
    struct in_addr gtpu_addr_ipv4;
};

struct ip_filter_rule {
    uint8_t action;
    uint8_t direction;
    uint8_t proto;
    struct in_addr src;
    struct in_addr smask;
    struct in_addr dest;
    struct in_addr dmask;
    int sport_num;
    struct range *sport;
    int dport_num;
    struct range *dport;
};

struct sdf_filter {
    struct ip_filter_rule *rule;
    uint16_t *tos_traffic_class;
    u32 *security_param_idx;
    u32 *flow_label;
    u32 *bi_id;
};

struct pdi {
    struct in_addr *ue_addr_ipv4;
    struct local_f_teid *f_teid;
    struct sdf_filter *sdf;
};

#define QER_ID_SIZE sizeof(u32)
#define URR_ID_SIZE sizeof(u32)

struct pdr_node {
    struct hlist_node hlist;
    struct pdr *pdr;
};

struct pdr {
    struct hlist_node hlist_id;
    struct hlist_node hlist_i_teid;
    struct hlist_node hlist_addr;

    u64 seid;
    u16 id;
    u32 precedence;
    u8 *outer_header_removal;
    struct pdi *pdi;
    u32 *far_id;
    struct far __rcu *far;
    u32 *qer_ids; 
    u32 qer_num;
    u8  qfi;
    u32 *urr_ids;
    u32 urr_num;
    struct urr *urr;
    
    /* deprecated: AF_UNIX socket for buffer */
    struct sockaddr_un addr_unix;
    struct socket *sock_for_buf;

    u16 af;
    struct in_addr role_addr_ipv4;
    struct sock *sk;
    struct net_device *dev;
    struct rcu_head rcu_head;

    /* Drop Counter */
    u64 ul_drop_cnt;
    u64 dl_drop_cnt;

    /* Packet Statistics */
    u64                     ul_pkt_cnt;
    u64                     dl_pkt_cnt;
    u64                     ul_byte_cnt;
    u64                     dl_byte_cnt;
};

extern void pdr_context_delete(struct pdr *);
extern struct pdr *find_pdr_by_id(struct gtp5g_dev *, u64, u16);
extern struct pdr *pdr_find_by_gtp1u(struct gtp5g_dev *, struct sk_buff *,
        unsigned int, u32, u8);
extern struct pdr *pdr_find_by_ipv4(struct gtp5g_dev *, struct sk_buff *,
        unsigned int, __be32);
extern int find_qer_id_in_pdr(struct pdr *, u32);
extern int find_urr_id_in_pdr(struct pdr *, u32);

extern void pdr_append(u64, u16, struct pdr *, struct gtp5g_dev *);
extern void pdr_update_hlist_table(struct pdr *, struct gtp5g_dev *);

extern void unix_sock_client_delete(struct pdr *);
extern int unix_sock_client_new(struct pdr *);
extern int unix_sock_client_update(struct pdr *, struct far *);

static inline bool pdr_addr_is_netlink(struct pdr *pdr)
{
    return (pdr->addr_unix.sun_path[0] == '/' && pdr->addr_unix.sun_path[1] == 0);
}

#endif // __PDR_H__
