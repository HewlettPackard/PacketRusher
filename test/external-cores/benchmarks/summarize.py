#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Keep failed/planned measurements visible alongside medians and ranges."""
import argparse
import json
import statistics
from collections import defaultdict
from pathlib import Path


def distribution(values):
    return {'median':statistics.median(values),'min':min(values),'max':max(values)} if values else None


def summarize(state):
    state=Path(state)
    results=json.loads((state/'results.json').read_text())
    environment=json.loads((state/'environment.json').read_text())
    groups=defaultdict(list)
    for cohort in results:
        for row in cohort['measurements']:
            if not row['warmup']:
                groups[(cohort['backend'],row['transport'],row['direction'],row['offered_bps'])].append({**row,'cohort_valid':cohort['success']})
    summary={'environment':environment,'cohorts':len(results),'failed_cohorts':sum(not row['success'] for row in results),
             'planned_cohorts':environment['repetitions']*3,'groups':[]}
    for backend in ('userspace','ebpf','gtp5g'):
        for transport,rates in (('tcp',[None]),('udp',environment['rates_bps'])):
            for direction in ('uplink','downlink'):
                for rate in rates:
                    rows=groups[(backend,transport,direction,rate)]
                    passed=[row for row in rows if row['success'] and row['cohort_valid']]
                    group={'backend':backend,'transport':transport,'direction':direction,'offered_bps':rate,
                           'planned':environment['repetitions'],'recorded':len(rows),'passed':len(passed),
                           'failures':[row.get('error','cohort failed strict lifecycle/core/report validation') for row in rows if not row['success'] or not row['cohort_valid']]}
                    for name in ('receiver_bps','sender_bps','lost_percent','jitter_ms','retransmits'):
                        group[name]=distribution([row['metrics'][name] for row in passed if row['metrics'].get(name) is not None])
                    for name in ('guest_busy_percent_one_cpu','guest_softirq_percent_one_cpu','iperf_client_cpu_percent_one_cpu'):
                        group[name]=distribution([row['cpu'][name] for row in passed])
                    for name in ('packetrusher','upf','iperf_server'):
                        group[name+'_cpu']=distribution([row['cpu']['processes'][name]['percent_one_cpu'] for row in passed])
                    summary['groups'].append(group)
    (state/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
    return summary


if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('state');args=parser.parse_args()
    print(json.dumps(summarize(args.state),indent=2))
