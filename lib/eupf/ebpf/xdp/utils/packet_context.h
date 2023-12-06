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

#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/types.h>
#include <linux/udp.h>
#include <linux/tcp.h>
#include "xdp/utils/gtpu.h"

/* Header cursor to keep track of current parsing position */
struct packet_context {
    char *data;
    const char *data_end;
    struct upf_counters *counters;
    struct n3_n6_counters *n3_n6_counter;
    struct xdp_md *xdp_ctx;
    struct ethhdr *eth;
    struct iphdr *ip4;
    struct ipv6hdr *ip6;
    struct udphdr *udp;
    struct tcphdr *tcp;
    struct gtpuhdr *gtp;
};
