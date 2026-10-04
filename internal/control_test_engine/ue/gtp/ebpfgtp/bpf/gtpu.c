// SPDX-License-Identifier: Apache-2.0
// PacketRusher's dual-stack UE GTP-U datapath over IPv4 N3. No PFCP/UPF implementation is included.
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/udp.h>
#include <linux/pkt_cls.h>

#define SEC(name) __attribute__((section(name), used))
#define INLINE static __attribute__((always_inline)) inline
#define __uint(name, value) int (*name)[value]
#define __type(name, value) value *name
#define htons(x) __builtin_bswap16((__u16)(x))
#define htonl(x) __builtin_bswap32((__u32)(x))
#define ntohs(x) htons(x)
#define ntohl(x) htonl(x)
#define NEXT (-1)
#define MAX_N3_MTU 65535
#define OVERHEAD 44

static long (*bpf_loop)(__u32, void *, void *, __u64) = (void *)BPF_FUNC_loop;
static void *(*lookup)(void *, const void *) = (void *)BPF_FUNC_map_lookup_elem;
static long (*change_type)(struct __sk_buff *, __u32) = (void *)BPF_FUNC_skb_change_type;
static long (*change_head)(struct __sk_buff *, __u32, __u64) = (void *)BPF_FUNC_skb_change_head;
static long (*store_bytes)(struct __sk_buff *, __u32, const void *, __u32, __u64) = (void *)BPF_FUNC_skb_store_bytes;
static long (*redirect_neigh)(__u32, struct bpf_redir_neigh *, int, __u64) = (void *)BPF_FUNC_redirect_neigh;
static long (*load_bytes)(struct __sk_buff *, __u32, void *, __u32) = (void *)BPF_FUNC_skb_load_bytes;
static long (*adjust_room)(struct __sk_buff *, __s32, __u32, __u64) = (void *)BPF_FUNC_skb_adjust_room;
static long (*redirect)(__u32, __u64) = (void *)BPF_FUNC_redirect;

struct binding {
    __u32 local, peer, ue, uplink_teid, downlink_teid, endpoint, mtu, next_hop, stage_tx;
    __u8 qfi, ipv6, prefix_valid, padding;
    __u8 iid[8], prefix[8];
};
struct down_key { __u32 local, peer, teid; };
struct {
    __uint(type, BPF_MAP_TYPE_HASH); __uint(max_entries, 4096);
    __type(key, __u32); __type(value, struct binding);
} sessions SEC(".maps");
struct {
    __uint(type, BPF_MAP_TYPE_HASH); __uint(max_entries, 8192);
    __type(key, struct down_key); __type(value, __u32);
} downlinks SEC(".maps");
struct {
    __uint(type, BPF_MAP_TYPE_HASH); __uint(max_entries, 256);
    __type(key, __u32); __type(value, __u32);
} locals SEC(".maps");
struct {
    __uint(type, BPF_MAP_TYPE_HASH); __uint(max_entries, 4096);
    __type(key, __u32); __type(value, __u32);
} stages SEC(".maps");
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY); __uint(max_entries, 5);
    __type(key, __u32); __type(value, __u64);
} counters SEC(".maps");

INLINE void count(__u32 index) {
    volatile __u64 key = index;
    __u64 *n = lookup(&counters, (void *)&key);
    if (n) __sync_fetch_and_add(n, 1);
}
INLINE __u16 fold(__u32 sum) {
    sum = (sum & 0xffff) + (sum >> 16);
    sum = (sum & 0xffff) + (sum >> 16);
    return sum;
}
// Initialized helper reads also cover options without variable packet pointers.
static __attribute__((noinline)) int ip_checksum(struct __sk_buff *skb, __u32 offset, __u32 size) {
    if (size < 20 || size > 60 || (size & 3)) return 0;
    __u32 words[15] = {};
    if (load_bytes(skb, offset, words, size)) return 0;
    __u32 sum = 0;
#pragma unroll
    for (int i = 0; i < 15; i++) sum += (words[i] & 0xffff) + (words[i] >> 16);
    return fold(sum) == 0xffff;
}
struct checksum_context {
    struct __sk_buff *skb;
    __u32 offset, length, sum, failed;
};
static long udp_checksum_chunk(__u32 index, struct checksum_context *ctx) {
    __u32 offset = index * 32;
    if (offset >= ctx->length) return 1;
    __u32 size = ctx->length - offset;
    asm volatile("" : "+r"(size));
    if (size > 32) size = 32;
    if (!size) return 1;
    unsigned char bytes[32] = {};
    if (load_bytes(ctx->skb, ctx->offset + offset, bytes, size)) { ctx->failed = 1; return 1; }
#pragma unroll
    for (int i = 0; i < 16; i++) ctx->sum += ((__u32)bytes[i * 2] << 8) | bytes[i * 2 + 1];
    return 0;
}

struct gtp_header {
    __u8 flags, type; __u16 length; __u32 teid;
    __u16 sequence; __u8 npdu, next;
    __u8 extension_length, pdu_type, qfi, end;
} __attribute__((packed));
struct encap_header { struct iphdr ip; struct udphdr udp; struct gtp_header gtp; };

INLINE int ipv6_owned(const struct in6_addr *address, const struct binding *cfg) {
    if (!cfg->ipv6) return 0;
#pragma unroll
    for (int i = 0; i < 8; i++) if (address->s6_addr[i + 8] != cfg->iid[i]) return 0;
    // Same allocated IID; link-local traffic precedes prefix discovery.
    if (address->s6_addr[0] == 0xfe && address->s6_addr[1] == 0x80) {
#pragma unroll
        for (int i = 2; i < 8; i++) if (address->s6_addr[i]) return 0;
        return 1;
    }
    if (!cfg->prefix_valid) return 0;
#pragma unroll
    for (int i = 0; i < 8; i++) if (address->s6_addr[i] != cfg->prefix[i]) return 0;
    return 1;
}

INLINE int uplink_owned(struct __sk_buff *skb, __u32 offset, __u32 length, const struct binding *cfg) {
    unsigned char version = 0;
    if (load_bytes(skb, offset, &version, 1)) return 0;
    if (version >> 4 == 4) {
        struct iphdr ip = {};
        if (load_bytes(skb, offset, &ip, sizeof(ip)) || ip.ihl < 5 || ip.ihl * 4 > length ||
            !cfg->ue || ip.saddr != cfg->ue || ntohs(ip.tot_len) != length) return 0;
    } else if (version >> 4 == 6) {
        struct ipv6hdr ip = {};
        if (load_bytes(skb, offset, &ip, sizeof(ip)) ||
            ntohs(ip.payload_len) + 40 != length || !ipv6_owned(&ip.saddr, cfg)) return 0;
    } else return 0;
    return 1;
}

// TUN egress is L3. Bind ownership to its stable index before examining a
// source, so a device-bound application cannot borrow another UE mapping.
SEC("tc/egress")
int encap(struct __sk_buff *skb) {
    __u32 endpoint = skb->ifindex;
    struct binding *value = lookup(&sessions, &endpoint);
    if (!value) return TC_ACT_SHOT;
    struct binding cfg = *value;
    if (cfg.endpoint != endpoint || !cfg.stage_tx || skb->len > cfg.mtu || skb->gso_segs > 1) return TC_ACT_SHOT;
    if (!uplink_owned(skb, 0, skb->len, &cfg)) return TC_ACT_SHOT;
    // Kernel room helpers limit non-GSO growth by SKB_MAX_ALLOC. Leave larger
    // valid packets on the owned TUN for normal kernel checksum completion and
    // the joined, current-owner UDP sender instead of failing a jumbo MTU.
    if (skb->len > 8000) { count(4); return TC_ACT_OK; }
    __u32 *n3 = lookup(&locals, &cfg.local);
    if (!n3) return TC_ACT_SHOT;
    __u32 overhead = cfg.qfi ? 44 : 36;
    struct encap_header h = {};
    h.ip.version = 4; h.ip.ihl = 5; h.ip.ttl = 64; h.ip.protocol = 17;
    h.ip.tot_len = htons(skb->len + overhead);
    h.ip.frag_off = htons(0x4000);
    h.ip.saddr = cfg.local; h.ip.daddr = cfg.peer;
    __u32 sum = 0;
    unsigned char *bytes = (void *)&h.ip;
#pragma unroll
    for (int i = 0; i < 10; i++) sum += ((__u32)bytes[i * 2] << 8) | bytes[i * 2 + 1];
    h.ip.check = htons(~fold(sum));
    h.udp.source = htons(2152); h.udp.dest = htons(2152);
    h.udp.len = htons(skb->len + overhead - 20);
    h.gtp.flags = cfg.qfi ? 0x34 : 0x30; h.gtp.type = 255; h.gtp.length = htons(skb->len + overhead - 36);
    h.gtp.teid = htonl(cfg.uplink_teid); h.gtp.next = 0x85;
    h.gtp.extension_length = 1; h.gtp.pdu_type = 0x10; h.gtp.qfi = cfg.qfi;
    if (adjust_room(skb, overhead, BPF_ADJ_ROOM_MAC,
        BPF_F_ADJ_ROOM_ENCAP_L3_IPV4 | BPF_F_ADJ_ROOM_ENCAP_L4_UDP | BPF_F_ADJ_ROOM_NO_CSUM_RESET) ||
        store_bytes(skb, 0, &h, overhead, 0) || change_head(skb, ETH_HLEN, 0)) return TC_ACT_SHOT;
    struct ethhdr eth = {.h_proto = htons(ETH_P_IP)};
    if (store_bytes(skb, 0, &eth, sizeof(eth), 0)) return TC_ACT_SHOT;
    // A real checksum-off veth transmission completes and clears PARTIAL.
    // TCX redirect itself bypasses that completion on the original TUN.
    return redirect(cfg.stage_tx, 0);
}

SEC("tc/ingress")
int relay(struct __sk_buff *skb) {
    __u32 stage_peer = skb->ingress_ifindex;
    __u32 *endpoint = lookup(&stages, &stage_peer);
    if (!endpoint) return TC_ACT_SHOT;
    struct binding *value = lookup(&sessions, endpoint);
    if (!value) return TC_ACT_SHOT;
    struct binding cfg = *value;
    struct iphdr outer = {};
    struct gtp_header gtp = {};
    if (cfg.endpoint != *endpoint || !cfg.stage_tx ||
        load_bytes(skb, ETH_HLEN, &outer, sizeof(outer)) ||
        outer.version != 4 || outer.ihl != 5 || outer.protocol != 17 ||
        outer.saddr != cfg.local || outer.daddr != cfg.peer ||
        load_bytes(skb, ETH_HLEN + 28, &gtp, 8) ||
        gtp.type != 255 || ntohl(gtp.teid) != cfg.uplink_teid) return TC_ACT_SHOT;
    __u32 overhead = cfg.qfi ? 44 : 36;
    struct udphdr udp = {};
    if (skb->len < ETH_HLEN + overhead || ntohs(outer.tot_len) != skb->len - ETH_HLEN ||
        load_bytes(skb, ETH_HLEN + 20, &udp, sizeof(udp)) ||
        udp.source != htons(2152) || udp.dest != htons(2152) ||
        ntohs(udp.len) != skb->len - ETH_HLEN - 20 ||
        ntohs(gtp.length) != skb->len - ETH_HLEN - 36 ||
        gtp.flags != (cfg.qfi ? 0x34 : 0x30)) return TC_ACT_SHOT;
    if (cfg.qfi && (load_bytes(skb, ETH_HLEN + 36, &gtp.sequence, 8) ||
        gtp.sequence || gtp.npdu || gtp.next != 0x85 || gtp.extension_length != 1 ||
        gtp.pdu_type != 0x10 || gtp.qfi != cfg.qfi || gtp.end)) return TC_ACT_SHOT;
    __u32 inner_length = skb->len - ETH_HLEN - overhead;
    if (inner_length > cfg.mtu || !uplink_owned(skb, ETH_HLEN + overhead, inner_length, &cfg)) return TC_ACT_SHOT;
    __u32 *n3 = lookup(&locals, &cfg.local);
    if (!n3) return TC_ACT_SHOT;
    __u32 n3_index = *n3 & 0x7fffffff, loopback = *n3 >> 31;
    count(0);
    // Explicit N3 device lookup avoids inheriting the inner socket's TUN/VRF.
    // Kernel routing on that device resolves direct and gateway neighbours.
    // Local delivery must install an outer IPv4 destination: plain redirect
    // retains the UE's IPv6/VRF route on the loopback receive path.
    if (loopback) {
        // The owned staging veth's dummy Ethernet destination is unicast but
        // not its peer MAC. Local delivery must not inherit PACKET_OTHERHOST.
        if (change_type(skb, 0)) return TC_ACT_SHOT;
        return redirect_neigh(n3_index, 0, 0, 0);
    }
    // The helper reads the entire union, including the IPv6-sized unused tail.
    // Volatile initialization prevents LLVM dead-store elimination of that
    // tail and avoids partially initialized stack exposure under CAP_BPF.
    volatile struct bpf_redir_neigh nh = {};
    nh.nh_family = 2; nh.ipv4_nh = cfg.next_hop;
    return redirect_neigh(n3_index, (struct bpf_redir_neigh *)&nh, sizeof(nh), 0);
}

// Classify hidden/malformed/fragmented ND into the bounded control path rather
// than delivering it to host SLAAC. The shared control parser accepts only a
// fully validated, unfragmented direct Router Advertisement.
static __attribute__((noinline)) int advertisement(struct __sk_buff *skb, __u32 offset, __u32 length, __u8 next) {
    __u32 cursor = 40;
#pragma unroll
    for (int i = 0; i < 8; i++) {
        unsigned char header[8] = {};
        if (next == 58) {
            if (cursor >= length || load_bytes(skb, offset + cursor, header, 1)) return 1;
            return header[0] == 134;
        }
        if (next == 0 || next == 43 || next == 60 || next == 51) {
            if (cursor + 2 > length || load_bytes(skb, offset + cursor, header, 2)) return 1;
            __u32 size = ((__u32)header[1] + 1) * 8;
            if (next == 51) size = ((__u32)header[1] + 2) * 4;
            if (size > length - cursor) return 1;
            next = header[0]; cursor += size;
        } else if (next == 44) {
            if (cursor + 8 > length || load_bytes(skb, offset + cursor, header, 8)) return 1;
            if (header[0] == 58) return 1;
            next = header[0]; cursor += 8;
        } else return 0;
    }
    return 1;
}

SEC("tc/ingress")
int decap(struct __sk_buff *skb) {
    struct ethhdr eth = {};
    struct iphdr outer = {};
    if (load_bytes(skb, 0, &eth, sizeof(eth)) || eth.h_proto != htons(ETH_P_IP) ||
        load_bytes(skb, ETH_HLEN, &outer, sizeof(outer)) || outer.version != 4 || outer.protocol != 17) return NEXT;
    __u32 local = outer.daddr;
    __u32 *owner = lookup(&locals, &local);
    if (!owner) return NEXT;
    // Let the kernel reassemble outer fragments. Only the exclusively bound
    // management socket can consume their completed owned UDP datagrams.
    if (outer.frag_off & htons(0x3fff)) { count(3); return NEXT; }
    __u32 ihl = outer.ihl * 4;
    if (ihl < 20 || ihl > 60) return NEXT;
    struct udphdr udp = {};
    if (load_bytes(skb, ETH_HLEN + ihl, &udp, sizeof(udp)) || udp.dest != htons(2152)) return NEXT;
    __u32 total = ntohs(outer.tot_len), length = ntohs(udp.len);
    if ((*owner & 0x7fffffff) != skb->ingress_ifindex || total < ihl + 20 || total > MAX_N3_MTU ||
        total + ETH_HLEN > skb->len || length < 20 || total != length + ihl ||
        udp.source != htons(2152) || !ip_checksum(skb, ETH_HLEN, ihl)) goto drop;
    if (udp.check) {
        __u32 src = ntohl(outer.saddr), dst = ntohl(outer.daddr);
        struct checksum_context ctx = {.skb = skb, .offset = ETH_HLEN + ihl, .length = length,
            .sum = (src >> 16) + (src & 0xffff) + (dst >> 16) + (dst & 0xffff) + 17 + length};
        if (bpf_loop(2048, udp_checksum_chunk, &ctx, 0) < 0 || ctx.failed) goto drop;
        if (fold(ctx.sum) != 0xffff) {
            // TC cannot distinguish RX CHECKSUM_PARTIAL from a corrupt wire
            // checksum. The exclusively owned UDP socket preserves kernel
            // checksum validation and reassembly before strict Go admission.
            count(3); return NEXT;
        }
    }
    __u32 gtp_offset = ETH_HLEN + ihl + 8;
    struct gtp_header gtp = {};
    if (load_bytes(skb, gtp_offset, &gtp, 8) || ntohs(gtp.length) + 8 != length - 8) goto drop;
    if (gtp.type == 1) {
        if (length < 20 || gtp.flags != 0x32 || gtp.teid ||
            load_bytes(skb, gtp_offset + 8, &gtp.sequence, 4) || gtp.npdu || gtp.next) goto drop;
        return NEXT;
    }
    if (gtp.type != 255 || length < 36) goto drop;
    struct down_key key = {local, outer.saddr, ntohl(gtp.teid)};
    __u32 *endpoint = lookup(&downlinks, &key);
    if (!endpoint) goto drop;
    // A full-width scalar slot prevents LLVM from partially overwriting a
    // spilled map pointer when the optional-header branches extend lifetimes.
    volatile __u64 endpoint_key = *endpoint;
    struct binding *value = lookup(&sessions, (void *)&endpoint_key);
    if (!value) goto drop;
    struct binding cfg = *value;
    if (cfg.endpoint != endpoint_key || cfg.local != key.local || cfg.peer != key.peer || cfg.downlink_teid != key.teid) goto drop;
    if ((gtp.flags & 0xf8) != 0x30) goto drop;
    __u32 gtplen = 8;
    if (gtp.flags == 0x34 || gtp.flags == 0x36) {
        if (length < 24 || load_bytes(skb, gtp_offset + 8, &gtp.sequence, 8)) goto drop;
        // Generic chains and optional N-PDU fields use the exclusively owned
        // UDP socket's shared bounded parser, after canonical tuple admission.
        if ((gtp.flags == 0x34 && gtp.sequence) || gtp.npdu || gtp.next != 0x85 ||
            gtp.extension_length != 1 || gtp.end || (gtp.pdu_type & 0x0f) || (gtp.qfi & 0x80)) { count(3); return NEXT; }
        if (gtp.pdu_type >> 4 != 0 || (gtp.qfi & 0x3f) != cfg.qfi) goto drop;
        gtplen = 16;
    } else if (gtp.flags == 0x32) {
        if (load_bytes(skb, gtp_offset + 8, &gtp.sequence, 4) || gtp.next) goto drop;
        if (gtp.npdu) { count(3); return NEXT; }
        gtplen = 12;
    } else if (gtp.flags != 0x30) { count(3); return NEXT; }
    __u32 inner_length = length - 8 - gtplen, inner_offset = gtp_offset + gtplen;
    if (inner_length > cfg.mtu) goto drop;
    unsigned char version = 0;
    if (load_bytes(skb, inner_offset, &version, 1)) goto drop;
    __u64 flags = 0;
    if (version >> 4 == 4) {
        struct iphdr inner = {};
        if (load_bytes(skb, inner_offset, &inner, sizeof(inner)) || inner.ihl < 5 ||
            !cfg.ue || inner.daddr != cfg.ue || ntohs(inner.tot_len) != inner_length ||
            inner.ihl * 4 > inner_length || !ip_checksum(skb, inner_offset, inner.ihl * 4)) goto drop;
    } else if (version >> 4 == 6) {
        struct ipv6hdr ip = {};
        if (!cfg.ipv6 || inner_length < 40 || load_bytes(skb, inner_offset, &ip, sizeof(ip)) ||
            ntohs(ip.payload_len) + 40 != inner_length) goto drop;
        if (advertisement(skb, inner_offset, inner_length, ip.nexthdr)) return NEXT;
        if (!ipv6_owned(&ip.daddr, &cfg)) goto drop;
        flags = BPF_F_ADJ_ROOM_DECAP_L3_IPV6;
    } else goto drop;
    if (adjust_room(skb, -((__s32)ihl + 8 + (__s32)gtplen), BPF_ADJ_ROOM_MAC, flags)) goto drop;
    if (flags) {
        __u16 protocol = htons(ETH_P_IPV6);
        if (store_bytes(skb, 12, &protocol, sizeof(protocol), 0)) goto drop;
    }
    count(1);
    return redirect(cfg.endpoint, BPF_F_INGRESS);
drop:
    count(2);
    return TC_ACT_SHOT;
}

char LICENSE[] SEC("license") = "Apache 2.0";
