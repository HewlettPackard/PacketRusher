// SPDX-License-Identifier: Apache-2.0
// PacketRusher's IPv4 GTP-U datapath. No PFCP/UPF implementation is included.
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
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
#define MAX_N3_MTU 1500
#define OVERHEAD 44

static void *(*lookup)(void *, const void *) = (void *)BPF_FUNC_map_lookup_elem;
static long (*push_encap)(struct __sk_buff *, __u32, void *, __u32) = (void *)BPF_FUNC_lwt_push_encap;
static long (*pull_data)(struct __sk_buff *, __u32) = (void *)BPF_FUNC_skb_pull_data;
static long (*adjust_room)(struct __sk_buff *, __s32, __u32, __u64) = (void *)BPF_FUNC_skb_adjust_room;
static long (*redirect)(__u32, __u64) = (void *)BPF_FUNC_redirect;

struct binding {
    __u32 local, peer, ue, uplink_teid, downlink_teid, endpoint, mtu;
    __u8 qfi, padding[3];
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
    __uint(type, BPF_MAP_TYPE_ARRAY); __uint(max_entries, 3);
    __type(key, __u32); __type(value, __u64);
} counters SEC(".maps");

INLINE void count(__u32 index) {
    __u64 *n = lookup(&counters, &index);
    if (n) __sync_fetch_and_add(n, 1);
}
INLINE __u16 fold(__u32 sum) {
    sum = (sum & 0xffff) + (sum >> 16);
    sum = (sum & 0xffff) + (sum >> 16);
    return sum;
}
INLINE int ipv4_checksum(const unsigned char *ip, void *end) {
    if ((void *)(ip + 20) > end) return 0;
    __u32 sum = 0;
#pragma unroll
    for (int i = 0; i < 10; i++) sum += ((__u32)ip[i * 2] << 8) | ip[i * 2 + 1];
    return fold(sum) == 0xffff;
}

struct gtp_header {
    __u8 flags, type; __u16 length; __u32 teid;
    __u16 sequence; __u8 npdu, next;
    __u8 extension_length, pdu_type, qfi, end;
} __attribute__((packed));
struct encap_header { struct iphdr ip; struct udphdr udp; struct gtp_header gtp; };

SEC("lwt_xmit")
int encap(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data, *end = (void *)(long)skb->data_end;
    struct iphdr *ip = data;
    if ((void *)(ip + 1) > end || *(unsigned char *)ip != 0x45) return BPF_DROP;
    __u32 source = ip->saddr;
    struct binding *value = lookup(&sessions, &source);
    if (!value) return BPF_DROP;
    struct binding cfg = *value;
    // The assigned source and endpoint determine ownership; other routes/UEs
    // cannot borrow this program's mapping. Endpoint limits prevent multi-
    // segment TCP skbs before this hook; never emit one oversized GTP length.
    if (cfg.ue != source || skb->len > cfg.mtu || skb->len < 20 ||
        ntohs(ip->tot_len) != skb->len || skb->gso_segs > 1) return BPF_DROP;
    struct encap_header h = {};
    h.ip.version = 4; h.ip.ihl = 5; h.ip.ttl = 64; h.ip.protocol = 17;
    h.ip.tot_len = htons(skb->len + OVERHEAD);
    h.ip.frag_off = htons(0x4000); // Outer fragmentation is unsupported.
    h.ip.saddr = cfg.local; h.ip.daddr = cfg.peer;
    h.udp.source = htons(2152); h.udp.dest = htons(2152);
    h.udp.len = htons(skb->len + 24); // UDP + complete GTP/PDU container.
    // IPv4 UDP checksum zero is valid; the inner checksum is preserved.
    h.gtp.flags = 0x34; h.gtp.type = 255; h.gtp.length = htons(skb->len + 8);
    h.gtp.teid = htonl(cfg.uplink_teid); h.gtp.next = 0x85;
    h.gtp.extension_length = 1; h.gtp.pdu_type = 0x10; h.gtp.qfi = cfg.qfi;
    if (push_encap(skb, BPF_LWT_ENCAP_IP, &h, sizeof(h))) return BPF_DROP;
    count(0);
    // Kernel routing supplies N3 neighbour resolution and Ethernet framing.
    return BPF_LWT_REROUTE;
}

SEC("tc/ingress")
int decap(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data, *end = (void *)(long)skb->data_end;
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > end || eth->h_proto != htons(ETH_P_IP)) return NEXT;
    struct iphdr *outer = (void *)(eth + 1);
    if ((void *)(outer + 1) > end || outer->version != 4 || outer->protocol != 17) return NEXT;
    __u32 local = outer->daddr;
    __u32 *owner = lookup(&locals, &local);
    if (!owner) return NEXT; // Unowned host addresses/ports are untouched.
    __u32 ihl = outer->ihl * 4;
    if (ihl < 20 || ihl > 60) return NEXT;
    struct udphdr *udp = (void *)outer + ihl;
    // Non-first fragments cannot identify a port; leave them to the host.
    if (outer->frag_off & htons(0x1fff)) return NEXT;
    if ((void *)(udp + 1) > end || udp->dest != htons(2152)) return NEXT;
    // From here only our exclusively reserved local UDP/2152 is affected.
    if (*owner != skb->ingress_ifindex || ihl != 20 || skb->len > MAX_N3_MTU + ETH_HLEN ||
        (outer->frag_off & htons(0x2000)) || pull_data(skb, skb->len)) goto drop;
    data = (void *)(long)skb->data; end = (void *)(long)skb->data_end;
    outer = data + ETH_HLEN; udp = (void *)(outer + 1);
    if ((void *)(udp + 1) > end || !ipv4_checksum((void *)outer, end)) goto drop;
    __u32 total = ntohs(outer->tot_len), length = ntohs(udp->len);
    if (total < 40 || total > MAX_N3_MTU || (void *)outer + total > end ||
        length < 20 || total != length + 20 || udp->source != htons(2152)) goto drop;
    // TC runs before the UDP stack: validate a nonzero checksum ourselves.
    if (udp->check) {
        __u32 src = ntohl(outer->saddr), dst = ntohl(outer->daddr);
        __u32 sum = (src >> 16) + (src & 0xffff) + (dst >> 16) + (dst & 0xffff) + 17 + length;
        unsigned char *bytes = (void *)udp;
#pragma clang loop unroll(disable)
        for (__u32 i = 0; i < 750; i++) {
            __u32 offset = i * 2;
            if (offset >= length) break;
            if ((void *)(bytes + offset + 1) > end) goto drop;
            sum += (__u32)bytes[offset] << 8;
            if (offset + 1 < length) {
                if ((void *)(bytes + offset + 2) > end) goto drop;
                sum += bytes[offset + 1];
            }
        }
        if (fold(sum) != 0xffff) goto drop;
    }
    struct gtp_header *gtp = (void *)(udp + 1);
    asm volatile("" : "+r"(gtp)); // Preserve the explicit packet bound for the verifier.
    if ((void *)gtp + 8 > end || ntohs(gtp->length) + 8 != length - 8) goto drop;
    // Keepalives alone reach the owned management socket. The Go responder also
    // verifies this peer belongs to a currently committed session.
    if (gtp->type == 1) {
        if ((void *)gtp + 12 > end || length != 20 || gtp->flags != 0x32 ||
            gtp->teid || gtp->npdu || gtp->next) goto drop;
        return NEXT;
    }
    // End Markers never release a NAS/NGAP-owned session; unsupported controls drop.
    if (gtp->type != 255 || length < 36) goto drop;
    struct down_key key = {local, outer->saddr, ntohl(gtp->teid)};
    __u32 *ue = lookup(&downlinks, &key);
    if (!ue) goto drop;
    struct binding *value = lookup(&sessions, ue);
    if (!value) goto drop;
    struct binding cfg = *value;
    // A single canonical entry is the commit point of handover. Staged/stale
    // downlink keys cannot deliver before/after that atomic map replacement.
    if (cfg.local != key.local || cfg.peer != key.peer || cfg.downlink_teid != key.teid) goto drop;
    __u32 gtplen = 8;
    // TS 29.281 §5.1: S and E independently require the four optional
    // header octets. free5UPF sends E+S, including sequence zero. Sequence
    // numbers do not change tuple/TEID ownership or permit N-PDU forwarding.
    if (gtp->flags == 0x34 || gtp->flags == 0x36) {
        if ((void *)(gtp + 1) > end || (gtp->flags == 0x34 && gtp->sequence) || gtp->npdu || gtp->next != 0x85 ||
            gtp->extension_length != 1 || gtp->pdu_type != 0 || gtp->qfi != cfg.qfi || gtp->end) goto drop;
        gtplen = 16;
    } else if (gtp->flags == 0x32) {
        if ((void *)gtp + 12 > end || gtp->npdu || gtp->next) goto drop;
        gtplen = 12;
    } else if (gtp->flags != 0x30) goto drop;
    __u32 inner_length = length - 8 - gtplen;
    struct iphdr *inner = (void *)gtp + gtplen;
    asm volatile("" : "+r"(inner));
    if ((void *)(inner + 1) > end || *(unsigned char *)inner != 0x45 ||
        inner->daddr != cfg.ue || ntohs(inner->tot_len) != inner_length ||
        inner_length > cfg.mtu || !ipv4_checksum((void *)inner, end)) goto drop;
    __u32 endpoint = cfg.endpoint;
    if (adjust_room(skb, -(20 + 8 + (__s32)gtplen), BPF_ADJ_ROOM_MAC, 0)) goto drop;
    count(1);
    return redirect(endpoint, BPF_F_INGRESS);
drop:
    count(2);
    return TC_ACT_SHOT;
}

char LICENSE[] SEC("license") = "Apache 2.0";
