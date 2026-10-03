#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Disposable-guest comparison against one genuine free5GC SMF/UPF."""
import argparse
import hashlib
import json
import os
import random
import signal
import socket
import struct
import subprocess
import sys
import time
from pathlib import Path
from measure import command, measure, preflight_qfis


class Cancelled(BaseException):
    pass


def cancel(signum, _):
    # subprocess.wait may treat InterruptedError as syscall EINTR and retry.
    # A distinct cancellation exception must reach the ownership finally block.
    raise Cancelled(f'benchmark interrupted by signal {signum}')


def record(path, value):
    Path(path).write_text(json.dumps(value,indent=2)+'\n')


def imports(fixture):
    sys.path.insert(0,str(Path(fixture).resolve()))
    global native, prepare, probe
    import native, prepare, probe


def cohort(args):
    state=Path(args.state)
    profile=json.loads((state.parent/'profile.json').read_text())
    binary=Path(args.packetrusher).resolve()
    config=json.loads((state.parent/'config/packetrusher.yaml').read_text())
    config['ue']['tunnelbackend']=args.backend
    record(state/'config.json',config)
    processes={name:int(pid) for name,pid in json.loads(args.processes).items()}
    owned,files=[],[]
    capture=None
    rows=[]
    result={'backend':args.backend,'repetition':args.repetition,'success':False,'measurements':rows}
    for signum in (signal.SIGINT,signal.SIGTERM): signal.signal(signum,cancel)
    log=(state/'packetrusher.log').open('wb'); files.append(log)
    process=None
    def control(action,label=None):
        response=subprocess.run([str(binary),'control','--socket',str(state/'control.sock'),'--ue','1',
             '--action',action,'--timeout','45s'],check=True,capture_output=True,text=True,timeout=50)
        value=json.loads(response.stdout)
        record(state/f'{label or action}.json',value)
        return value['ues'][0]
    try:
        probe.await_core_count(profile,state,'initial',0,10)
        capture_log=(state/'capture.log').open('wb'); files.append(capture_log)
        capture=native.launch_owned(['tcpdump','-n','-U','--immediate-mode','-i','eth0','-w',str(state/'preflight-n3.pcap'),'udp','port','2152'],stdout=capture_log,stderr=subprocess.STDOUT)
        probe.until(lambda:capture.poll() is None and (state/'preflight-n3.pcap').exists() and (state/'preflight-n3.pcap').stat().st_size>=24,5,'preflight capture')
        process=native.launch_owned([str(binary),'--config',str(state/'config.json'),'--tunnel-backend',args.backend,
             '--report-json',str(state/'report.json'),'multi-ue','-n','1','--numPduSessions','1',
             '--tunnel','--tunnel-shared=true','--tunnel-vrf=false','--control-socket',str(state/'control.sock')],stdout=log,stderr=subprocess.STDOUT)
        owned.append(process)
        probe.until(lambda:process.poll() is None and (state/'control.sock').is_socket(),20,'UE control socket')
        ready=control('wait')
        probe.require(ready['ready'] and ready['connected'] and ready['state']=='registered' and ready['active_pdu_sessions']==[1],f'incomplete PDU: {ready}')
        probe.await_core_count(profile,state,'registered',1,5)
        route=json.loads(subprocess.check_output(['ip','-json','route','get',prepare.DN_IP,'from',prepare.UE_IP],text=True))
        record(state/'ue-route.json',route)
        expected='valgnb'+socket.inet_aton(prepare.RAN_IP).hex() if args.backend=='gtp5g' else 'val0000000120'
        probe.require(route[0]['dev']==expected,f'unexpected owned backend route: {route}')
        link=json.loads(subprocess.check_output(['ip','-json','link','show','dev',route[0]['dev']],text=True))
        record(state/'ue-link.json',link)
        probe.require(link[0]['mtu']==1400,'backend MTU differs from1400')
        if args.backend=='gtp5g':probe.require(link[0].get('linkinfo',{}).get('info_kind')=='gtp5g','route does not select kernel gtp5g')
        endpoint=json.loads(subprocess.check_output(['ip','-json','addr','show','dev','val0000000120'],text=True))
        record(state/'ue-endpoint.json',endpoint)
        probe.require(any(addr.get('local')==prepare.UE_IP for addr in endpoint[0]['addr_info']),'stable endpoint does not own UE address')
        nonce=b'PACKETRUSHER-BENCH-PREFLIGHT-'+os.urandom(24).hex().encode()
        with socket.socket(socket.AF_INET,socket.SOCK_DGRAM) as sock:
            sock.bind((prepare.UE_IP,0)); sock.settimeout(5)
            for sequence in range(3):
                payload=nonce+struct.pack('!I',sequence)
                sock.sendto(payload,(prepare.DN_IP,9000))
                data,peer=sock.recvfrom(4096)
                probe.require(data==payload and peer==(prepare.DN_IP,9000),'real UPF preflight echo failed')
        probe.until(lambda:probe.gtpu_proof(state/'preflight-n3.pcap',nonce),3,'preflight N3 evidence')
        capture.send_signal(signal.SIGINT)
        probe.require(capture.wait(timeout=5)==0,'preflight capture failed')
        capture=None
        result['preflight_user_plane']=probe.gtpu_proof(state/'preflight-n3.pcap',nonce)
        result['preflight_qfis']=preflight_qfis(state/'preflight-n3.pcap',nonce)
        processes['packetrusher']=process.pid
        cases=[(direction,'tcp',None) for direction in ('uplink','downlink')]
        cases += [(direction,'udp',rate) for rate in args.rates for direction in ('uplink','downlink')]
        random.Random(args.seed+args.repetition).shuffle(cases)
        for index,(direction,transport,rate) in enumerate(cases):
            if process.poll() is not None: raise RuntimeError('PacketRusher exited during benchmark')
            name=f'{index:02d}-{transport}-{direction}-{rate or "uncapped"}'
            # A separate retained warmup avoids charging startup/omit traffic
            # to the measured receiver interval or its CPU sample.
            for warmup,seconds in ((True,1),(False,args.seconds)):
                row={'direction':direction,'transport':transport,'offered_bps':rate,'warmup':warmup}
                try:
                    row.update(measure(command(direction,transport,seconds,rate),state/(name+('-warmup' if warmup else '')),processes,seconds))
                except Cancelled:
                    raise
                except Exception as error:
                    row.update(success=False,error=str(error))
                rows.append(row)
                record(state/'result.json',result)
        still=control('wait','post-traffic-ready')
        probe.require(still['ready'] and still['connected'] and still['state']=='registered' and still['active_pdu_sessions']==[1],'traffic retired PDU')
        probe.await_core_count(profile,state,'post-traffic',1,5)
        retired=control('deregister')
        probe.require(retired['state']=='parked' and not retired['connected'] and not retired['ready'] and retired['active_pdu_sessions']==[],'PDU not retired')
        probe.await_core_count(profile,state,'deregistered',0,12)
        process.send_signal(signal.SIGINT)
        probe.require(process.wait(timeout=10)==0,'UE shutdown failed')
        probe.validate_report(json.loads((state/'report.json').read_text()),1)
        result['core_counts']=[0,1,1,0]
        result['success']=all(row['success'] for row in rows)
    except (Exception,Cancelled) as error:
        result['error']=str(error)
        # Retire a live owned UE on failure before the next cohort can proceed.
        if not isinstance(error,Cancelled) and process and process.poll() is None and (state/'control.sock').is_socket():
            try:
                control('deregister','failure-deregister')
                probe.await_core_count(profile,state,'failure-deregistered',0,12)
            except Exception as cleanup: result['retirement_error']=str(cleanup)
    finally:
        with native.defer_cancellation():
            if capture: native.stop_owned(capture,2)
            for child in reversed(owned): native.stop_owned(child,10)
            for file in files: file.close()
            record(state/'result.json',result)
    return 0 if result['success'] else 1


def run(args):
    if os.geteuid()!=0 or os.stat('/proc/self/ns/net').st_ino==os.stat('/proc/1/ns/net').st_ino:
        raise RuntimeError('run via private namespace wrapper in disposable guest')
    if not Path('/sys/module/gtp5g').is_dir(): raise RuntimeError('genuine UPF and kernel backend require guest gtp5g')
    state=Path(args.state).resolve()
    if state.exists(): raise RuntimeError('benchmark state must be fresh')
    state.mkdir(parents=True)
    record(state/'results.json',[])
    record(state/'environment.json',{'repetitions':args.repetitions,'rates_bps':args.rates,'pilot':args.pilot,'setup_complete':False})
    os.environ['GOMAXPROCS']='2'
    prepare.generate('free5gc',state,native=True,prefix=args.prefix,sessions=1,backend='userspace',upf='free5gc')
    # Benchmark-only headroom; hosted/native acceptance keeps its original
    # subscription. Use the official UE and session AMBR schema already generated.
    subscriber=state/'subscriber.js'
    subscription=subscriber.read_text()
    if '"1 Gbps"' not in subscription:raise RuntimeError('expected benchmark AMBR schema missing')
    subscriber.write_text(subscription.replace('"1 Gbps"','"10 Gbps"'))
    smf_path=state/'config/smfcfg.yaml'
    smf=json.loads(smf_path.read_text())
    # The release's default1000-byte URR threshold caused one real PFCP usage
    # report per payload and dominated the pilot. Match a quiet1TB/1h policy
    # for all backends, retaining the same genuine SMF/UPF and packet rules.
    smf['configuration']['urrPeriod']=3600
    smf['configuration']['urrThreshold']=1_000_000_000_000
    record(smf_path,smf)
    binary=Path(args.packetrusher).resolve()
    record(state/'environment.json',{'kernel':subprocess.check_output(['uname','-a'],text=True).strip(),
        'iperf':subprocess.check_output(['iperf3','--version'],text=True).strip(),
        'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),
        'binary_version':json.loads(subprocess.check_output([str(binary),'version','--json'],text=True)),
        'affinity':sorted(os.sched_getaffinity(0)),'GOMAXPROCS':2,'mtu':1400,'tcp_mss':1200,'udp_payload':1200,
        'rates_bps':args.rates,'seconds':args.seconds,'repetitions':args.repetitions,'seed':args.seed,
        'module_version':Path('/sys/module/gtp5g/version').read_text().strip(),
        'logging_level':4,'routing':'source policy, no VRF','captures_during_timing':False,
        'pilot':args.pilot,'setup_complete':True,'ue_and_session_ambr':'10 Gbps',
        'urr_period_seconds':3600,'urr_threshold_bytes':1_000_000_000_000,
        'iperf_binary_sha256':hashlib.sha256(Path('/usr/bin/iperf3').read_bytes()).hexdigest(),
        'module_sha256':hashlib.sha256(Path(args.module).read_bytes()).hexdigest(),
        'source_hashes':{str(path):hashlib.sha256(path.read_bytes()).hexdigest() for path in
          [Path(__file__),Path(__file__).with_name('measure.py'),*[Path(args.fixture)/name for name in ('prepare.py','native.py','core.py','probe.py','startup.py')]]},
        'cpu_topology':json.loads(subprocess.check_output(['lscpu','--json'],text=True))})
    processes,files=[],[]
    workers=set()
    for signum in (signal.SIGINT,signal.SIGTERM): signal.signal(signum,cancel)
    results=[]
    link_owned=False
    try:
        subprocess.run(['ip','link','set','lo','up'],check=True)
        child=native.launch_owned(['unshare','--net','sleep','3600']);processes.append(child)
        probe.until(lambda:child.poll() is None and os.stat(f'/proc/{child.pid}/ns/net').st_ino!=os.stat('/proc/self/ns/net').st_ino,5,'RAN namespace')
        net=['nsenter','--target',str(child.pid),'--net']
        subprocess.run(['ip','link','add','core0','type','veth','peer','name','ran0'],check=True);link_owned=True
        subprocess.run(['ip','link','set','ran0','netns',str(child.pid)],check=True)
        for namespace,commands in (([],[['ip','addr','add',prepare.CORE_IP+'/24','dev','core0'],['ip','link','set','core0','up']]),
             (net,[['ip','link','set','lo','up'],['ip','link','set','ran0','name','eth0'],['ip','addr','add',prepare.RAN_IP+'/24','dev','eth0'],['ip','link','set','eth0','up']])):
            for command_ in commands: subprocess.run(namespace+command_,check=True)
        # Every backend gets identical N3 offload settings, including kernel.
        for namespace,device in (([],'core0'),(net,'eth0')):
            subprocess.run(namespace+['ethtool','-K',device,'tx','off','rx','off','tso','off','gso','off','gro','off'],check=True)
            (state/f'offloads-{device}.txt').write_text(subprocess.check_output(namespace+['ethtool','-k',device],text=True))
        db=state/'mongo';db.mkdir()
        log=(state/'mongo.log').open('wb');files.append(log)
        mongo=native.launch_owned(['mongod','--dbpath',str(db),'--bind_ip','127.0.0.1','--nounixsocket','--wiredTigerCacheSizeGB','.25'],stdout=log,stderr=subprocess.STDOUT);processes.append(mongo)
        probe.until(lambda:mongo.poll() is None and subprocess.run(['mongosh','--quiet','mongodb://127.0.0.1:27017/free5gc','--eval','db.adminCommand({ping:1}).ok'],capture_output=True,timeout=3).returncode==0,30,'private Mongo')
        subprocess.run(['mongosh','--quiet','mongodb://127.0.0.1:27017/free5gc',str(state/'subscriber.js')],check=True,timeout=10)
        log=(state/'core-supervisor.log').open('wb');files.append(log)
        supervisor=native.launch_owned(['python3',str(Path(args.fixture)/'core.py'),'--prefix',args.prefix,'--state',str(state),'--stop-capture-when-ready'],stdout=log,stderr=subprocess.STDOUT);processes.append(supervisor)
        probe.until(lambda:supervisor.poll() is None and (state/'core-ready').exists(),65,'genuine core/PFCP')
        nf={}
        for pid in Path(f'/proc/{supervisor.pid}/task/{supervisor.pid}/children').read_text().split():
            try:
                exe=Path(os.readlink(f'/proc/{pid}/exe'))
                if exe.name in {'amf','smf','upf','nrf','udm','udr','ausf','nssf','pcf'}:nf[exe.name]=int(pid)
            except FileNotFoundError: pass
        probe.require(len(nf)==9,'missing genuine NF PID identity')
        record(state/'nf-identities.json',{name:{'pid':pid,'binary_sha256':hashlib.sha256(Path(f'/proc/{pid}/exe').read_bytes()).hexdigest()} for name,pid in nf.items()})
        log=(state/'iperf-server.log').open('wb');files.append(log)
        server=native.launch_owned(['iperf3','-s','-B',prepare.DN_IP,'-p','5201','-J'],stdout=log,stderr=subprocess.STDOUT);processes.append(server)
        probe.until(lambda:server.poll() is None and ':5201' in subprocess.check_output(['ss','-H','-ltn'],text=True),5,'owned DN iperf server')
        nf['iperf_server']=server.pid
        schedule=[]
        rng=random.Random(args.seed)
        for repetition in range(args.repetitions):
            order=['userspace','ebpf','gtp5g'];rng.shuffle(order)
            schedule.extend((repetition,backend) for backend in order)
        record(state/'schedule.json',schedule)
        for index,(repetition,backend) in enumerate(schedule):
            if supervisor.poll() is not None or server.poll() is not None: raise RuntimeError('real core/server exited')
            folder=state/f'{index:02d}-{backend}-rep{repetition}';folder.mkdir()
            log=(folder/'driver.log').open('wb');files.append(log)
            command_ = net+[sys.executable,str(Path(__file__).resolve()),'--fixture',args.fixture,'--packetrusher',str(binary),
               '--state',str(folder),'--cohort','--backend',backend,'--repetition',str(repetition),
               '--seconds',str(args.seconds),'--seed',str(args.seed),'--processes',json.dumps(nf),'--rates',*map(str,args.rates)]
            worker=native.launch_owned(command_,stdout=log,stderr=subprocess.STDOUT);processes.append(worker)
            workers.add(worker.pid)
            status=worker.wait(timeout=(len(args.rates)*2+2)*(args.seconds+20)+90)
            value=json.loads((folder/'result.json').read_text())
            results.append({'index':index,'exit_code':status,'directory':folder.name,**value})
            record(state/'results.json',results)
            profile=json.loads((state/'profile.json').read_text())
            probe.await_core_count(profile,state,f'between-{index}',0,12)
        return 0 if all(row['success'] for row in results) else 1
    finally:
        with native.defer_cancellation():
            if link_owned:subprocess.run(['ip','link','del','core0'],check=False)
            # Active cohorts own a2s iperf kill,10s UE join and2s capture join.
            # Give that inner finally time to finish before the parent kills it.
            for child in reversed(processes):native.stop_owned(child,20 if child.pid in workers else 12)
            for file in files:file.close()


if __name__=='__main__':
    parser=argparse.ArgumentParser()
    parser.add_argument('--fixture',required=True);parser.add_argument('--packetrusher',required=True)
    parser.add_argument('--prefix');parser.add_argument('--state',required=True)
    parser.add_argument('--repetitions',type=int,default=3);parser.add_argument('--seconds',type=int,default=5)
    parser.add_argument('--rates',type=int,nargs='+',default=[50_000_000,200_000_000,800_000_000,2_000_000_000])
    parser.add_argument('--seed',type=int,default=20261004)
    parser.add_argument('--module',default='/home/tester/packetrusher-real-cores/gtp5g/gtp5g.ko')
    parser.add_argument('--pilot',action='store_true',help='explicit setup pilot; permits fewer than3 repetitions, not a final comparison')
    parser.add_argument('--cohort',action='store_true');parser.add_argument('--backend',choices=['userspace','ebpf','gtp5g'])
    parser.add_argument('--repetition',type=int,default=0);parser.add_argument('--processes',default='{}')
    args=parser.parse_args()
    if not 1<=args.repetitions<=10 or (not args.pilot and args.repetitions<3):parser.error('final comparison requires3–10 repetitions; --pilot permits1–2')
    if not 1<=args.seconds<=60 or not args.rates or any(rate<=0 for rate in args.rates):parser.error('positive rates and1–60seconds required')
    imports(args.fixture)
    sys.exit(cohort(args) if args.cohort else run(args))
