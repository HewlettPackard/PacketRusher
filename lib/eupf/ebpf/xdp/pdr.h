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
#include <linux/bpf.h>
#include <linux/ipv6.h>

#include "xdp/sdf_filter.h"

#define PDR_MAP_UPLINK_SIZE 1024
#define PDR_MAP_DOWNLINK_IPV4_SIZE 1024
#define PDR_MAP_DOWNLINK_IPV6_SIZE 1024
#define FAR_MAP_SIZE 1024


enum outer_header_removal_values {
    OHR_GTP_U_UDP_IPv4 = 0,
    OHR_GTP_U_UDP_IPv6 = 1,
    OHR_UDP_IPv4 = 2,
    OHR_UDP_IPv6 = 3,
    OHR_IPv4 = 4,
    OHR_IPv6 = 5,
    OHR_GTP_U_UDP_IP = 6,
    OHR_VLAN_S_TAG = 7,
    OHR_S_TAG_C_TAG = 8,
};

// Possible optimizations: 
// 0. Store SDFs in a separate map. PDR will have only id of corresponding SDF.
// 1. Combine SrcAddress.Type and DstAddress.Type into one __u8 field. Then to retrieve and put data will be used operators & and | .
// 2. Put all fields into one big structure. Sort in specific order to reduce paddings inside structure.

struct sdf_rules {
    struct sdf_filter sdf_filter;
    __u8 outer_header_removal;
    __u32 far_id;
    __u32 qer_id;
};

struct pdr_info {
    __u32 far_id;
    __u32 qer_id;
    __u8 outer_header_removal;
    __u8 sdf_mode; // 0 - no sdf, 1 - sdf only, 2 - sdf + default
    struct sdf_rules sdf_rules;
};

/* ipv4 -> PDR */ 
struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32);
    __type(value, struct pdr_info);
    __uint(max_entries, PDR_MAP_DOWNLINK_IPV4_SIZE);
} pdr_map_downlink_ip4 SEC(".maps");

/* ipv6 -> PDR */
struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct in6_addr);
    __type(value, struct pdr_info);
    __uint(max_entries, PDR_MAP_DOWNLINK_IPV6_SIZE);
} pdr_map_downlink_ip6 SEC(".maps");


/* teid -> PDR */
struct
{
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u32);
    __type(value, struct pdr_info);
    __uint(max_entries, PDR_MAP_UPLINK_SIZE);
} pdr_map_uplink_ip4 SEC(".maps");

enum far_action_mask {
    FAR_DROP = 0x01,
    FAR_FORW = 0x02,
    FAR_BUFF = 0x04,
    FAR_NOCP = 0x08,
    FAR_DUPL = 0x10,
    FAR_IPMA = 0x20,
    FAR_IPMD = 0x40,
    FAR_DFRT = 0x80,
};

enum outer_header_creation_values {
    OHC_GTP_U_UDP_IPv4 = 0x01,
    OHC_GTP_U_UDP_IPv6 = 0x02,
    OHC_UDP_IPv4 = 0x04,
    OHC_UDP_IPv6 = 0x08,
};

struct far_info {
    __u8 action;
    __u8 outer_header_creation;
    __u32 teid;
    __u32 remoteip;
    __u32 localip;
    __u32 if_index;
    /* first octet DSCP value in the Type-of-Service, second octet shall contain the ToS/Traffic Class mask field, which shall be set to "0xFC". */
    __u16 transport_level_marking;
};

/* FAR ID -> FAR */
struct
{
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32);
    __type(value, struct far_info);
    __uint(max_entries, FAR_MAP_SIZE);
} far_map SEC(".maps");
