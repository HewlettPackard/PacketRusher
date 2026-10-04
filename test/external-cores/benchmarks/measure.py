#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Matched iperf3 commands and raw guest/process CPU accounting."""
import json
import math
import os
import resource
import socket
import struct
import subprocess
import time
from pathlib import Path


def preflight_qfis(path, nonce):
    data=Path(path).read_bytes(); offset=24; result={'uplink':set(),'downlink':set()}
    endian={'d4c3b2a1':'<','a1b2c3d4':'>'}.get(data[:4].hex())
    if endian is None:raise AssertionError('unsupported preflight PCAP')
    while offset+16<=len(data):
        size=struct.unpack(endian+'IIII',data[offset:offset+16])[2]
        frame=data[offset+16:offset+16+size];offset+=16+size
        if len(frame)<70 or frame[12:14]!=b'\x08\x00':continue
        ip=frame[14:];ihl=(ip[0]&15)*4
        if ip[9]!=17:continue
        gtp=ip[ihl+8:]
        if nonce not in gtp or len(gtp)<16 or gtp[1]!=255:continue
        peers=(socket.inet_ntoa(ip[12:16]),socket.inet_ntoa(ip[16:20]))
        direction='uplink' if peers==('172.30.5.30','172.30.5.20') else 'downlink' if peers==('172.30.5.20','172.30.5.30') else None
        if direction and gtp[0]&4 and gtp[11]==0x85 and gtp[12]==1:
            result[direction].add(gtp[14]&63)
    if any(values!={1} for values in result.values()):raise AssertionError(f'expected same real configured QFI1 both directions: {result}')
    return {direction:sorted(values) for direction,values in result.items()}


def command(direction, transport, seconds, rate=None):
    args = ['iperf3', '-c', '10.45.0.1', '-B', '10.45.0.2', '-p', '5201',
            '-J', '--get-server-output', '-t', str(seconds), '-i', '0',
            '--connect-timeout', '3000', '--snd-timeout', '5000']
    if direction == 'downlink': args += ['-R','--rcv-timeout','5000']
    if transport == 'udp': args += ['-u', '-l', '1200', '-b', str(rate)]
    else: args += ['-M', '1200']
    return args


def ticks(pid):
    raw = Path(f'/proc/{pid}/stat').read_text()
    # comm may contain spaces or parentheses; fields after its final ')' start
    # at field3. Process utime/stime aggregate all its live threads.
    values = raw[raw.rfind(')')+2:].split()
    return {'user': int(values[11]), 'system': int(values[12]), 'start': int(values[19])}


def snapshot(processes):
    cpu = next(line.split()[1:] for line in Path('/proc/stat').read_text().splitlines() if line.startswith('cpu '))
    usage = resource.getrusage(resource.RUSAGE_CHILDREN)
    return {'monotonic': time.monotonic(), 'guest_cpu_ticks': list(map(int,cpu)),
            'softirqs': Path('/proc/softirqs').read_text(),
            'processes': {name: ticks(pid) for name,pid in processes.items()},
            'reaped_child_cpu_seconds': usage.ru_utime+usage.ru_stime}


def cpu_delta(before, after):
    elapsed = after['monotonic']-before['monotonic']
    hz = os.sysconf('SC_CLK_TCK')
    delta = [end-start for start,end in zip(before['guest_cpu_ticks'],after['guest_cpu_ticks'])]
    # Linux guest/guest_nice are already counted in user/nice; do not double count.
    busy = sum(delta[i] for i in (0,1,2,5,6))
    processes = {}
    for name,start in before['processes'].items():
        end = after['processes'][name]
        if start['start'] != end['start']: raise AssertionError(f'PID reuse for {name}')
        seconds = (end['user']+end['system']-start['user']-start['system'])/hz
        processes[name] = {'cpu_seconds':seconds,'percent_one_cpu':100*seconds/elapsed}
    return {'wall_seconds':elapsed,'guest_busy_percent_one_cpu':100*busy/hz/elapsed,
            'guest_softirq_percent_one_cpu':100*delta[6]/hz/elapsed,
            'guest_irq_percent_one_cpu':100*delta[5]/hz/elapsed,
            'guest_steal_percent_one_cpu':100*delta[7]/hz/elapsed,
            'processes':processes,
            'iperf_client_cpu_seconds':after['reaped_child_cpu_seconds']-before['reaped_child_cpu_seconds'],
            'iperf_client_cpu_percent_one_cpu':100*(after['reaped_child_cpu_seconds']-before['reaped_child_cpu_seconds'])/elapsed,
            'clock_ticks_per_second':hz}


def result_metrics(value, transport):
    if value.get('error'): raise AssertionError(value['error'])
    end = value.get('end',{})
    if transport == 'tcp':
        received = end.get('sum_received')
        if not received or received.get('bytes',0) <= 0: raise AssertionError('missing actual TCP receiver bytes')
        return {'receiver_bps':received['bits_per_second'],'receiver_bytes':received['bytes'],
                'sender_bps':end.get('sum_sent',{}).get('bits_per_second'),
                'retransmits':end.get('sum_sent',{}).get('retransmits')}
    # iperf3 3.16 supplies receiver statistics in sum_received; retain raw JSON
    # and accept sum only when its sender flag explicitly identifies a receiver.
    received = end.get('sum_received') or end.get('sum')
    if not received or received.get('sender') is not False: raise AssertionError('missing UDP receiver statistics')
    fields = {name:received[name] for name in ('bits_per_second','bytes','packets','lost_packets','lost_percent','jitter_ms')}
    if not all(math.isfinite(fields[name]) for name in ('bits_per_second','lost_percent','jitter_ms')):
        raise AssertionError('nonfinite UDP measurement')
    sent=end.get('sum_sent',{})
    if sent.get('sender') is not True or not math.isfinite(sent.get('bits_per_second',float('nan'))):
        raise AssertionError('missing actual UDP sender rate')
    return {'receiver_bps':fields.pop('bits_per_second'),'sender_bps':sent['bits_per_second'],**fields}


def measure(args, output, processes, seconds, counter_source=None, profile='none', programs=None):
    output = Path(output)
    program_before=programs.snapshot() if programs else None
    if program_before:
        output.with_suffix('.program-before.json').write_text(json.dumps(program_before,indent=2)+'\n')
    kernel_before=counter_source.snapshot() if counter_source else None
    if kernel_before:
        output.with_suffix('.kernel-before.json').write_text(json.dumps(kernel_before,indent=2)+'\n')
    before = snapshot(processes)
    output.with_suffix('.cpu-before.json').write_text(json.dumps(before,indent=2)+'\n')
    start = time.monotonic()
    error = None
    with output.with_suffix('.json').open('wb') as stdout, output.with_suffix('.stderr').open('wb') as stderr:
        executed=args
        if profile=='cpu-clock':
            from profiling import perf_command
            executed=perf_command(args,output)
        process = subprocess.Popen(executed,stdout=stdout,stderr=stderr,start_new_session=True)
        try:
            code = process.wait(timeout=seconds+12)
        except BaseException as error:
            import signal
            try: os.killpg(process.pid,signal.SIGTERM)
            except ProcessLookupError: pass
            try: process.wait(timeout=2)
            except subprocess.TimeoutExpired:
                try: os.killpg(process.pid,signal.SIGKILL)
                except ProcessLookupError: pass
                process.wait()
            failure={'command':args,'exit_code':process.returncode,'success':False,'error':str(error)}
            try:
                after=snapshot(processes)
                output.with_suffix('.cpu-after.json').write_text(json.dumps(after,indent=2)+'\n')
                failure['cpu']=cpu_delta(before,after)
            except Exception as cpu_error:failure['cpu_error']=str(cpu_error)
            if counter_source:
                try:
                    kernel_after=counter_source.snapshot()
                    output.with_suffix('.kernel-after.json').write_text(json.dumps(kernel_after,indent=2)+'\n')
                    from counters import delta
                    failure['kernel_counters']=delta(kernel_before,kernel_after)
                except Exception as counter_error:failure['kernel_counter_error']=str(counter_error)
            output.with_suffix('.measurement.json').write_text(json.dumps(failure,indent=2)+'\n')
            raise
    after = snapshot(processes)
    row = {'command':args,'exit_code':code,'wall_seconds':time.monotonic()-start,
           'cpu':cpu_delta(before,after),'success':False,'profile':profile,
           'executed_command':executed}
    try:
        if code: raise AssertionError(f'iperf exit={code}')
        value=json.loads(output.with_suffix('.json').read_text())
        row['metrics']=result_metrics(value,'udp' if '-u' in args else 'tcp')
        if counter_source:
            # Both dumps bracket CPU sampling, outside the timed iperf process.
            kernel_after=counter_source.snapshot()
            output.with_suffix('.kernel-after.json').write_text(json.dumps(kernel_after,indent=2)+'\n')
            from counters import delta
            row['kernel_counters']=delta(kernel_before,kernel_after)
        if programs:
            from profiling import runtime_delta
            program_after=programs.snapshot()
            output.with_suffix('.program-after.json').write_text(json.dumps(program_after,indent=2)+'\n')
            row['program_runtime']=runtime_delta(program_before,program_after)
        if profile=='cpu-clock':
            from profiling import retain_report
            retain_report(output)
        row['success']=True
    except Exception as exc: row['error']=str(exc)
    output.with_suffix('.cpu-before.json').write_text(json.dumps(before,indent=2)+'\n')
    output.with_suffix('.cpu-after.json').write_text(json.dumps(after,indent=2)+'\n')
    output.with_suffix('.measurement.json').write_text(json.dumps(row,indent=2)+'\n')
    return row
