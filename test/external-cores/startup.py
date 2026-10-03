#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Observe real NF registration and accepted PFCP packets before the UE starts."""
import json
import socket
import struct
import urllib.request
from pathlib import Path
from prepare import CORE_IP

HTTP = urllib.request.build_opener(urllib.request.ProxyHandler({}))


def free_registered(profile):
    evidence = {}
    nrf = profile["nf_addresses"]["nrf"]
    for name in profile["nf_addresses"]:
        if name == "nrf":
            continue
        kind = name.upper()
        url = f"http://{nrf}:8000/nnrf-disc/v1/nf-instances?target-nf-type={kind}&requester-nf-type=AMF"
        with HTTP.open(url, timeout=1) as response:
            profiles = json.load(response).get("nfInstances", [])
        ready = [nf for nf in profiles if nf.get("nfType") == kind and nf.get("nfStatus") == "REGISTERED"]
        if not ready:
            return None
        evidence[name] = [nf["nfInstanceId"] for nf in ready]
    return evidence


def accepted_pfcp(path):
    """Read a live Ethernet/loopback PCAP; ignore its unfinished final record."""
    try:
        data = Path(path).read_bytes()
    except FileNotFoundError:
        return None
    if len(data) < 24:
        return None
    endian = {b"\xd4\xc3\xb2\xa1": "<", b"\xa1\xb2\xc3\xd4": ">"}.get(data[:4])
    if endian is None or struct.unpack(endian+"I", data[20:24])[0] != 1:
        raise ValueError("PFCP startup capture must use Ethernet/loopback PCAP")
    offset = 24
    while offset + 16 <= len(data):
        size = struct.unpack(endian+"IIII", data[offset:offset+16])[2]
        offset += 16
        if size > 65535:
            raise ValueError("oversized PFCP capture record")
        if offset + size > len(data):
            break
        frame, offset = data[offset:offset+size], offset+size
        if len(frame) < 42 or frame[12:14] != b"\x08\x00":
            continue
        ip = frame[14:]
        ihl, length = (ip[0]&15)*4, struct.unpack("!H", ip[2:4])[0]
        if ip[0]>>4 != 4 or ihl < 20 or ip[9] != 17 or length > len(ip) or length < ihl+16:
            continue
        peers = (socket.inet_ntoa(ip[12:16]), socket.inet_ntoa(ip[16:20]))
        if peers not in [(CORE_IP, "127.0.0.4"), ("127.0.0.4", CORE_IP)]:
            continue
        udp = ip[ihl:length]
        if struct.unpack("!HH", udp[:4]) != (8805,8805):
            continue
        udp_length = struct.unpack("!H", udp[4:6])[0]
        if udp_length < 16 or udp_length > len(udp):
            continue
        pfcp = udp[8:udp_length]
        # PFCPv1 Association Setup Response has no session SEID (S=0).
        if pfcp[0]>>5 != 1 or pfcp[0]&1 or pfcp[1] != 6 or len(pfcp) != 4+struct.unpack("!H",pfcp[2:4])[0]:
            continue
        position = 8
        while position+4 <= len(pfcp):
            kind, count = struct.unpack("!HH", pfcp[position:position+4])
            position += 4
            if position+count > len(pfcp):
                break
            if kind == 19 and count == 1 and pfcp[position] == 1:
                return {"source":peers[0],"destination":peers[1],"sequence":int.from_bytes(pfcp[4:7],"big"),"cause":1}
            position += count
    return None


def open_registered_and_associated(state):
    state = Path(state)
    for name in ["udr","udm","ausf","pcf","nssf","smf","amf"]:
        path = state/"core"/(name+".log")
        if not path.exists() or "NF registered [Heartbeat:" not in path.read_text(errors="replace"):
            return None
    for name in ["smf","upf"]:
        path = state/"core"/(name+".log")
        if not path.exists():
            return None
        events = [line for line in path.read_text(errors="replace").splitlines() if "PFCP associated " in line or "PFCP de-associated " in line]
        if not events or "PFCP associated " not in events[-1]:
            return None
    return accepted_pfcp(state/"pfcp-startup.pcap")
