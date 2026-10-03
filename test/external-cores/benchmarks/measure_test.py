# SPDX-License-Identifier: Apache-2.0
import unittest
from measure import command, result_metrics, ticks
import os


class MeasurementGuards(unittest.TestCase):
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
