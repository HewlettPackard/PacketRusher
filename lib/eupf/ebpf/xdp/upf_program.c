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

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include "xdp/program_array.h"

SEC("xdp/upf")
int upf_func(struct xdp_md *ctx) {
    bpf_printk("upf_program start\n");

    bpf_printk("tail call to UPF_PROG_TYPE_QER key\n");
    bpf_tail_call(ctx, &upf_pipeline, UPF_PROG_TYPE_QER);
    bpf_printk("tail call to UPF_PROG_TYPE_QER key failed\n");
    return XDP_ABORTED;
}

char _license[] SEC("license") = "GPL";