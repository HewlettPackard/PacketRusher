#ifndef __FRAMED_ROUTE_H__
#define __FRAMED_ROUTE_H__

#include <linux/kernel.h>
#include <linux/types.h>
#include <linux/rculist.h>
#include <linux/rcupdate.h>
#include <linux/skbuff.h>
#include <linux/bitops.h>
#include <net/ip.h>

/* Forward declarations */
struct pdr;
struct gtp5g_dev;
struct pdi;

/**
 * struct framed_route_node - Represents a single framed route entry
 * @hlist: Hash list node for hash table linkage
 * @rcu: RCU head for safe memory reclamation
 * @pdr: Pointer to the associated PDR
 * @network_addr: Network address (after masking), in network byte order
 * @netmask: Netmask for the framed route, in network byte order
 */
struct framed_route_node {
    struct hlist_node hlist;
    struct rcu_head rcu;
    struct pdr *pdr;
    __be32 network_addr;
    __be32 netmask;
};

/*
 * ============================================================================
 * Inline Helper Functions
 * ============================================================================
 */

/**
 * netmask_to_prefix - Convert IPv4 netmask to CIDR prefix length
 * @netmask: Netmask in network byte order
 *
 * Return: CIDR prefix length (0-32)
 *
 * Example: 255.255.255.0 -> 24
 */
static inline u8 netmask_to_prefix(const __be32 netmask)
{
    return hweight32(ntohl(netmask));
}

/**
 * ipv4_match - Check if an IPv4 address matches a network/mask
 * @target_addr: Address to check (network byte order)
 * @network_addr: Network address (network byte order)
 * @netmask: Network mask (network byte order)
 *
 * Return: true if address matches, false otherwise
 */
static inline bool ipv4_match(const __be32 target_addr,
                              const __be32 network_addr,
                              const __be32 netmask)
{
    return !((target_addr ^ network_addr) & netmask);
}

/*
 * ============================================================================
 * Node Memory Management (Refactoring #1)
 * ============================================================================
 */

/**
 * framed_route_node_alloc - Allocate a new framed route node
 *
 * Return: Pointer to allocated node, or NULL on failure
 */
struct framed_route_node *framed_route_node_alloc(void);

/**
 * framed_route_node_free - Free a framed route node (direct free)
 * @node: Node to free
 *
 * Note: Only use this when you're certain no RCU readers can access the node,
 * such as in an RCU callback or during error cleanup before the node was
 * added to any hash table.
 */
void framed_route_node_free(struct framed_route_node *node);

/**
 * framed_route_node_free_rcu - Free a framed route node after RCU grace period
 * @node: Node to free
 *
 * Use this when the node might still be accessed by RCU readers.
 */
void framed_route_node_free_rcu(struct framed_route_node *node);

/*
 * ============================================================================
 * CIDR Parsing
 * ============================================================================
 */

/**
 * parse_framed_route_cidr - Parse CIDR format string
 * @route_str: CIDR string (e.g., "192.168.1.0/24")
 * @network_addr: Output network address (network byte order)
 * @netmask: Output netmask (network byte order)
 *
 * Return: 0 on success, negative error code on failure
 */
int parse_framed_route_cidr(const char *route_str, __be32 *network_addr,
                            __be32 *netmask);

/*
 * ============================================================================
 * Hash Table Operations (Refactoring #4)
 * ============================================================================
 */

/**
 * framed_route_hash_add - Add a framed route node to hash table by precedence
 * @gtp: GTP5G device
 * @node: Framed route node to add
 * @precedence: PDR precedence for ordering
 *
 * Nodes are ordered by precedence (lower precedence = higher priority).
 */
void framed_route_hash_add(struct gtp5g_dev *gtp,
                           struct framed_route_node *node,
                           u32 precedence);

/**
 * framed_route_hash_del - Remove a framed route node from hash table
 * @node: Node to remove
 */
void framed_route_hash_del(struct framed_route_node *node);

/**
 * framed_route_cleanup_pdi - Clean up all framed route nodes from a PDI
 * @pdi: PDI containing framed routes
 * @use_rcu: If true, use kfree_rcu; if false, use kfree directly
 *
 * This removes all nodes from hash tables and frees them.
 */
void framed_route_cleanup_pdi(struct pdi *pdi, bool use_rcu);

/*
 * ============================================================================
 * IP Matching Functions
 * ============================================================================
 */

/**
 * ip_match_with_framed_routes - Check if IP matches rule or any framed route
 * @ip_addr: IP address to check (network byte order)
 * @rule_addr: Rule network address
 * @rule_mask: Rule netmask
 * @pdi: PDI containing framed routes (may be NULL)
 *
 * Return: true if IP matches rule or any framed route
 */
bool ip_match_with_framed_routes(const __be32 ip_addr, const __be32 rule_addr,
                                 const __be32 rule_mask, const struct pdi *pdi);

#endif /* __FRAMED_ROUTE_H__ */
