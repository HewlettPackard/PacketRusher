#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Use existing exact core binaries in owned namespaces, with a fresh database."""
import argparse
import json
import os
import signal
import subprocess
import time
from contextlib import contextmanager
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


@contextmanager
def defer_cancellation():
    # The original cancellation still propagates after finally. Further
    # signals cannot strand the remaining owned cohorts during bounded reaping.
    handlers = {signum:signal.getsignal(signum) for signum in (signal.SIGINT,signal.SIGTERM)}
    for signum in handlers:
        signal.signal(signum,signal.SIG_IGN)
    try:
        yield
    finally:
        for signum,handler in handlers.items():
            signal.signal(signum,handler)


def run(core, prefix, binary, state, sessions=None, backend="userspace", upf=None, upf_prefix=None, pdu_type="IPv4"):
    if os.geteuid() != 0 or os.stat('/proc/self/ns/net').st_ino == os.stat('/proc/1/ns/net').st_ino:
        raise RuntimeError("native runner requires root in a new network namespace; use native.sh")
    if (core == "open5gs" or sessions) and not Path('/dev/net/tun').is_char_device():
        raise RuntimeError("isolated /dev/net/tun missing; use native.sh")
    state = Path(state).resolve()
    generate(core, state, native=True, prefix=prefix, sessions=sessions, backend=backend, upf=upf, pdu_type=pdu_type)
    profile = json.loads((state/'profile.json').read_text())
    if (profile['upf_implementation'] == 'free5gc' or backend == 'gtp5g') and not Path('/sys/module/gtp5g').is_dir():
        raise RuntimeError('the genuine free5GC UPF or selected kernel backend requires gtp5g in the disposable guest; PacketRusher eBPF does not replace the UPF')
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
        if backend == 'ebpf':
            # Only these newly owned veth endpoints are changed. Partial veth
            # checksum/GSO frames are outside the backend's Ethernet profile.
            for namespace, device in [([], 'core0'), (net, 'eth0')]:
                subprocess.run(namespace+['ethtool','-K',device,'tx','off','rx','off','tso','off','gso','off','gro','off'],check=True)
                with (state/f'offloads-{device}.txt').open('w') as output:
                    subprocess.run(namespace+['ethtool','-k',device],check=True,stdout=output)
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
        core_command = ['python3',str(HERE/'core.py'),'--prefix',str(Path(prefix).resolve()),'--state',str(state)]
        if upf_prefix:
            core_command += ['--upf-prefix',str(Path(upf_prefix).resolve())]
        supervisor = launch_owned(core_command,stdout=core_log,stderr=subprocess.STDOUT)
        processes.append(supervisor)
        def ready():
            if supervisor.poll() is not None:
                raise RuntimeError('real core startup failed; inspect owned core logs')
            return (state/'core-ready').exists()
        until(ready, 65, 'real core listeners')
        # Keep both sides' live TCP exchange, not just a post-shutdown timeout.
        diagnostic_log = (state/'diagnostic-capture.log').open('wb')
        files.append(diagnostic_log)
        diagnostic_capture = launch_owned(['tcpdump','-n','-U','--immediate-mode','-i','any','-w',str(state/'oam-diagnostic.pcap'),'tcp','port','9090' if core == 'open5gs' else '8000'],stdout=diagnostic_log,stderr=subprocess.STDOUT)
        processes.append(diagnostic_capture)
        for namespace, label in [([], 'core-startup'), (net, 'ran-startup')]:
            subprocess.run(namespace+['python3',str(HERE/'diagnostics.py'),'--state',str(state),'--label',label],check=False,timeout=15)
        probe_process = launch_owned(net+['python3',str(HERE/'probe.py'),'--packetrusher',str(Path(binary).resolve()),'--state',str(state)])
        status = probe_process.wait(timeout=180)
        if status != 0:
            raise subprocess.CalledProcessError(status, probe_process.args)
    finally:
        with defer_cancellation():
            # Stop the UE/probe before retiring its endpoint's namespace.
            if probe_process:
                stop_owned(probe_process)
            if link_owned:
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
    parser.add_argument('--sessions',type=int,choices=[0,1])
    parser.add_argument('--backend',choices=['userspace','ebpf','gtp5g'],default='userspace')
    parser.add_argument('--upf',choices=['free5gc','open5gs'])
    parser.add_argument('--upf-prefix')
    parser.add_argument('--pdu-session-type',choices=['IPv4','IPv6','IPv4v6'],default='IPv4',help='native Open5GS userspace/eBPF only; ordinary profiles remain IPv4')
    args=parser.parse_args()
    run(args.core,args.prefix,args.packetrusher,args.state,args.sessions,args.backend,args.upf,args.upf_prefix,args.pdu_session_type)
