/**
 * framed_route.c - Framed Route handling for GTP5G
 *
 * This module implements framed route functionality for the GTP5G kernel module,
 * providing support for routing packets to/from networks behind UEs.
 */

#include <linux/version.h>
#include <linux/types.h>
#include <linux/inet.h>
#include <linux/string.h>
#include <linux/slab.h>

#include "dev.h"
#include "pdr.h"
#include "framed_route.h"
#include "hash.h"
#include "log.h"
#include "genl_pdr.h"

/*
 * ============================================================================
 * Node Memory Management (Refactoring #1)
 * ============================================================================
 */

struct framed_route_node *framed_route_node_alloc(void)
{
    struct framed_route_node *node;

    node = kzalloc(sizeof(*node), GFP_ATOMIC);
    if (node)
        INIT_HLIST_NODE(&node->hlist);

    return node;
}

void framed_route_node_free(struct framed_route_node *node)
{
    if (node)
        kfree(node);
}

void framed_route_node_free_rcu(struct framed_route_node *node)
{
    if (node)
        kfree_rcu(node, rcu);
}

/*
 * ============================================================================
 * CIDR Parsing
 * ============================================================================
 */

int parse_framed_route_cidr(const char *route_str, __be32 *network_addr,
                            __be32 *netmask)
{
    const char *slash_pos;
    const char *end;
    int prefix_len;
    u8 ip[4];
    u32 mask;
    int ret;
    int len;

    if (!route_str || !network_addr || !netmask) {
        GTP5G_ERR(NULL, "Invalid arguments to parse_framed_route_cidr\n");
        return -EINVAL;
    }

    /* Find the slash separator */
    slash_pos = strchr(route_str, '/');
    if (!slash_pos) {
        GTP5G_ERR(NULL, "Invalid framed route format (missing '/'): %s\n",
                  route_str);
        return -EINVAL;
    }

    len = slash_pos - route_str;
    if (len <= 0 || len > 15) {  /* Max IPv4 string: "255.255.255.255" */
        GTP5G_ERR(NULL, "Invalid IP address length in framed route: %s\n",
                  route_str);
        return -EINVAL;
    }

    /* Parse IP address using kernel's in4_pton */
    if (!in4_pton(route_str, len, ip, '/', &end) || end != slash_pos) {
        GTP5G_ERR(NULL, "Failed to parse IP address in framed route: %s\n",
                  route_str);
        return -EINVAL;
    }

    /* Parse prefix length */
    ret = kstrtoint(slash_pos + 1, 10, &prefix_len);
    if (ret < 0 || prefix_len < 0 || prefix_len > 32) {
        GTP5G_ERR(NULL, "Invalid prefix length in framed route: %s\n",
                  route_str);
        return -EINVAL;
    }

    /* Build network address in network byte order */
    memcpy(network_addr, ip, sizeof(*network_addr));

    /* Calculate netmask */
    if (prefix_len == 0)
        mask = 0;
    else
        mask = htonl(0xFFFFFFFF << (32 - prefix_len));
    *netmask = mask;

    /* Apply mask to get network address */
    *network_addr &= *netmask;

    return 0;
}

/*
 * ============================================================================
 * Hash Table Operations (Refactoring #4)
 * ============================================================================
 */

void framed_route_hash_add(struct gtp5g_dev *gtp,
                           struct framed_route_node *node,
                           u32 precedence)
{
    struct hlist_head *head;
    struct framed_route_node *iter;
    struct framed_route_node *last = NULL;

    if (!gtp || !gtp->framed_route_hash || !node)
        return;

    /* Remove from old position if already in a hash */
    if (!hlist_unhashed(&node->hlist))
        hlist_del_rcu(&node->hlist);

    head = &gtp->framed_route_hash[ipv4_hashfn(node->network_addr) %
                                   gtp->hash_size];

    /* Find insertion point based on precedence (lower = higher priority) */
    hlist_for_each_entry_rcu(iter, head, hlist) {
        if (!iter->pdr)
            continue;
        if (precedence > iter->pdr->precedence)
            last = iter;
        else
            break;
    }

    if (!last)
        hlist_add_head_rcu(&node->hlist, head);
    else
        hlist_add_behind_rcu(&node->hlist, &last->hlist);

    GTP5G_TRC(NULL, "Added framed route %pI4/%u to hash\n",
              &node->network_addr, netmask_to_prefix(node->netmask));
}

void framed_route_hash_del(struct framed_route_node *node)
{
    if (node && !hlist_unhashed(&node->hlist))
        hlist_del_rcu(&node->hlist);
}

void framed_route_cleanup_pdi(struct pdi *pdi, bool use_rcu)
{
    int j;

    if (!pdi || !pdi->framed_route_nodes)
        return;

    for (j = 0; j < pdi->framed_route_num; j++) {
        struct framed_route_node *node = pdi->framed_route_nodes[j];

        if (!node)
            continue;

        /*
         * Only remove from hash when use_rcu=true (called from genl_pdr.c).
         * When use_rcu=false, we're in RCU callback (pdr_context_free) and
         * nodes were already removed from hash in pdr_context_delete.
         */
        if (use_rcu)
            framed_route_hash_del(node);

        if (use_rcu)
            framed_route_node_free_rcu(node);
        else
            framed_route_node_free(node);
    }

    kfree(pdi->framed_route_nodes);
    pdi->framed_route_nodes = NULL;
    pdi->framed_route_num = 0;
}

/*
 * ============================================================================
 * IP Matching with Framed Routes
 * ============================================================================
 */

bool ip_match_with_framed_routes(const __be32 ip_addr, const __be32 rule_addr,
                                 const __be32 rule_mask, const struct pdi *pdi)
{
    struct framed_route_node *node;
    int j;

    /* First try exact match with rule */
    if (ipv4_match(ip_addr, rule_addr, rule_mask))
        return true;

    /* If no exact match, check framed routes */
    if (!pdi || !pdi->framed_route_nodes || pdi->framed_route_num == 0)
        return false;

    for (j = 0; j < pdi->framed_route_num; j++) {
        node = pdi->framed_route_nodes[j];
        if (node && ipv4_match(ip_addr, node->network_addr, node->netmask)) {
            GTP5G_TRC(NULL, "IP %pI4 matched framed route %pI4/%u\n",
                      &ip_addr, &node->network_addr,
                      netmask_to_prefix(node->netmask));
            return true;
        }
    }

    return false;
}

