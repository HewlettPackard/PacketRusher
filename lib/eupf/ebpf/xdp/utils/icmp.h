// Copyright 2023 Edgecom LLC
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#pragma once

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/types.h>
#include <linux/icmp.h>

#include "xdp/utils/csum.h"
#include "xdp/utils/parsers.h"
#include "xdp/utils/packet_context.h"

static __always_inline void fill_icmp_header(struct icmphdr *icmp) {
    icmp->type = ICMP_TIME_EXCEEDED;
    icmp->code = ICMP_EXC_TTL;
    icmp->un.gateway = 0;
    icmp->checksum = 0;
}

// static __always_inline __u32 add_icmp_over_ip4_headers(struct packet_context *ctx, int saddr, int daddr) {
//     static const size_t icmp_encap_size = sizeof(struct iphdr) + sizeof(struct icmphdr);

//     if (!ctx->ip4)
//         return -1;

//     const __u32 ip_packet_len = bpf_ntohs(ctx->ip4->tot_len);

//     int result = bpf_xdp_adjust_head(ctx->xdp_ctx, (__s32)-icmp_encap_size);
//     if (result)
//         return -1;

//     char *data = (char *)(long)ctx->xdp_ctx->data;
//     const char *data_end = (const char *)(long)ctx->xdp_ctx->data_end;

//     struct ethhdr *orig_eth = (struct ethhdr *)(data + icmp_encap_size);
//     if ((const char *)(orig_eth + 1) > data_end)
//         return -1;

//     struct ethhdr *eth = (struct ethhdr *)data;
//     __builtin_memcpy(eth, orig_eth, sizeof(*eth));
//     eth->h_proto = bpf_htons(ETH_P_IP);

//     struct iphdr *ip = (struct iphdr *)(eth + 1);
//     if ((const char *)(ip + 1) > data_end)
//         return -1;

//     /* Add the outer IP header */
//     fill_ip_header(ip, saddr, daddr, 0, ip_packet_len + icmp_encap_size);
//     ip->protocol = IPPROTO_ICMP;
//     ip->check = ipv4_csum(ip, sizeof(*ip));

//     /* Add the ICMP header */
//     struct icmphdr *icmp = (struct icmphdr *)(ip + 1);
//     if ((const char *)(icmp + 1) > data_end)
//         return -1;

//     fill_icmp_header(icmp);
//     const __s8 icmp_payload_size = data_end - (const char *)icmp;
//     icmp->checksum = ipv4_csum(icmp, icmp_payload_size);

//     /* Update packet pointers */
//     context_set_ip4(ctx, (char *)(long)ctx->xdp_ctx->data, (const char *)(long)ctx->xdp_ctx->data_end, eth, ip, 0, 0);
//     return 0;
// }

static __always_inline __u32 prepare_icmp_echo_reply(struct packet_context *ctx, int saddr, int daddr) {
    if (!ctx->ip4)
        return -1;

    struct ethhdr *eth = ctx->eth;
    swap_mac(eth);

    const char *data_end = (const char *)(long)ctx->xdp_ctx->data_end;
    struct iphdr *ip = ctx->ip4;
    if ((const char *)(ip + 1) > data_end)
        return -1;

    swap_ip(ip);

    struct icmphdr *icmp = (struct icmphdr *)(ip + 1);
    if ((const char *)(icmp + 1) > data_end)
        return -1;

    if(icmp->type != ICMP_ECHO)
        return -1;

    __u16 old = *(__u16*)&icmp->type;
    icmp->type = ICMP_ECHOREPLY;
    icmp->code = 0;
    
    ipv4_csum_replace(&icmp->checksum, old, *(__u16*)&icmp->type);

    return 0;
}