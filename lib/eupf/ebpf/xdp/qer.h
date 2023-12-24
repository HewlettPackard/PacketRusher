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

enum gate_status {
    GATE_STATUS_OPEN = 0,
    GATE_STATUS_CLOSED = 1,
    GATE_STATUS_RESERVED1 = 2,
    GATE_STATUS_RESERVED2 = 3,
};

struct qer_info {
    __u8 ul_gate_status;
    __u8 dl_gate_status;
    __u8 qfi;
    __u32 ul_maximum_bitrate;
    __u32 dl_maximum_bitrate;
    __u64 ul_start;
    __u64 dl_start;
};

#define QER_MAP_SIZE 1024

/* QER ID -> QER */
struct
{
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32);
    __type(value, struct qer_info);
    __uint(max_entries, QER_MAP_SIZE);
} qer_map SEC(".maps");

static __always_inline enum xdp_action limit_rate_simple(struct xdp_md *ctx, __u64 *end, const __u64 rate) {
    static const __u64 NSEC_PER_SEC = 1000000000ULL;

    /* Currently 0 rate means that traffic rate is not limited */
    if (rate == 0)
        return XDP_PASS;
        
    __u64 now = bpf_ktime_get_ns();
    if (now > *end) {
        __u64 tx_time = (ctx->data_end - ctx->data) * 8 * NSEC_PER_SEC / rate;
        *end = now + tx_time;
        return XDP_PASS;
    }

    return XDP_DROP;
}

static __always_inline enum xdp_action limit_rate_sliding_window(const __u64 packet_size, __u64 *windows_start, const __u64 rate) {
    static const __u64 NSEC_PER_SEC = 1000000000ULL;
    static const __u64 window_size = 5000000ULL;

    /* Currently 0 rate means that traffic rate is not limited */
    if (rate == 0)
        return XDP_PASS;

    __u64 tx_time = packet_size * 8 * NSEC_PER_SEC / rate;
    __u64 now = bpf_ktime_get_ns();

    __u64 start = *(volatile __u64 *)windows_start;
    if (start + tx_time > now)
        return XDP_DROP;

    if (start + window_size < now) {
        *(volatile __u64 *)&windows_start = now - window_size + tx_time;
        return XDP_PASS;
    }

    *(volatile __u64 *)&windows_start = start + tx_time;
    //__sync_fetch_and_add(&window->start, tx_time);
    return XDP_PASS;
}
