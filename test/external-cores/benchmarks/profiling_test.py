# SPDX-License-Identifier: Apache-2.0
import unittest
from copy import deepcopy
from profiling import perf_command, runtime_delta


class ProfileGuards(unittest.TestCase):
    def test_software_sampling_wraps_the_owned_actual_sender(self):
        sender=['iperf3','-c','10.45.0.1','-B','10.45.0.2','-M','1200']
        command=perf_command(sender,'/tmp/owned/uplink')
        self.assertEqual(command[command.index('--')+1:],sender)
        self.assertEqual(command[command.index('-e')+1],'cpu-clock')
        self.assertEqual(command[command.index('-o')+1],'/tmp/owned/uplink.perf.data')

    def test_stats_require_same_owner_program_tag_and_monotonic_counts(self):
        before={'pid':5,'start_ticks':77,'programs':[{'id':9,'name':'encap','tag':'abc','run_time_ns':200,'run_cnt':10}]}
        after=deepcopy(before);after['programs'][0].update(run_time_ns=1200,run_cnt=20)
        self.assertEqual(runtime_delta(before,after)[0]['mean_ns_per_run'],100)
        for key,value in [('tag','other'),('run_cnt',0),('run_time_ns',0)]:
            invalid=deepcopy(after);invalid['programs'][0][key]=value
            with self.assertRaises(AssertionError):runtime_delta(before,invalid)
        invalid=deepcopy(after);invalid['start_ticks']=78
        with self.assertRaises(AssertionError):runtime_delta(before,invalid)
        invalid=deepcopy(after);invalid['programs'][0].pop('run_time_ns')
        with self.assertRaises(AssertionError):runtime_delta(before,invalid)


if __name__=='__main__':unittest.main()
