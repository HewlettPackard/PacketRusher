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
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/types.h>
#include <linux/udp.h>
#include <linux/tcp.h>

#include "xdp/utils/packet_context.h"
#include "xdp/utils/trace.h"

#define ETH_P_IPV6_BE	0xDD86
#define ETH_P_IP_BE 	0x0008

static __always_inline int parse_ethernet(struct packet_context *ctx) {
    struct ethhdr *eth = (struct ethhdr *)ctx->data;
    if ((const char *)(eth + 1) > ctx->data_end)
        return -1;

    /* TODO: Add vlan support */

    ctx->data += sizeof(*eth);
    ctx->eth = eth;
    return bpf_ntohs(eth->h_proto);
}

/* 0x3FFF mask to check for fragment offset field */
#define IP_FRAGMENTED 65343

static __always_inline int parse_ip4(struct packet_context *ctx) {
    struct iphdr *ip4 = (struct iphdr *)ctx->data;
    if ((const char *)(ip4 + 1) > ctx->data_end)
        return -1;

    /* do not support fragmented packets as L4 headers may be missing */
    // if (ip4->frag_off & IP_FRAGMENTED)
    //	return -1;

    ctx->data += ip4->ihl * 4; /* header + options */
    ctx->ip4 = ip4;
    return ip4->protocol;
}

static __always_inline int parse_ip6(struct packet_context *ctx) {
    struct ipv6hdr *ip6 = (struct ipv6hdr *)ctx->data;
    if ((const char *)(ip6 + 1) > ctx->data_end)
        return -1;
    
    /* TODO: Add extention headers support */

    ctx->data += sizeof(*ip6);
    ctx->ip6 = ip6;
    return ip6->nexthdr;
}

static __always_inline int parse_udp(struct packet_context *ctx) {
    struct udphdr *udp = (struct udphdr *)ctx->data;
    if ((const char *)(udp + 1) > ctx->data_end)
        return -1;

    ctx->data += sizeof(*udp);
    ctx->udp = udp;
    return bpf_ntohs(udp->dest);
}

static __always_inline int parse_tcp(struct packet_context *ctx) {
    struct tcphdr *tcp = (struct tcphdr *)ctx->data;
    if ((const char *)(tcp + 1) > ctx->data_end)
        return -1;

    //TODO: parse header lenght correctly (tcp options)

    ctx->data += sizeof(*tcp);
    ctx->tcp = tcp;
    return bpf_ntohs(tcp->dest);
}

static __always_inline int parse_l4(int ip_protocol, struct packet_context *ctx)
{
    switch (ip_protocol) {
        case IPPROTO_UDP:
            return parse_udp(ctx);
        case IPPROTO_TCP:
            return parse_tcp(ctx);
        default:
            return 0;
    }
}

static __always_inline void swap_mac(struct ethhdr *eth) {
    __u8 mac[6];
    __builtin_memcpy(mac, eth->h_source, sizeof(mac));
    __builtin_memcpy(eth->h_source, eth->h_dest, sizeof(eth->h_source));
    __builtin_memcpy(eth->h_dest, mac, sizeof(eth->h_dest));
}

static __always_inline void swap_port(struct udphdr *udp) {
    __u16 tmp = udp->dest;
    udp->dest = udp->source;
    udp->source = tmp;
    /* Update UDP checksum */
    udp->check = 0;
    // cs = 0;
    // ipv4_l4_csum(udp, sizeof(*udp), &cs, iph);
    // udp->check = cs;
}

static __always_inline void swap_ip(struct iphdr *iph) {
    __u32 tmp_ip = iph->daddr;
    iph->daddr = iph->saddr;
    iph->saddr = tmp_ip;

    /* Don't need to recalc csum in case of ip swap */
    // ip->check = ipv4_csum(ip, sizeof(*ip));
}

static __always_inline void context_set_ip4(struct packet_context *ctx, char *data, const char *data_end, struct ethhdr *eth, struct iphdr *ip4, struct udphdr *udp, struct gtpuhdr *gtp) {
    ctx->data = data;
    ctx->data_end = data_end;
    ctx->eth = eth;
    ctx->ip4 = ip4;
    ctx->ip6 = 0;
    ctx->udp = udp;
    ctx->gtp = gtp;
}

static __always_inline void context_reset(struct packet_context *ctx, char *data, const char *data_end) {
    ctx->data = data;
    ctx->data_end = data_end;
    ctx->eth = 0;
    ctx->ip4 = 0;
    ctx->ip6 = 0;
    ctx->udp = 0;
    ctx->gtp = 0;
}


static __always_inline long context_reinit(struct packet_context *ctx, char *data, const char *data_end) {
    context_reset(ctx, data, data_end);

    int ethertype = parse_ethernet(ctx);
    switch (ethertype) {
        case ETH_P_IPV6: {
            if (-1 == parse_ip6(ctx)) {
                upf_printk("upf: can't parse ip6");
                return -1;
            }
            return 0;
        }
        case ETH_P_IP: {
            if (-1 == parse_ip4(ctx)) {
                upf_printk("upf: can't parse ip4");
                return -1;
            }
            return 0;
        }

        default:
            /* do nothing with non-ip packets */
            upf_printk("upf: can't process not an ip packet: %d", ethertype);
            return -1;
    }
}
