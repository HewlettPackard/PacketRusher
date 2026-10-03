#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
import argparse
import json
from pathlib import Path
from prepare import HERE, CORE_IP, RAN_IP, FREE_IPS, generate, write

def compose(core, state):
    state = Path(state).resolve()
    generate(core, state)
    versions = json.loads((HERE/'versions.json').read_text())
    repo = HERE.parent.parent
    mount = str(state)+':/artifacts'
    images = versions['images']
    services = {
        'db': {'image': images['library/mongo:8.0.16'], 'command': ['mongod','--bind_ip_all','--nounixsocket','--wiredTigerCacheSizeGB','0.25'], 'networks': {'core': {'ipv4_address':'172.30.5.10'}}, 'volumes': ['db:/data/db', mount+':ro'], 'healthcheck': {'test':['CMD','mongosh','--quiet','--eval','quit(db.adminCommand({ping:1}).ok ? 0 : 1)'], 'interval':'2s','timeout':'3s','retries':30}},
        'ran': {'build': {'context':str(repo),'dockerfile':'test/external-cores/ran.Dockerfile'}, 'networks': {'core':{'ipv4_address':RAN_IP}}, 'volumes':[mount], 'command':['python3','/probe/probe.py','--state','/artifacts'], 'init':True},
    }
    if core == 'free5gc':
        for nf, ip in FREE_IPS.items():
            services[nf] = {'image':images[f'free5gc/{nf}:v4.3.0'], 'entrypoint':[f'./{nf}'], 'command':['-c',f'/free5gc/config/{nf}cfg.yaml'], 'volumes':[str(state/'config')+':/free5gc/config:ro'], 'networks': {'core':{'ipv4_address':ip}}, 'depends_on': {'db':{'condition':'service_healthy'}}, 'init':True}
    else:
        services['core'] = {'build':{'context':str(repo),'dockerfile':'test/external-cores/open5gs.Dockerfile'}, 'networks':{'core':{'ipv4_address':CORE_IP}}, 'volumes':[mount], 'devices':['/dev/net/tun:/dev/net/tun'], 'cap_add':['NET_ADMIN'], 'command':['python3','/probe/core.py','--prefix','/opt/open5gs','--state','/artifacts'], 'depends_on':{'db':{'condition':'service_healthy'}}, 'init':True}
        services['ran'].update({'devices':['/dev/net/tun:/dev/net/tun'], 'cap_add':['NET_ADMIN']})
    write(state/'compose.json',{'services':services,'networks':{'core':{'ipam':{'config':[{'subnet':'172.30.5.0/24'}]}}},'volumes':{'db':{}}})

if __name__=='__main__':
    parser=argparse.ArgumentParser()
    parser.add_argument('--core',choices=['free5gc','open5gs'],required=True)
    parser.add_argument('--state',required=True)
    args=parser.parse_args()
    compose(args.core,args.state)
