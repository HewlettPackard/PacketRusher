# SPDX-License-Identifier: Apache-2.0
import unittest
import json
import tempfile
from pathlib import Path
from unittest.mock import patch
import measure as module
from counters import NAMES
from measure import command, result_metrics, ticks
import os


class MeasurementGuards(unittest.TestCase):
    def test_owned_kernel_dumps_stay_outside_cpu_and_iperf_interval(self):
        events=[]
        original=module.snapshot
        class Source:
            def snapshot(self):
                events.append('kernel')
                return {'owner_pid':42,'owner_start_ticks':100,'map_id':17,'values':{name:len(events) for name in NAMES.values()}}
        def cpu(processes):
            events.append('cpu')
            return original(processes)
        value={'end':{'sum_received':{'bytes':1200,'bits_per_second':1000}}}
        with tempfile.TemporaryDirectory() as directory,patch('measure.snapshot',side_effect=cpu):
            path=Path(directory)/'flow'
            row=module.measure(['python3','-c',f'print({json.dumps(value)!r})'],path,{},1,Source())
            self.assertTrue(row['success'],row)
            self.assertEqual(events,['kernel','cpu','cpu','kernel'])
            self.assertTrue(path.with_suffix('.kernel-before.json').exists())
            self.assertTrue(path.with_suffix('.kernel-after.json').exists())
            self.assertEqual(row['kernel_counters'],{name:3 for name in NAMES.values()})

    def test_failed_post_kernel_snapshot_keeps_raw_flow_but_cannot_pass(self):
        class Source:
            calls=0
            def snapshot(self):
                self.calls+=1
                if self.calls==2:
                    raise AssertionError('owned map disappeared')
                return {'owner_pid':42,'owner_start_ticks':100,'map_id':17,'values':{name:0 for name in NAMES.values()}}
        value={'end':{'sum_received':{'bytes':1200,'bits_per_second':1000}}}
        with tempfile.TemporaryDirectory() as directory:
            path=Path(directory)/'flow'
            row=module.measure(['python3','-c',f'print({json.dumps(value)!r})'],path,{},1,Source())
            self.assertFalse(row['success'])
            self.assertIn('owned map disappeared',row['error'])
            self.assertEqual(json.loads(path.with_suffix('.json').read_text()),value)
            self.assertTrue(path.with_suffix('.cpu-after.json').exists())

    def test_matched_tcp_mss_udp_payload_source_and_reverse(self):
        for direction in ('uplink','downlink'):
            for transport in ('tcp','udp'):
                args=command(direction,transport,5,50_000_000)
                self.assertEqual(args[args.index('-B')+1],'10.45.0.2')
                self.assertEqual('-R' in args,direction=='downlink')
                self.assertEqual('--rcv-timeout' in args,direction=='downlink')
                self.assertEqual(args[args.index('-l' if transport=='udp' else '-M')+1],'1200')

    def test_sender_only_udp_stats_cannot_masquerade_as_receiver(self):
        value={'end':{'sum':{'sender':True,'bits_per_second':10}}}
        with self.assertRaises(AssertionError):result_metrics(value,'udp')

    def test_process_cpu_reads_actual_stat_without_comm_whitespace_confusion(self):
        value=ticks(os.getpid())
        self.assertGreater(value['start'],0)
        self.assertGreaterEqual(value['user'],0)

    def test_partial_error_json_never_counts_as_tcp_completion(self):
        with self.assertRaises(AssertionError):result_metrics({'error':'lost peer','end':{'sum_received':{'bytes':10}}},'tcp')


if __name__=='__main__':unittest.main()
