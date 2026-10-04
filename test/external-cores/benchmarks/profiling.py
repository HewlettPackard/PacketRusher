# SPDX-License-Identifier: Apache-2.0
"""Optional sender CPU profiles; deliberately separate from throughput evidence."""
import json
import os
import subprocess
import time
from pathlib import Path


def process_start(pid):
    raw = Path(f'/proc/{pid}/stat').read_text()
    return int(raw[raw.rfind(')')+2:].split()[19])


class OwnedPrograms:
    def __init__(self, pid):
        self.pid = pid
        self.start = process_start(pid)
        self.ids = set()
        for path in Path(f'/proc/{pid}/fdinfo').iterdir():
            for line in path.read_text().splitlines():
                if line.startswith('prog_id:'):
                    self.ids.add(int(line.split(':', 1)[1]))
        if not self.ids:
            raise AssertionError('no programs owned by the PacketRusher PID')

    def snapshot(self):
        if process_start(self.pid) != self.start:
            raise AssertionError('PacketRusher PID was reused')
        result = {'pid': self.pid, 'start_ticks': self.start,
                  'monotonic': time.monotonic(), 'programs': []}
        for program_id in sorted(self.ids):
            value = json.loads(subprocess.check_output(
                ['bpftool', '-j', 'prog', 'show', 'id', str(program_id)], text=True))
            if isinstance(value, list):
                if len(value) != 1:
                    raise AssertionError('ambiguous owned program identity')
                value = value[0]
            if value['id'] != program_id:
                raise AssertionError('owned program identity changed')
            result['programs'].append(value)
        return result


def runtime_delta(before, after):
    if (before['pid'], before['start_ticks']) != (after['pid'], after['start_ticks']):
        raise AssertionError('program owner changed')
    original = {p['id']: p for p in before['programs']}
    current = {p['id']: p for p in after['programs']}
    if original.keys() != current.keys():
        raise AssertionError('owned program set changed')
    result = []
    for program_id, start in original.items():
        end = current[program_id]
        if (start['name'], start['tag']) != (end['name'], end['tag']):
            raise AssertionError('owned program tag changed')
        if not all(field in start and field in end for field in ('run_time_ns', 'run_cnt')):
            raise AssertionError('guest BPF runtime statistics are not enabled')
        elapsed = end['run_time_ns'] - start['run_time_ns']
        count = end['run_cnt'] - start['run_cnt']
        if elapsed < 0 or count < 0:
            raise AssertionError('owned runtime statistics reset')
        result.append({'id': program_id, 'name': start['name'], 'tag': start['tag'],
                       'run_time_ns': elapsed, 'runs': count,
                       'mean_ns_per_run': elapsed/count if count else None})
    return result


def perf_command(args, output):
    # Software sampling works without assuming a virtualized hardware PMU.
    # perf launches the actual iperf sender in the same owned process group.
    return ['perf', 'record', '-e', 'cpu-clock', '-F', '199', '-g',
            '-o', str(output)+'.perf.data', '--', *args]


def retain_report(output):
    data = str(output)+'.perf.data'
    for suffix, args in [('.perf-flat.txt', ['report', '--stdio', '--no-children',
                                           '--sort', 'symbol,dso']),
                         ('.perf-callgraph.txt', ['report', '--stdio', '--children',
                                                 '--call-graph', 'graph,0.5,caller'])]:
        with Path(str(output)+suffix).open('wb') as stdout, \
                Path(str(output)+suffix+'.stderr').open('wb') as stderr:
            subprocess.run(['perf', *args, '-i', data], stdout=stdout, stderr=stderr,
                           check=True, timeout=20)
