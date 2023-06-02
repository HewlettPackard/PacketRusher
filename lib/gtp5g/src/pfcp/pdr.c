#include <linux/version.h>

#include "dev.h"
#include "link.h"
#include "pdr.h"
#include "far.h"
#include "gtp.h"
#include "genl_pdr.h"
#include "genl_far.h"
#include "seid.h"
#include "hash.h"
#include "genl.h"
#include "log.h"
#include <linux/types.h>


static void seid_pdr_id_to_hex_str(u64 seid_int, u16 pdr_id, char *buff)
{
    seid_and_u32id_to_hex_str(seid_int, (u32)(pdr_id), buff);
}

static void pdr_context_free(struct rcu_head *head)
{
    struct pdr *pdr = container_of(head, struct pdr, rcu_head);
    struct pdi *pdi;
    struct sdf_filter *sdf;

    if (!pdr)
        return;

    sock_put(pdr->sk);

    if (pdr->outer_header_removal)
        kfree(pdr->outer_header_removal);

    pdi = pdr->pdi;
    if (pdi) {
        if (pdi->ue_addr_ipv4)
            kfree(pdi->ue_addr_ipv4);
        if (pdi->f_teid)
            kfree(pdi->f_teid);
        if (pdr->far_id)
            kfree(pdr->far_id);
        if (pdr->qer_ids)
            kfree(pdr->qer_ids);
        if (pdr->urr_ids)
            kfree(pdr->urr_ids);

        sdf = pdi->sdf;
        if (sdf) {
            if (sdf->rule) {
                if (sdf->rule->sport)
                    kfree(sdf->rule->sport);
                if (sdf->rule->dport)
                    kfree(sdf->rule->dport);
                kfree(sdf->rule);
            }
            if (sdf->tos_traffic_class)
                kfree(sdf->tos_traffic_class);
            if (sdf->security_param_idx)
                kfree(sdf->security_param_idx);
            if (sdf->flow_label)
                kfree(sdf->flow_label);
            if (sdf->bi_id)
                kfree(sdf->bi_id);

            kfree(sdf);
        }
        kfree(pdi);
    }

    unix_sock_client_delete(pdr);
    kfree(pdr);
}

void pdr_context_delete(struct pdr *pdr)
{
    if (!pdr)
        return;

    if (!hlist_unhashed(&pdr->hlist_id))
        hlist_del_rcu(&pdr->hlist_id);

    if (!hlist_unhashed(&pdr->hlist_i_teid))
        hlist_del_rcu(&pdr->hlist_i_teid);

    if (!hlist_unhashed(&pdr->hlist_addr))
        hlist_del_rcu(&pdr->hlist_addr);

    call_rcu(&pdr->rcu_head, pdr_context_free);
}

// Delete the AF_UNIX client
void unix_sock_client_delete(struct pdr *pdr)
{
    if (!pdr || pdr_addr_is_netlink(pdr))
        return;

    if (pdr->sock_for_buf)
        sock_release(pdr->sock_for_buf);

    pdr->sock_for_buf = NULL;
}

// Create a AF_UNIX client by specific name sent from user space
int unix_sock_client_new(struct pdr *pdr)
{
    struct socket **psock = &pdr->sock_for_buf;
    struct sockaddr_un *addr = &pdr->addr_unix;
    int err;

    if (strlen(addr->sun_path) == 0) {
        return -EINVAL;
    }

    if (pdr_addr_is_netlink(pdr)) {
        return 0;
    }

    err = sock_create(AF_UNIX, SOCK_DGRAM, 0, psock);
    if (err) {
        return err;
    }

    err = (*psock)->ops->connect(*psock, (struct sockaddr *)addr,
            sizeof(addr->sun_family) + strlen(addr->sun_path), 0);
    if (err) {
        unix_sock_client_delete(pdr);
        return err;
    }

    return 0;
}

// Handle PDR/FAR changed and affect buffering
int unix_sock_client_update(struct pdr *pdr, struct far *far)
{
    if (!pdr || pdr_addr_is_netlink(pdr))
        return 0;

    unix_sock_client_delete(pdr);

    if ((far && (far->action & FAR_ACTION_BUFF)) || pdr->urr_num > 0)
        return unix_sock_client_new(pdr);

    return 0;
}

struct pdr *find_pdr_by_id(struct gtp5g_dev *gtp, u64 seid, u16 pdr_id)
{
    struct hlist_head *head;
    struct pdr *pdr;
    char seid_pdr_id[SEID_U32ID_HEX_STR_LEN] = {0};

    seid_pdr_id_to_hex_str(seid, pdr_id, seid_pdr_id);
    head = &gtp->pdr_id_hash[str_hashfn(seid_pdr_id) % gtp->hash_size];
    hlist_for_each_entry_rcu(pdr, head, hlist_id) {
        if (pdr->seid == seid && pdr->id == pdr_id)
            return pdr;
    }

    return NULL;
}

static int ipv4_match(__be32 target_addr, __be32 ifa_addr, __be32 ifa_mask)
{
    return !((target_addr ^ ifa_addr) & ifa_mask);
}

static bool ports_match(struct range *match_list, int list_len, __be16 port)
{
    int i;

    if (!list_len)
        return true;

    for (i = 0; i < list_len; i++) {
        if (match_list[i].start <= port && match_list[i].end >= port)
            return true;
    }
    return false;
}

static int sdf_filter_match(struct sdf_filter *sdf, struct sk_buff *skb,
        unsigned int hdrlen, u8 direction)
{
    #define IP_PROTO_RESERVED 0xff
    struct iphdr *iph;
    struct ip_filter_rule *rule;
    const __be16 *pptr;
    __be16 _ports[2];

    if (!sdf)
        return 1;

    if (!pskb_may_pull(skb, hdrlen + sizeof(struct iphdr)))
        goto mismatch;

    iph = (struct iphdr *)(skb->data + hdrlen);

    if (sdf->rule) {
        rule = sdf->rule;
        if (rule->direction != direction)
            goto mismatch;

        if (rule->proto != IP_PROTO_RESERVED && rule->proto != iph->protocol)
            goto mismatch;

        if (!ipv4_match(iph->saddr, rule->src.s_addr, rule->smask.s_addr))
            goto mismatch;

        if (!ipv4_match(iph->daddr, rule->dest.s_addr, rule->dmask.s_addr))
            goto mismatch;

        if (rule->sport_num + rule->dport_num > 0) {
            if (!(pptr = skb_header_pointer(skb, hdrlen + sizeof(struct iphdr), sizeof(_ports), _ports)))
                goto mismatch;

            if (!ports_match(rule->sport, rule->sport_num, ntohs(pptr[0])))
                goto mismatch;

            if (!ports_match(rule->dport, rule->dport_num, ntohs(pptr[1])))
                goto mismatch;
        }
    }

/*
    if (sdf->tos_traffic_class)
        GTP5G_ERR(NULL, "TODO: SDF's ToS traffic class\n");

    if (sdf->security_param_idx)
        GTP5G_ERR(NULL, "TODO: SDF's Security parameter index\n");

    if (sdf->flow_label)
        GTP5G_ERR(NULL, "TODO: SDF's Flow label\n");

    if (sdf->bi_id)
        GTP5G_ERR(NULL, "TODO: SDF's SDF filter id\n");
*/

    return 1;
mismatch:
    return 0;
}

struct pdr *pdr_find_by_gtp1u(struct gtp5g_dev *gtp, struct sk_buff *skb,
        unsigned int hdrlen, u32 teid, u8 type)
{
#ifdef MATCH_IP
    struct iphdr *outer_iph;
#endif
    struct iphdr *iph;
    __be32 *target_addr = NULL;
    struct hlist_head *head;
    struct pdr *pdr;
    struct pdi *pdi;

    if (!gtp)
        return NULL;

    if (ntohs(skb->protocol) != ETH_P_IP)
        return NULL;

    if (type == GTPV1_MSG_TYPE_TPDU) {
        if (!pskb_may_pull(skb, hdrlen + sizeof(struct iphdr)))
            return NULL;
        iph = (struct iphdr *)(skb->data + hdrlen);
        target_addr = (gtp->role == GTP5G_ROLE_UPF ? &iph->saddr : &iph->daddr);
    }

    head = &gtp->i_teid_hash[u32_hashfn(teid) % gtp->hash_size];
    hlist_for_each_entry_rcu(pdr, head, hlist_i_teid) {
        pdi = pdr->pdi;
        if (!pdi)
            continue;

        // GTP-U packet must check teid
        if (!(pdi->f_teid && pdi->f_teid->teid == teid))
            continue;

        if (type != GTPV1_MSG_TYPE_TPDU)
            return pdr;

        // check outer IP dest addr to distinguish between N3 and N9 packet whil e act as i-upf
#ifdef MATCH_IP
    #if LINUX_VERSION_CODE >= KERNEL_VERSION(4, 0, 0)
            outer_iph = (struct iphdr *)(skb->head + skb->network_header);
            if (!(pdi->f_teid && pdi->f_teid->gtpu_addr_ipv4.s_addr == outer_iph->daddr))
                continue;
    #else
            outer_iph = (struct iphdr *)(skb->network_header);
            if (!(pdi->f_teid && pdi->f_teid->gtpu_addr_ipv4.s_addr == outer_iph->daddr))
                continue;
    #endif
#endif
        if (pdi->ue_addr_ipv4)
            if (!(pdr->af == AF_INET && target_addr && *target_addr == pdi->ue_addr_ipv4->s_addr))
                continue;

        if (pdi->sdf)
            if (!sdf_filter_match(pdi->sdf, skb, hdrlen, GTP5G_SDF_FILTER_OUT))
                continue;

        GTP5G_INF(NULL, "Match PDR ID:%d\n", pdr->id);

        return pdr;
    }

    return NULL;
}

struct pdr *pdr_find_by_ipv4(struct gtp5g_dev *gtp, struct sk_buff *skb,
        unsigned int hdrlen, __be32 addr)
{
    struct hlist_head *head;
    struct pdr *pdr;
    struct pdi *pdi;

    head = &gtp->addr_hash[ipv4_hashfn(addr) % gtp->hash_size];

    hlist_for_each_entry_rcu(pdr, head, hlist_addr) {
        pdi = pdr->pdi;

        // TODO: Move the value we check into first level
        if (!(pdr->af == AF_INET && pdi->ue_addr_ipv4->s_addr == addr))
            continue;

        if (pdi->sdf)
            if (!sdf_filter_match(pdi->sdf, skb, hdrlen, GTP5G_SDF_FILTER_OUT))
                continue;

        return pdr;
    }

    return NULL;
}

void pdr_append(u64 seid, u16 pdr_id, struct pdr *pdr, struct gtp5g_dev *gtp)
{
    u32 i;
    char seid_pdr_id_hexstr[SEID_U32ID_HEX_STR_LEN] = {0};

    seid_pdr_id_to_hex_str(seid, pdr_id, seid_pdr_id_hexstr);
    i = str_hashfn(seid_pdr_id_hexstr) % gtp->hash_size;
    hlist_add_head_rcu(&pdr->hlist_id, &gtp->pdr_id_hash[i]);
}

void pdr_update_hlist_table(struct pdr *pdr, struct gtp5g_dev *gtp)
{
    struct hlist_head *head;
    struct pdr *ppdr;
    struct pdr *last_ppdr;
    struct pdi *pdi;
    struct local_f_teid *f_teid;

    if (!hlist_unhashed(&pdr->hlist_i_teid))
        hlist_del_rcu(&pdr->hlist_i_teid);

    if (!hlist_unhashed(&pdr->hlist_addr))
        hlist_del_rcu(&pdr->hlist_addr);

    pdi = pdr->pdi;
    if (!pdi)
        return;

    f_teid = pdi->f_teid;
    if (f_teid) {
        last_ppdr = NULL;
        head = &gtp->i_teid_hash[u32_hashfn(f_teid->teid) % gtp->hash_size];
        hlist_for_each_entry_rcu(ppdr, head, hlist_i_teid) {
            if (pdr->precedence > ppdr->precedence)
                last_ppdr = ppdr;
            else
                break;
        }
        if (!last_ppdr)
            hlist_add_head_rcu(&pdr->hlist_i_teid, head);
        else
            hlist_add_behind_rcu(&pdr->hlist_i_teid, &last_ppdr->hlist_i_teid);
    } else if (pdi->ue_addr_ipv4) {
        last_ppdr = NULL;
        head = &gtp->addr_hash[u32_hashfn(pdi->ue_addr_ipv4->s_addr) % gtp->hash_size];
        hlist_for_each_entry_rcu(ppdr, head, hlist_addr) {
            if (pdr->precedence > ppdr->precedence)
                last_ppdr = ppdr;
            else
                break;
        }
        if (!last_ppdr)
            hlist_add_head_rcu(&pdr->hlist_addr, head);
        else
            hlist_add_behind_rcu(&pdr->hlist_addr, &last_ppdr->hlist_addr);
    }
}
