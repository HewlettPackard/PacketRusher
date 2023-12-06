/**
 * Copyright 2023 Edgecom LLC
 * 
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * 
 *     http://www.apache.org/licenses/LICENSE-2.0
 * 
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#pragma once

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/types.h>
#include <sys/socket.h>

#include "xdp/utils/trace.h"

struct route_stat {
    __u64 fib_lookup_ip4_cache;
    __u64 fib_lookup_ip4_ok;
    __u64 fib_lookup_ip4_error_drop;
    __u64 fib_lookup_ip4_error_pass;
    __u64 fib_lookup_ip6_cache;
    __u64 fib_lookup_ip6_ok;
    __u64 fib_lookup_ip6_error_drop;
    __u64 fib_lookup_ip6_error_pass;
};

struct
{
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, __u32);
    __type(value, struct route_stat);
    __uint(max_entries, 1);
} upf_route_stat SEC(".maps");

#ifdef ENABLE_ROUTE_CACHE

#warning "Routing cache enabled"

#define ROUTE_CACHE_IPV4_SIZE 256
#define ROUTE_CACHE_IPV6_SIZE 256

struct route_record {
    int ifindex;
    __u8 smac[6];
    __u8 dmac[6];
};

/* ipv4 -> fib cached result */
struct
{
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, __u32);
    __type(value, struct route_record);
    __uint(max_entries, ROUTE_CACHE_IPV4_SIZE);
} upf_route_cache_ip4 SEC(".maps");

/* ipv6 -> fib cached result */
struct
{
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct in6_addr);
    __type(value, struct route_record);
    __uint(max_entries, ROUTE_CACHE_IPV6_SIZE);
} upf_route_cache_ip6 SEC(".maps");

static __always_inline void update_route_cache_ipv4(const struct bpf_fib_lookup *fib_params, __u32 daddr) {
    struct route_record route = {
        .ifindex = fib_params->ifindex,
    };
    __builtin_memcpy(route.smac, fib_params->smac, ETH_ALEN);
    __builtin_memcpy(route.dmac, fib_params->dmac, ETH_ALEN);
    bpf_map_update_elem(&upf_route_cache_ip4, &daddr, &route, BPF_ANY);
}
#endif

static __always_inline enum xdp_action do_route_ipv4(struct xdp_md *ctx, struct ethhdr *eth, __u32 ifindex, __u8 (*smac)[6], __u8 (*dmac)[6]) {
    //_decr_ttl(ether_proto, l3hdr);
    __builtin_memcpy(eth->h_source, smac, ETH_ALEN);
    __builtin_memcpy(eth->h_dest, dmac, ETH_ALEN);

    if (ifindex == ctx->ingress_ifindex)
        return XDP_TX;
    return bpf_redirect(ifindex, 0);
}

static __always_inline enum xdp_action route_ipv4(struct xdp_md *ctx, struct ethhdr *eth, const struct iphdr *ip4, __u32 ifindex) {
    const __u32 key = 0;
    struct route_stat *statistic = bpf_map_lookup_elem(&upf_route_stat, &key);
    if (!statistic) {
        return XDP_ABORTED;
    }

#ifdef ENABLE_ROUTE_CACHE
    struct route_record *cache = bpf_map_lookup_elem(&upf_route_cache_ip4, &ip4->daddr);
    if (cache) {
        upf_printk("upf: bpf_fib_lookup %pI4 -> %pI4: cached ifindex: %d", &ip4->saddr, &ip4->daddr, cache->ifindex);
        statistic->fib_lookup_ip4_cache += 1;
        return do_route_ipv4(ctx, eth, cache->ifindex, &cache->smac, &cache->dmac);
    }
#endif

    struct bpf_fib_lookup fib_params = {};
    fib_params.family = AF_INET;
    fib_params.tos = ip4->tos;
    fib_params.l4_protocol = ip4->protocol;
    fib_params.sport = 0;
    fib_params.dport = 0;
    fib_params.tot_len = bpf_ntohs(ip4->tot_len);
    //fib_params.ipv4_src = ip4->saddr;
    fib_params.ipv4_dst = ip4->daddr;
    fib_params.ifindex = ifindex;

    int rc = bpf_fib_lookup(ctx, &fib_params, sizeof(fib_params), 0 /*BPF_FIB_LOOKUP_OUTPUT*/);
    switch (rc) {
        case BPF_FIB_LKUP_RET_SUCCESS:
            upf_printk("upf: bpf_fib_lookup1 %pI4 -> %pI4: nexthop: %pI4", &ip4->saddr, &ip4->daddr, &fib_params.ipv4_dst);
            statistic->fib_lookup_ip4_ok += 1;

#ifdef ENABLE_ROUTE_CACHE
            update_route_cache_ipv4(&fib_params, ip4->daddr);
#endif
            return do_route_ipv4(ctx, eth, fib_params.ifindex, &fib_params.smac, &fib_params.dmac);

        case BPF_FIB_LKUP_RET_BLACKHOLE:
        case BPF_FIB_LKUP_RET_UNREACHABLE:
        case BPF_FIB_LKUP_RET_PROHIBIT:
            upf_printk("upf: bpf_fib_lookup2 %pI4 -> %pI4: %d", &ip4->saddr, &ip4->daddr, rc);
            statistic->fib_lookup_ip4_error_drop += 1;
            return XDP_DROP;
        case BPF_FIB_LKUP_RET_NOT_FWDED:
        case BPF_FIB_LKUP_RET_FWD_DISABLED:
        case BPF_FIB_LKUP_RET_UNSUPP_LWT:
        case BPF_FIB_LKUP_RET_NO_NEIGH:
        case BPF_FIB_LKUP_RET_FRAG_NEEDED:
        default:
            upf_printk("upf: bpf_fib_lookup3 %pI4 -> %pI4: %d", &ip4->saddr, &ip4->daddr, rc);
            statistic->fib_lookup_ip4_error_pass += 1;
            return XDP_PASS; /* Let's kernel takes care */
    }
}

static __always_inline enum xdp_action route_ipv6(struct xdp_md *ctx, struct ethhdr *eth, const struct ipv6hdr *ip6) {
    const __u32 key = 0;
    struct route_stat *statistic = bpf_map_lookup_elem(&upf_route_stat, &key);
    if (!statistic) {
        return XDP_ABORTED;
    }

    struct bpf_fib_lookup fib_params = {};
    fib_params.family = AF_INET;
    // fib_params.tos = ip6->flow_lbl;
    fib_params.l4_protocol = ip6->nexthdr;
    fib_params.sport = 0;
    fib_params.dport = 0;
    fib_params.tot_len = bpf_ntohs(ip6->payload_len);
    __builtin_memcpy(fib_params.ipv6_src, &ip6->saddr, sizeof(ip6->saddr));
    __builtin_memcpy(fib_params.ipv6_dst, &ip6->daddr, sizeof(ip6->daddr));
    fib_params.ifindex = ctx->ingress_ifindex;

    int rc = bpf_fib_lookup(ctx, &fib_params, sizeof(fib_params), 0 /*BPF_FIB_LOOKUP_OUTPUT*/);
    switch (rc) {
        case BPF_FIB_LKUP_RET_SUCCESS:
            upf_printk("upf: bpf_fib_lookup %pI6c -> %pI6c: nexthop: %pI4", &ip6->saddr, &ip6->daddr, &fib_params.ipv4_dst);
            statistic->fib_lookup_ip6_ok += 1;
            //_decr_ttl(ether_proto, l3hdr);
            __builtin_memcpy(eth->h_dest, fib_params.dmac, ETH_ALEN);
            __builtin_memcpy(eth->h_source, fib_params.smac, ETH_ALEN);
            upf_printk("upf: bpf_redirect: if=%d %lu -> %lu", fib_params.ifindex, fib_params.smac, fib_params.dmac);

            if (fib_params.ifindex == ctx->ingress_ifindex)
                return XDP_TX;

            return bpf_redirect(fib_params.ifindex, 0);
        case BPF_FIB_LKUP_RET_BLACKHOLE:
        case BPF_FIB_LKUP_RET_UNREACHABLE:
        case BPF_FIB_LKUP_RET_PROHIBIT:
            upf_printk("upf: bpf_fib_lookup %pI6c -> %pI6c: %d", &ip6->saddr, &ip6->daddr, rc);
            statistic->fib_lookup_ip6_error_drop += 1;
            return XDP_DROP;
        case BPF_FIB_LKUP_RET_NOT_FWDED:
        case BPF_FIB_LKUP_RET_FWD_DISABLED:
        case BPF_FIB_LKUP_RET_UNSUPP_LWT:
        case BPF_FIB_LKUP_RET_NO_NEIGH:
        case BPF_FIB_LKUP_RET_FRAG_NEEDED:
        default:
            upf_printk("upf: bpf_fib_lookup %pI6c -> %pI6c: %d", &ip6->saddr, &ip6->daddr, rc);
            statistic->fib_lookup_ip6_error_pass += 1;
            return XDP_PASS; /* Let's kernel takes care */
    }
}