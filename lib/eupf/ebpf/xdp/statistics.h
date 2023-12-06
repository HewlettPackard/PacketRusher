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

struct upf_counters {
    __u64 rx_arp;
    __u64 rx_icmp;
    __u64 rx_icmp6;
    __u64 rx_ip4;
    __u64 rx_ip6;
    __u64 rx_tcp;
    __u64 rx_udp;
    __u64 rx_other;
    __u64 rx_gtp_echo;
    __u64 rx_gtp_pdu;
    __u64 rx_gtp_other;
    __u64 rx_gtp_unexp;
};

struct n3_n6_counters {
    __u64 rx_n3;
    __u64 tx_n3;
    __u64 rx_n6;
    __u64 tx_n6;
};

#define EUPF_MAX_XDP_ACTION 8
#define EUPF_MAX_XDP_ACTION_MASK 0x07

struct upf_statistic {
    struct upf_counters upf_counters;
    struct n3_n6_counters upf_n3_n6_counter;
    __u64 xdp_actions[EUPF_MAX_XDP_ACTION];
};

struct
{
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, __u32);
    __type(value, struct upf_statistic);
    __uint(max_entries, 1);
} upf_ext_stat SEC(".maps");


