#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Use existing exact core binaries in owned namespaces, with a fresh database."""
import argparse
import json
import os
import signal
import subprocess
import time
from pathlib import Path
from prepare import HERE, CORE_IP, RAN_IP, generate
from probe import until

def launch_owned(command, **kwargs):
    # An owned session keeps timeout cleanup independent of the invoking shell.
    return subprocess.Popen(command, start_new_session=True, **kwargs)


def stop_owned(process, timeout=12):
    # The leader may have exited while a grandchild still owns the namespace.
    try: os.killpg(process.pid, signal.SIGTERM)
    except ProcessLookupError: pass
    try: process.wait(timeout=timeout)
    except subprocess.TimeoutExpired: pass
    finally:
        try: os.killpg(process.pid, signal.SIGKILL)
        except ProcessLookupError: pass
        process.wait()


def interrupt(signum, _):
    raise InterruptedError(f"native runner interrupted by signal {signum}")


def run(core, prefix, binary, state):
    if os.geteuid() != 0 or os.stat('/proc/self/ns/net').st_ino == os.stat('/proc/1/ns/net').st_ino:
        raise RuntimeError("native runner requires root in a new network namespace; use native.sh")
    if core == "open5gs" and not Path('/dev/net/tun').is_char_device():
        raise RuntimeError("isolated /dev/net/tun missing; use native.sh")
    state = Path(state).resolve()
    generate(core, state, native=True, prefix=prefix)
    processes, files = [], []
    probe_process = None
    for signum in (signal.SIGINT, signal.SIGTERM):
        signal.signal(signum, interrupt)
    link_owned = False
    try:
        subprocess.run(['ip','link','set','lo','up'],check=True)
        child = launch_owned(['unshare','--net','sleep','3600'])
        processes.append(child)
        until(lambda: child.poll() is None and os.stat(f'/proc/{child.pid}/ns/net').st_ino != os.stat('/proc/self/ns/net').st_ino, 5, 'child RAN namespace')
        subprocess.run(['ip','link','add','core0','type','veth','peer','name','ran0'],check=True)
        link_owned = True
        subprocess.run(['ip','link','set','ran0','netns',str(child.pid)],check=True)
        subprocess.run(['ip','addr','add',CORE_IP+'/24','dev','core0'],check=True)
        subprocess.run(['ip','link','set','core0','up'],check=True)
        net = ['nsenter','--target',str(child.pid),'--net']
        for command in [['ip','link','set','lo','up'],['ip','link','set','ran0','name','eth0'],['ip','addr','add',RAN_IP+'/24','dev','eth0'],['ip','link','set','eth0','up']]:
            subprocess.run(net+command,check=True)
        db = state/'mongo'
        db.mkdir()
        mongo_log = (state/'mongo-stdout.log').open('wb')
        files.append(mongo_log)
        mongo = launch_owned(['mongod','--dbpath',str(db),'--bind_ip','127.0.0.1','--port','27017','--nounixsocket','--wiredTigerCacheSizeGB','0.25'],stdout=mongo_log,stderr=subprocess.STDOUT)
        processes.append(mongo)
        def ping():
            if mongo.poll() is not None:
                raise RuntimeError('owned MongoDB exited')
            return subprocess.run(['mongosh','--quiet','mongodb://127.0.0.1:27017/'+core,'--eval','db.adminCommand({ping:1}).ok'],capture_output=True,timeout=3).returncode == 0
        until(ping, 30, 'private MongoDB')
        subprocess.run(['mongosh','--quiet','mongodb://127.0.0.1:27017/'+core,str(state/'subscriber.js')],check=True,timeout=10)
        core_log = (state/'core-supervisor.log').open('wb')
        files.append(core_log)
        supervisor = launch_owned(['python3',str(HERE/'core.py'),'--prefix',str(Path(prefix).resolve()),'--state',str(state)],stdout=core_log,stderr=subprocess.STDOUT)
        processes.append(supervisor)
        def ready():
            if supervisor.poll() is not None:
                raise RuntimeError('real core startup failed; inspect owned core logs')
            return (state/'core-ready').exists()
        until(ready, 65, 'real core listeners')
        # Keep both sides' live TCP exchange, not just a post-shutdown timeout.
        diagnostic_log = (state/'diagnostic-capture.log').open('wb')
        files.append(diagnostic_log)
        diagnostic_capture = launch_owned(['tcpdump','-n','-U','-i','any','-w',str(state/'oam-diagnostic.pcap'),'tcp','port','9090' if core == 'open5gs' else '8000'],stdout=diagnostic_log,stderr=subprocess.STDOUT)
        processes.append(diagnostic_capture)
        for namespace, label in [([], 'core-startup'), (net, 'ran-startup')]:
            subprocess.run(namespace+['python3',str(HERE/'diagnostics.py'),'--state',str(state),'--label',label],check=False,timeout=15)
        probe_process = launch_owned(net+['python3',str(HERE/'probe.py'),'--packetrusher',str(Path(binary).resolve()),'--state',str(state)])
        status = probe_process.wait(timeout=180)
        if status != 0:
            raise subprocess.CalledProcessError(status, probe_process.args)
    finally:
        # Stop the UE/probe before retiring its endpoint's network namespace.
        if probe_process:
            stop_owned(probe_process)
        if link_owned:
            # Retire our veth while its owning child namespace still exists.
            subprocess.run(['ip','link','del','core0'],check=False)
        for process in reversed(processes):
            stop_owned(process)
        for file in files:
            file.close()

if __name__ == '__main__':
    parser=argparse.ArgumentParser()
    parser.add_argument('--core',choices=['free5gc','open5gs'],required=True)
    parser.add_argument('--prefix',required=True)
    parser.add_argument('--packetrusher',required=True)
    parser.add_argument('--state',required=True)
    args=parser.parse_args()
    run(args.core,args.prefix,args.packetrusher,args.state)
