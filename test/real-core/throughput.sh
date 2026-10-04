#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# © Copyright 2026 Valentin D'Emmanuele
# Measures the throughput of the user plane against a running 5G core: every UE runs
# iperf3 through its PDU session at the same time, TCP then UDP, uplink then downlink.
# usage: [UES=4] [BACKEND=ebpf] [DEDICATED=1] throughput.sh <packetrusher> <config.yml> <iperf3 server>
# The core knows the subscribers following the configured MSIN, and one iperf3 server
# per UE listens behind the UPF: for port in $(seq 5201 5204); do iperf3 -s -p $port -D; done
set -euo pipefail
bin=$1 config=$2 server=$3 ues=${UES:-4}
dir=$(mktemp -d)

"$bin" --config "$config" ${BACKEND:+--tunnel-backend "$BACKEND"} multi-ue -n "$ues" --control-socket "$dir/control" \
	--tunnel --tunnel-vrf=false ${DEDICATED:+--dedicatedGnb} > "$dir/log" &
packetrusher=$!
trap 'kill $packetrusher 2>/dev/null || true' EXIT
timeout 10 bash -c "until [[ -S $dir/control ]]; do sleep 0.1; done"
for ue in $(seq "$ues"); do
	"$bin" control --socket "$dir/control" --ue "$ue" --timeout 60s --action wait > /dev/null
done
mapfile -t addresses < <(ip -4 -o addr show | awk '$2 ~ /^val/ { sub("/.*", "", $4); print $4 }')

measure() { # <label> [<iperf3 flag>...]
	local i clients=()
	for i in "${!addresses[@]}"; do
		iperf3 -c "$server" -p $((5201 + i)) -B "${addresses[$i]}" -t 5 -f m "${@:2}" > "$dir/iperf.$i" &
		clients+=($!)
	done
	wait "${clients[@]}" || true
	grep -h receiver "$dir"/iperf.* | grep -oE '[0-9.]+ Mbits/sec' |
		awk -v label="$1" '{ total += $1 } END { printf "%s: %d Mbit/s over %d UEs\n", label, total, NR }'
}
measure "TCP uplink"
measure "TCP downlink" -R
measure "UDP uplink" -u -b 0 -l 1300 # datagrams that fit the tunnel MTU of the usual UPFs
measure "UDP downlink" -u -b 0 -l 1300 -R
kill -INT $packetrusher
wait $packetrusher || true
