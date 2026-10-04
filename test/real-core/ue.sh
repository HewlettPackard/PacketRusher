#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# © Copyright 2026 Valentin D'Emmanuele
# Registers one UE against a running 5G core, pings through its PDU session when
# an address is given, then deregisters and checks PacketRusher's report.
# usage: ue.sh <packetrusher> <config.yml> [<address reachable through the UPF>]
set -euo pipefail
bin=$1 config=$2 dn=${3:-}
dir=$(mktemp -d)
flags=(--numPduSessions 0)
[[ -n $dn ]] && flags=(--numPduSessions 1 --tunnel --tunnel-vrf=false)

"$bin" --config "$config" ${BACKEND:+--tunnel-backend "$BACKEND"} --report-json "$dir/report.json" multi-ue -n 1 --control-socket "$dir/control" "${flags[@]}" &
trap 'kill $! 2>/dev/null || true' EXIT
timeout 10 bash -c "until [[ -S $dir/control ]]; do sleep 0.1; done"
control() { "$bin" control --socket "$dir/control" --ue 1 --timeout 30s --action "$1"; }

control wait
if [[ -n $dn ]]; then
	ue=$(ip -4 -o addr show | awk '$2 ~ /^val/ { sub("/.*", "", $4); print $4 }')
	ping -c 3 -W 2 -I "$ue" "$dn"
fi
control deregister
kill -INT $!
wait $! || true
trap - EXIT

jq -e --argjson sessions "${flags[1]}" '
	[.procedures[] | select(.failure != 0 or .pending != 0)] == [] and
	(.procedures[] | select(.procedure == "registration") | .success) == 1 and
	([.procedures[] | select(.procedure == "pdu_session_establishment") | .success] | add // 0) == $sessions
' "$dir/report.json" >/dev/null || { cat "$dir/report.json"; exit 1; }
echo "UE registered, ${dn:+passed traffic, }deregistered"
