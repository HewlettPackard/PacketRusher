#!/bin/sh
# SPDX-License-Identifier: Apache-2.0
set -eu
cd "$(dirname "$0")/.."
compiler=${CLANG:-clang-18}
system_include=${BPF_SYSTEM_INCLUDE:-/usr/include/$(uname -m)-linux-gnu}
"$compiler" -target bpfel -mcpu=v3 -O2 -g -Wall -Werror \
  -fdebug-prefix-map="$PWD"=. -I"$system_include" \
  -c internal/control_test_engine/ue/gtp/ebpfgtp/bpf/gtpu.c \
  -o internal/control_test_engine/ue/gtp/ebpfgtp/gtpu_bpfel.o
