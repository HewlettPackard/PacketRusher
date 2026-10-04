/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

// GTP-U fast path for PacketRusher UE tunnels. The packets it does not handle are
// left to the userspace backend, and first to what else is attached to the interface:
// TC_ACT_UNSPEC, where TC_ACT_OK would end there.
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/in.h>
#include <linux/in6.h>
#include <linux/ip.h>
#include <linux/pkt_cls.h>
#include <linux/udp.h>
#include <stddef.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#define GTPU_PORT 2152
#define GTPU_GPDU 255
#define GTPU_PSC 0x85 // PDU Session Container extension header, TS 38.415

// TS 29.281 §5.1 header, followed by the fields present with a PDU Session Container.
struct gtpu {
	__u8 flags, type;
	__be16 length;
	__be32 teid;
	__be16 sequence;
	__u8 npdu, next;
	__u8 ext_len, pdu_type, qfi, ext_next;
} __attribute__((packed));

struct outer {
	struct iphdr ip;
	struct udphdr udp;
	struct gtpu gtp;
} __attribute__((packed));

struct uplink {
	__be32 local, peer; // N3 addresses of the gNB and the UPF
	__be32 teid;
	__u32 n3; // interface towards the UPF
	__be32 ue; // the UE's IPv4 address
	__u8 prefix[8]; // and its IPv6 /64, zero until the UPF advertised it
	__u8 qfi;
};

struct downlink_key {
	__be32 local;
	__be32 teid;
};

// Uplink session by TUN interface.
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 65536);
	__type(key, __u32);
	__type(value, struct uplink);
} uplinks SEC(".maps");

// TUN interface by downlink tunnel.
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 65536);
	__type(key, struct downlink_key);
	__type(value, __u32);
} downlinks SEC(".maps");

// A veth without checksum offload: the kernel splits the large segments of the
// UE's TCP stack and completes their checksums when it transmits through it.
volatile __u32 stage;

// Egress of a UE TUN: hand the UE's packet over to the stage, as an Ethernet frame.
// What else the kernel sends there, such as its own Router Solicitations, is left
// to the userspace backend.
SEC("tc")
int uplink(struct __sk_buff *skb)
{
	__u32 tun = skb->ifindex;
	struct uplink *up = bpf_map_lookup_elem(&uplinks, &tun);
	__u8 source[8];
	if (!up)
		return TC_ACT_UNSPEC;
	if (skb->protocol == bpf_htons(ETH_P_IP)) {
		if (bpf_skb_load_bytes(skb, offsetof(struct iphdr, saddr), source, 4) ||
		    __builtin_memcmp(source, &up->ue, 4))
			return TC_ACT_UNSPEC;
	} else if (!up->prefix[0] || bpf_skb_load_bytes(skb, 8, source, sizeof(source)) ||
		   __builtin_memcmp(source, up->prefix, sizeof(source))) {
		return TC_ACT_UNSPEC;
	}

	struct ethhdr eth = { .h_proto = skb->protocol };
	if (bpf_skb_change_head(skb, sizeof(eth), 0) || bpf_skb_store_bytes(skb, 0, &eth, sizeof(eth), 0))
		return TC_ACT_SHOT;
	skb->mark = skb->ifindex;
	return bpf_redirect(stage, 0);
}

// Ingress of the stage peer: encapsulate and route the packet out of N3.
SEC("tc")
int encap(struct __sk_buff *skb)
{
	__u32 tun = skb->mark, inner = skb->len - sizeof(struct ethhdr);
	struct uplink *up = bpf_map_lookup_elem(&uplinks, &tun);
	if (!up)
		return TC_ACT_SHOT;

	__u32 gtp = up->qfi ? sizeof(struct gtpu) : 8, size = sizeof(struct iphdr) + sizeof(struct udphdr) + gtp;
	struct outer h = {
		.ip = {
			.version = 4, .ihl = 5, .ttl = 64, .protocol = IPPROTO_UDP,
			.tot_len = bpf_htons(inner + size), .frag_off = bpf_htons(0x4000),
			.saddr = up->local, .daddr = up->peer,
		},
		.udp = {
			.source = bpf_htons(GTPU_PORT), .dest = bpf_htons(GTPU_PORT),
			.len = bpf_htons(inner + sizeof(struct udphdr) + gtp),
		},
		.gtp = {
			.flags = up->qfi ? 0x34 : 0x30, .type = GTPU_GPDU,
			.length = bpf_htons(inner + gtp - 8), .teid = up->teid,
			.next = GTPU_PSC, .ext_len = 1, .pdu_type = 0x10, .qfi = up->qfi,
		},
	};
	__u32 sum = bpf_csum_diff(0, 0, (void *)&h, sizeof(h.ip), 0);
	sum = (sum & 0xffff) + (sum >> 16);
	h.ip.check = ~((sum & 0xffff) + (sum >> 16));

	if (bpf_skb_adjust_room(skb, size, BPF_ADJ_ROOM_MAC,
				BPF_F_ADJ_ROOM_ENCAP_L3_IPV4 | BPF_F_ADJ_ROOM_ENCAP_L4_UDP) ||
	    bpf_skb_store_bytes(skb, sizeof(struct ethhdr), &h, up->qfi ? sizeof(h) : sizeof(h) - 8, 0))
		return TC_ACT_SHOT;
	// The stage received this frame as one for another host: a loopback N3 hands it
	// to the UPF as it is, and the UPF's host would drop it as such.
	bpf_skb_change_type(skb, PACKET_HOST);
	skb->mark = 0;
	return bpf_redirect_neigh(up->n3, NULL, 0, 0);
}

// Ingress of N3: decapsulate G-PDUs of known tunnels into the UE's TUN.
SEC("tc")
int decap(struct __sk_buff *skb)
{
	struct {
		struct ethhdr eth;
		struct outer h;
		__u8 inner[41]; // up to the ICMPv6 type of an IPv6 packet
	} __attribute__((packed)) p;
	if (skb->protocol != bpf_htons(ETH_P_IP) || skb->gso_segs > 1 ||
	    bpf_skb_load_bytes(skb, 0, &p, sizeof(p.eth) + sizeof(p.h) - 8 + 20))
		return TC_ACT_UNSPEC;
	struct outer *h = &p.h;
	if (h->ip.ihl != 5 || h->ip.protocol != IPPROTO_UDP || (h->ip.frag_off & bpf_htons(0x3fff)) ||
	    h->udp.dest != bpf_htons(GTPU_PORT) || h->gtp.type != GTPU_GPDU)
		return TC_ACT_UNSPEC;

	// Version 1, GTP. With any of the E, S and PN flags the three optional fields are
	// present; the only extension header handled here is a PDU Session Container.
	__u32 gtp;
	if ((h->gtp.flags & 0xf0) != 0x30)
		return TC_ACT_UNSPEC;
	if (!(h->gtp.flags & 0x07))
		gtp = 8;
	else if (!(h->gtp.flags & 0x04))
		gtp = 12;
	else if (h->gtp.next == GTPU_PSC && h->gtp.ext_len == 1 && !h->gtp.ext_next)
		gtp = sizeof(struct gtpu);
	else
		return TC_ACT_UNSPEC;
	struct downlink_key key = { .local = h->ip.daddr, .teid = h->gtp.teid };
	__u32 *tun = bpf_map_lookup_elem(&downlinks, &key);
	if (!tun)
		return TC_ACT_UNSPEC;

	__u32 size = sizeof(struct iphdr) + sizeof(struct udphdr) + gtp;
	__u64 flags = BPF_F_ADJ_ROOM_DECAP_L3_IPV4;
	if (bpf_skb_load_bytes(skb, sizeof(p.eth) + size, p.inner, 1))
		return TC_ACT_UNSPEC;
	if (p.inner[0] >> 4 == 6) {
		// Router Advertisements configure the UE's prefix in the userspace backend.
		if (bpf_skb_load_bytes(skb, sizeof(p.eth) + size, p.inner, sizeof(p.inner)) ||
		    (p.inner[6] == IPPROTO_ICMPV6 && p.inner[40] == 134))
			return TC_ACT_UNSPEC;
		flags = BPF_F_ADJ_ROOM_DECAP_L3_IPV6;
	}
	if (bpf_skb_adjust_room(skb, -(__s32)size, BPF_ADJ_ROOM_MAC, flags))
		return TC_ACT_UNSPEC;
	return bpf_redirect(*tun, BPF_F_INGRESS);
}

char _license[] SEC("license") = "Apache-2.0";
