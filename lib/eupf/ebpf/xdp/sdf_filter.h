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

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>
#include <linux/bpf.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/tcp.h>
#include <linux/udp.h>

#include "xdp/utils/trace.h"
#include "xdp/utils/packet_context.h"
#include "xdp/utils/types.h"
#include "xdp/pdr.h"

struct ip_subnet {
    __u8 type; // 0: any, 1: ip4, 2: ip6
    // If type != any, ip field has meaningful value.
    // If IPv4 -> lower 32 bits. If IPv6 -> all 128 bits.
    __u128 ip;
    // If type != any, mask field has meaningful value.
    // If IPv4 mask -> lower 32 bits. If IPv6 mask -> all 128 bits.
    // Should always be applied to matching ip (except type == any).
    __u128 mask;
};

struct port_range {
    __u16 lower_bound; // If not specified in SDF: 0
    __u16 upper_bound; // If not specified in SDF: 65535
};

struct sdf_filter {
    __u8 protocol; // Required by SDF. 0: icmp, 1: ip, 2: tcp, 3: udp
    struct ip_subnet src_addr;
    struct port_range src_port;
    struct ip_subnet dst_addr;
    struct port_range dst_port;
};


static __always_inline __u8 get_sdf_protocol(__u8 ip_protocol) {
    switch(ip_protocol)
    {
        case IPPROTO_ICMP: return 0;
        case IPPROTO_TCP: return 2;
        case IPPROTO_UDP: return 3;
        default: return 1;
    }
}

static __always_inline __u8 match_sdf_filter_ipv4(const struct packet_context *ctx, const struct sdf_filter *sdf) {
    if(!ctx || !ctx->ip4 || !sdf)
        return 0;

    const struct iphdr *ip4 = ctx->ip4;
    __u8 packet_protocol = get_sdf_protocol(ip4->protocol); //TODO: convert protocol in golang part
    __u16 packet_src_port = 0;
    __u16 packet_dst_port = 0;
    if(ctx->udp) {
        packet_src_port = bpf_ntohs(ctx->udp->source); //TODO: convert port in golang part
        packet_dst_port = bpf_ntohs(ctx->udp->dest); //TODO: convert port in golang part
    } else if (ctx->tcp) {
        packet_src_port = bpf_ntohs(ctx->tcp->source); //TODO: convert port in golang part
        packet_dst_port = bpf_ntohs(ctx->tcp->dest); //TODO: convert port in golang part
    } 

    __u32 sdf_src_ip = bpf_htonl(sdf->src_addr.ip);
    __u32 sdf_dst_ip = bpf_htonl(sdf->dst_addr.ip);
    __u32 sdf_src_mask = bpf_htonl(sdf->src_addr.mask);
    __u32 sdf_dst_mask = bpf_htonl(sdf->dst_addr.mask);
    upf_printk("SDF: filter protocol: %u", sdf->protocol);
    upf_printk("SDF: filter source ip: %pI4, destination ip: %pI4",  &sdf_src_ip,  &sdf_dst_ip);
    upf_printk("SDF: filter source ip mask: %pI4, destination ip mask: %pI4", &sdf_src_mask, &sdf_dst_mask);
    upf_printk("SDF: filter source port lower bound: %u, source port upper bound: %u", sdf->src_port.lower_bound, sdf->src_port.upper_bound);
    upf_printk("SDF: filter destination port lower bound: %u, destination port upper bound: %u", sdf->dst_port.lower_bound, sdf->dst_port.upper_bound);

    upf_printk("SDF: packet protocol: %u", packet_protocol);
    upf_printk("SDF: packet source ip: %pI4, destination ip: %pI4",  &ip4->saddr,  &ip4->daddr);
    upf_printk("SDF: packet source port: %u, destination port: %u", packet_src_port, packet_dst_port);
    
    if ((sdf->protocol != 1 && sdf->protocol != packet_protocol) 
        || ((ip4->saddr & sdf_src_mask) != sdf_src_ip)  
        || ((ip4->daddr & sdf_dst_mask) != sdf_dst_ip) 
        || (packet_src_port < sdf->src_port.lower_bound || packet_src_port > sdf->src_port.upper_bound) 
        || (packet_dst_port < sdf->dst_port.lower_bound || packet_dst_port > sdf->dst_port.upper_bound)) 
    {
        return 0;
    }
    
    upf_printk("Packet with source ip: %pI4, destination ip: %pI4 matches SDF filter",
               &ip4->saddr, &ip4->daddr);

    return 1;
}

static __always_inline __u8 match_sdf_filter_ipv6(const struct packet_context *ctx, const struct sdf_filter *sdf) {
    const struct ipv6hdr *ipv6 = ctx->ip6;  
    __u8 packet_protocol = get_sdf_protocol(ipv6->nexthdr);
    __u128 packet_src_ip_128 = *((__u128*)ipv6->saddr.s6_addr);
    __u128 packet_dst_ip_128 = *((__u128*)ipv6->daddr.s6_addr);
    __u16 packet_src_port = 0;
    __u16 packet_dst_port = 0;
    if(ctx->udp) {
        packet_src_port = bpf_ntohs(ctx->udp->source); //TODO: convert port in golang part
        packet_dst_port = bpf_ntohs(ctx->udp->dest); //TODO: convert port in golang part
    } else if (ctx->tcp) {
        packet_src_port = bpf_ntohs(ctx->tcp->source); //TODO: convert port in golang part
        packet_dst_port = bpf_ntohs(ctx->tcp->dest); //TODO: convert port in golang part
    }
    
    __u128 sdf_src_ip = bpf_htonl(sdf->src_addr.ip);
    __u128 sdf_dst_ip = bpf_htonl(sdf->dst_addr.ip);
    upf_printk("SDF: filter protocol: %u", sdf->protocol);
    upf_printk("SDF: filter source ip: %pI6c, destination ip: %pI6c",  &sdf_src_ip,  &sdf_dst_ip);
    upf_printk("SDF: filter source port lower bound: %u, source port upper bound: %u", sdf->src_port.lower_bound, sdf->src_port.upper_bound);
    upf_printk("SDF: filter destination port lower bound: %u, destination port upper bound: %u", sdf->dst_port.lower_bound, sdf->dst_port.upper_bound);
    upf_printk("SDF: filter source address mask: %pI4, destination address mask: %pI4", &sdf->dst_addr.mask, &sdf->dst_addr.mask);

    upf_printk("SDF: packet protocol: %u", packet_protocol);
    upf_printk("SDF: packet source ip: %pI6c, destination ip: %pI6c",  &packet_src_ip_128,  &packet_dst_ip_128);
    upf_printk("SDF: packet source port: %u, destination port: %u", packet_src_port, packet_dst_port);

    if ((sdf->protocol != 1 && sdf->protocol != packet_protocol) ||
        (packet_src_ip_128 & sdf->src_addr.mask) != sdf_src_ip || 
        (packet_dst_ip_128 & sdf->dst_addr.mask) != sdf_dst_ip ||
         packet_src_port < sdf->src_port.lower_bound || packet_src_port > sdf->src_port.upper_bound ||
         packet_dst_port < sdf->dst_port.lower_bound || packet_dst_port > sdf->dst_port.upper_bound) {
        return 0;
    }

    upf_printk("SDF: packet with source ip:%pI6c, destination ip:%pI6c matches SDF filter",
               &packet_src_ip_128, &packet_dst_ip_128);

    return 1;
}