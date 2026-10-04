#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Read only the exact PacketRusher process's owned TCX diagnostic map."""
import json
import re
import subprocess
import time
from pathlib import Path

NAMES={0:'uplink_encap_attempts',1:'downlink_decap_redirect_attempts',2:'ingress_drops',3:'downlink_delegation_attempts',4:'oversized_uplink_delegation_attempts'}


def decode_bytes(value, width):
    if not isinstance(value,list) or len(value)!=width:
        raise AssertionError(f'counter field must contain exactly{width} raw bytes')
    raw=bytes(int(item,0) if isinstance(item,str) else item for item in value)
    return int.from_bytes(raw,'little')


def decode_dump(rows):
    if not isinstance(rows,list):
        raise AssertionError('counter dump must be an array')
    values={}
    for row in rows:
        key=decode_bytes(row['key'],4)
        if key not in NAMES or key in values:
            raise AssertionError('unexpected/duplicate counter key')
        values[key]=decode_bytes(row['value'],8)
    if set(values)!=set(NAMES):
        raise AssertionError('incomplete TCX counter ABI')
    return {NAMES[key]:values[key] for key in NAMES}


def delta(before,after):
    if any(before[name]!=after[name] for name in ('map_id','owner_pid','owner_start_ticks')):
        raise AssertionError('TCX counter ownership changed during flow')
    result={name:after['values'][name]-before['values'][name] for name in NAMES.values()}
    if any(value<0 for value in result.values()):
        raise AssertionError('TCX counters reset during flow')
    return result


class OwnedCounters:
    def __init__(self,pid,proc='/proc',runner=subprocess.check_output):
        self.pid=pid
        self.fdinfo=Path(proc)/str(pid)/'fdinfo'
        self.start_ticks=self.owner_start()
        self.runner=runner
        candidates=[]
        for map_id in self.owned_ids():
            information=self.command('show',map_id)
            if isinstance(information,list):
                if len(information)!=1:
                    raise AssertionError('ambiguous bpftool map metadata')
                information=information[0]
            if information.get('name')=='counters':
                if information.get('id')!=map_id or information.get('type')!='array' or information.get('bytes_key')!=4 or information.get('bytes_value')!=8 or information.get('max_entries')!=5:
                    raise AssertionError('unexpected owned TCX counter-map ABI')
                candidates.append((map_id,information))
        if len(candidates)!=1:
            raise AssertionError(f'expected one exact-process TCX counters map, found{len(candidates)}')
        self.map_id,self.information=candidates[0]

    def owner_start(self):
        value=(self.fdinfo.parent/'stat').read_text()
        return int(value[value.rfind(')')+2:].split()[19])

    def owned_ids(self):
        result=set()
        for path in self.fdinfo.iterdir():
            try:
                value=path.read_text()
            except FileNotFoundError:
                continue
            match=re.search(r'^map_id:\s*(\d+)\s*$',value,re.M)
            if match:
                result.add(int(match.group(1)))
        return result

    def command(self,action,map_id):
        return json.loads(self.runner(['bpftool','-j','map',action,'id',str(map_id)],text=True,stderr=subprocess.PIPE,timeout=3))

    def snapshot(self):
        if self.owner_start()!=self.start_ticks or self.map_id not in self.owned_ids():
            raise AssertionError('PacketRusher no longer owns the recorded TCX map')
        raw=self.command('dump',self.map_id)
        return {'owner_pid':self.pid,'owner_start_ticks':self.start_ticks,'map_id':self.map_id,'monotonic':time.monotonic(),'information':self.information,'raw':raw,'values':decode_dump(raw)}
