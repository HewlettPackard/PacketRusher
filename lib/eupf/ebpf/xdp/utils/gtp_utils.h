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

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/types.h>
#include <linux/udp.h>
#include <linux/icmp.h>

#include "xdp/utils/gtpu.h"
#include "xdp/utils/packet_context.h"
#include "xdp/utils/trace.h"

static __always_inline __u32 parse_gtp(struct packet_context *ctx) {
    struct gtpuhdr *gtp = (struct gtpuhdr *)ctx->data;
    if ((const char *)(gtp + 1) > ctx->data_end)
        return -1;

    ctx->data += sizeof(*gtp);
    if (gtp->e || gtp->s || gtp->pn)
        ctx->data += sizeof(struct gtp_hdr_ext) + 4;
    ctx->gtp = gtp;
    return gtp->message_type;
}

static __always_inline __u32 handle_echo_request(struct packet_context *ctx) {
    struct ethhdr *eth = ctx->eth;
    struct iphdr *iph = ctx->ip4;
    struct udphdr *udp = ctx->udp;
    struct gtpuhdr *gtp = ctx->gtp;

    gtp->message_type = GTPU_ECHO_RESPONSE;

    /* TODO: add support GTP over IPv6 */
    swap_ip(iph);
    swap_port(udp);
    swap_mac(eth);
    upf_printk("upf: send gtp echo response [ %pI4 -> %pI4 ]", &iph->saddr, &iph->daddr);
    return XDP_TX;
}

static __always_inline int guess_eth_protocol(const char *data) {
    const __u8 ip_version = (*(const __u8 *)data) >> 4;
    switch (ip_version) {
        case 6: {
            return ETH_P_IPV6_BE;
        }
        case 4: {
            return ETH_P_IP_BE;
        }
        default:
            /* do nothing with non-ip packets */
            upf_printk("upf: can't process non-IP packet: %d", ip_version);
            return -1;
    }
}

static __always_inline long remove_gtp_header(struct packet_context *ctx) {
    if (!ctx->gtp) {
        upf_printk("upf: remove_gtp_header: not a gtp packet");
        return -1;
    }

    size_t ext_gtp_header_size = 0;
    struct gtpuhdr *gtp = ctx->gtp;
    if (gtp->e || gtp->s || gtp->pn)
        ext_gtp_header_size += sizeof(struct gtp_hdr_ext) + 4;

    const size_t gtp_encap_size = sizeof(struct iphdr) + sizeof(struct udphdr) + sizeof(struct gtpuhdr) + ext_gtp_header_size;

    char *data = (char *)(long)ctx->xdp_ctx->data;
    const char *data_end = (const char *)(long)ctx->xdp_ctx->data_end;
    struct ethhdr *eth = (struct ethhdr *)data;
    if ((const char *)(eth + 1) > data_end) {
        upf_printk("upf: remove_gtp_header: can't parse eth");
        return -1;
    }

    data += gtp_encap_size;
    struct ethhdr *new_eth = (struct ethhdr *)data;
    if ((const char *)(new_eth + 1) > data_end) {
        upf_printk("upf: remove_gtp_header: can't set new eth");
        return -1;
    }

    data += sizeof(*new_eth);
    if (data + 1 > data_end)
        return -1;

    const int eth_proto = guess_eth_protocol(data);

    if (eth_proto == -1)
        return -1;

    __builtin_memcpy(new_eth, eth, sizeof(*new_eth));
    
    new_eth->h_proto = eth_proto;

    long result = bpf_xdp_adjust_head(ctx->xdp_ctx, gtp_encap_size);
    if (result)
        return result;

    /* Update packet pointers */
    data = (char *)(long)ctx->xdp_ctx->data;
    data_end = (const char *)(long)ctx->xdp_ctx->data_end;
    return context_reinit(ctx, data, data_end);
}

static __always_inline void fill_ip_header(struct iphdr *ip, int saddr, int daddr, __u8 tos, int tot_len) {
    ip->version = 4;
    ip->ihl = 5; /* No options */
    ip->tos = tos;
    ip->tot_len = bpf_htons(tot_len);
    ip->id = 0;            /* No fragmentation */
    ip->frag_off = 0x0040; /* Don't fragment; Fragment offset = 0 */
    ip->ttl = 64;
    ip->protocol = IPPROTO_UDP;
    ip->check = 0;
    ip->saddr = saddr;
    ip->daddr = daddr;
}

static __always_inline void fill_udp_header(struct udphdr *udp, int port, int len) {
    udp->source = bpf_htons(port);
    udp->dest = udp->source;
    udp->len = bpf_htons(len);
    udp->check = 0;
}

static __always_inline void fill_gtp_header(struct gtpuhdr *gtp, int teid, int len) {
    *(__u8 *)gtp = GTP_FLAGS;
    gtp->e = 1;
    gtp->message_type = GTPU_G_PDU;
    gtp->message_length = bpf_htons(len);
    gtp->teid = bpf_htonl(teid);
}

static __always_inline void fill_gtp_ext_header(struct gtp_hdr_ext *gtp_ext) {
    gtp_ext->sqn = 0;
    gtp_ext->npdu = 0;
    gtp_ext->next_ext = GTPU_EXT_TYPE_PDU_SESSION_CONTAINER;
}

static __always_inline void fill_gtp_ext_header_psc(struct gtp_hdr_ext_pdu_session_container *gtp_ext, int qfi, int pdu_type) {
    gtp_ext->length = 1;
    gtp_ext->pdu_type = pdu_type;
    gtp_ext->spare1 = 0;
    gtp_ext->spare2 = 0;
    gtp_ext->rqi = 0;
    gtp_ext->qfi = qfi;
    gtp_ext->next_ext = 0;
}

static __always_inline __u32 add_gtp_over_ip4_headers(struct packet_context *ctx, int saddr, int daddr, __u8 tos, __u8 qfi, int teid) {
    static const size_t gtp_ext_hdr_size = sizeof(struct gtp_hdr_ext) + sizeof(struct gtp_hdr_ext_pdu_session_container);
    static const size_t gtp_full_hdr_size = sizeof(struct gtpuhdr) + gtp_ext_hdr_size;
    static const size_t gtp_encap_size = sizeof(struct iphdr) + sizeof(struct udphdr) + gtp_full_hdr_size;

    // int ip_packet_len = (ctx->xdp_ctx->data_end - ctx->xdp_ctx->data) - sizeof(*eth);
    int ip_packet_len = 0;
    if (ctx->ip4)
        ip_packet_len = bpf_ntohs(ctx->ip4->tot_len);
    else if (ctx->ip6)
        ip_packet_len = bpf_ntohs(ctx->ip6->payload_len) + sizeof(struct ipv6hdr);
    else
        return -1;

    int result = bpf_xdp_adjust_head(ctx->xdp_ctx, (__s32)-gtp_encap_size);
    if (result)
        return -1;

    char *data = (char *)(long)ctx->xdp_ctx->data;
    const char *data_end = (const char *)(long)ctx->xdp_ctx->data_end;

    struct ethhdr *orig_eth = (struct ethhdr *)(data + gtp_encap_size);
    if ((const char *)(orig_eth + 1) > data_end)
        return -1;

    struct ethhdr *eth = (struct ethhdr *)data;
    __builtin_memcpy(eth, orig_eth, sizeof(*eth));
    eth->h_proto = bpf_htons(ETH_P_IP);

    struct iphdr *ip = (struct iphdr *)(eth + 1);
    if ((const char *)(ip + 1) > data_end)
        return -1;

    /* Add the outer IP header */
    fill_ip_header(ip, saddr, daddr, tos, ip_packet_len + gtp_encap_size);

    /* Add the UDP header */
    struct udphdr *udp = (struct udphdr *)(ip + 1);
    if ((const char *)(udp + 1) > data_end)
        return -1;

    fill_udp_header(udp, GTP_UDP_PORT, ip_packet_len + sizeof(*udp) + gtp_full_hdr_size);

    /* Add the GTP header */
    struct gtpuhdr *gtp = (struct gtpuhdr *)(udp + 1);
    if ((const char *)(gtp + 1) > data_end)
        return -1;

    fill_gtp_header(gtp, teid, gtp_ext_hdr_size + ip_packet_len);

    /* Add the GTP ext header */
    struct gtp_hdr_ext *gtp_ext = (struct gtp_hdr_ext *)(gtp + 1);
    if ((const char *)(gtp_ext + 1) > data_end)
        return -1;

    fill_gtp_ext_header(gtp_ext);

    /* Add the GTP PDU session container header */
    struct gtp_hdr_ext_pdu_session_container *gtp_psc = (struct gtp_hdr_ext_pdu_session_container *)(gtp_ext + 1);
    if ((const char *)(gtp_psc + 1) > data_end)
        return -1;

    fill_gtp_ext_header_psc(gtp_psc, qfi, PDU_SESSION_CONTAINER_PDU_TYPE_DL_PSU);

    ip->check = ipv4_csum(ip, sizeof(*ip));

    /* TODO: implement UDP csum which pass ebpf verifier checks successfully */
    // cs = 0;
    // const void* udp_start = (void*)udp;
    // const __u16 udp_len = bpf_htons(udp->len);
    // ipv4_l4_csum(udp, udp_len, &cs, ip);
    // udp->check = cs;

    /* Update packet pointers */
    context_set_ip4(ctx, (char *)(long)ctx->xdp_ctx->data, (const char *)(long)ctx->xdp_ctx->data_end, eth, ip, udp, gtp);
    return 0;
}

static __always_inline void update_gtp_tunnel(struct packet_context *ctx, int srcip, int dstip, __u8 tos, int teid) {

    ctx->gtp->teid = bpf_htonl(teid);
    ctx->ip4->saddr = srcip;
    ctx->ip4->daddr = dstip;
    ctx->ip4->check = 0;
    ctx->ip4->check = ipv4_csum(ctx->ip4, sizeof(*ctx->ip4));
}
