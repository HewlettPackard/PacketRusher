#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Bounded, non-acceptance diagnostics for the owned core/RAN namespace."""
import argparse
import json
import socket
import subprocess
import time
import urllib.parse
from pathlib import Path


def inspect(state, label):
    state=Path(state).resolve()
    profile=json.loads((state/'profile.json').read_text())
    output={"label":label,"network":{}}
    for command in [['ip','-json','address'],['ip','-json','route'],['ip','-json','neighbour'],['ss','-ltnp']]:
        try:
            response=subprocess.run(command,capture_output=True,text=True,timeout=2)
            output['network'][' '.join(command)]={"status":response.returncode,"stdout":response.stdout,"stderr":response.stderr}
        except Exception as error:
            output['network'][' '.join(command)]={"error":str(error)}
    uri=urllib.parse.urlsplit(profile['oam'])
    response, started = bytearray(), time.monotonic()
    output['http']={"url":profile['oam'],"connected":False}
    try:
        with socket.create_connection((uri.hostname,uri.port),timeout=2) as connection:
            output['http']['connected']=True
            output['http']['connect_seconds']=time.monotonic()-started
            connection.settimeout(2)
            connection.sendall(f'GET {uri.path} HTTP/1.1\r\nHost: {uri.hostname}:{uri.port}\r\nConnection: close\r\n\r\n'.encode())
            deadline=time.monotonic()+2
            while len(response)<1024*1024 and time.monotonic()<deadline:
                data=connection.recv(min(4096,1024*1024-len(response)))
                if not data: break
                response.extend(data)
    except Exception as error:
        output['http']['error']=str(error)
    output['http']['elapsed_seconds']=time.monotonic()-started
    output['http']['bytes_received']=len(response)
    (state/f'diagnostic-{label}-http.txt').write_bytes(response)
    (state/f'diagnostic-{label}.json').write_text(json.dumps(output,indent=2)+'\n')

if __name__=='__main__':
    parser=argparse.ArgumentParser()
    parser.add_argument('--state',required=True)
    parser.add_argument('--label',required=True)
    args=parser.parse_args()
    inspect(args.state,args.label)
