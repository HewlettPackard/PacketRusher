# SPDX-License-Identifier: Apache-2.0
import copy
import socket
import struct
import tempfile
import unittest
from pathlib import Path
from probe import validate_report, gtpu_proof, open_registered_count
from prepare import CORE_IP, RAN_IP, UE_IP, DN_IP

def report(sessions):
    return {'schema_version':1,'ended_at':'2026-01-01T00:00:00Z','procedures':[dict(procedure=name,started=count,success=count,failure=0,cancelled=0,pending=0) for name,count in [('registration',1),('pdu_session_establishment',sessions)]]}

def packet(source, destination, payload, port_source, port_destination):
    udp=struct.pack('!HHHH',port_source,port_destination,len(payload)+8,0)+payload
    return b'\x45\x00'+struct.pack('!H',20+len(udp))+b'\x00'*4+b'\x40\x11\x00\x00'+socket.inet_aton(source)+socket.inet_aton(destination)+udp


def n3_capture(nonce, wrong_ue=False, duplicate=False):
    data=struct.pack('<IHHIIII',0xa1b2c3d4,2,4,0,0,65535,1)
    for uplink in [True,False]:
        for sequence in range(3):
            inner=packet(('10.45.0.3' if wrong_ue else UE_IP) if uplink else DN_IP,DN_IP if uplink else UE_IP,nonce+struct.pack('!I',0 if duplicate else sequence),40000 if uplink else 9000,9000 if uplink else 40000)
            # The uplink includes a real-shaped PDU Session Container/QFI.
            gtp=(b'\x34\xff'+struct.pack('!HI',len(inner)+8,7)+b'\x00\x00\x00\x85\x01\x10\x01\x00'+inner) if uplink else (b'\x30\xff'+struct.pack('!HI',len(inner),9)+inner)
            frame=b'\x00'*12+b'\x08\x00'+packet(RAN_IP if uplink else CORE_IP,CORE_IP if uplink else RAN_IP,gtp,2152,2152)
            data+=struct.pack('<IIII',0,0,len(frame),len(frame))+frame
    return data

class CompletionGuards(unittest.TestCase):
    def test_lazy_registered_gauge_is_empty_only_before_first_ue(self):
        header=b'# TYPE fivegs_amffunction_rm_registeredsubnbr gauge\n'
        self.assertEqual(open_registered_count(header,initial=True),0)
        with self.assertRaises(AssertionError): open_registered_count(header)
        for initial in [False,True]:
            with self.assertRaises(AssertionError): open_registered_count(b'other_metric 0\n',initial)
        for value in [b'NaN',b'+Inf',b'-1',b'0.5']:
            with self.assertRaises(AssertionError): open_registered_count(header+b'fivegs_amffunction_rm_registeredsubnbr{plmn="20893"} '+value+b'\n',initial=True)
        for value in [0,1]:
            self.assertEqual(open_registered_count(header+b'fivegs_amffunction_rm_registeredsubnbr{plmn="20893"} '+str(value).encode()+b'\n'),value)

    def test_registration_without_upf_requires_zero_pdu_attempts(self):
        validate_report(report(0),0)
        for field in ['started','success','failure','cancelled','pending']:
            bad=report(0)
            bad['procedures'][1][field]=1
            with self.assertRaises(AssertionError): validate_report(bad,0)

    def test_session_requires_actual_success_without_unfinished_attempts(self):
        validate_report(report(1),1)
        for procedure in [0,1]:
            for field in ['success','failure','cancelled','pending']:
                bad=report(1)
                bad['procedures'][procedure][field]=0 if field=='success' else 1
                with self.assertRaises(AssertionError): validate_report(bad,1)

    def test_partial_duplicate_and_wrong_schema_reports_cannot_pass(self):
        bad=report(1); bad.pop('ended_at')
        with self.assertRaises(AssertionError): validate_report(bad,1)
        bad=report(1); bad['schema_version']=2
        with self.assertRaises(AssertionError): validate_report(bad,1)
        bad=report(1); bad['procedures'][1]=copy.deepcopy(bad['procedures'][0])
        with self.assertRaises(AssertionError): validate_report(bad,1)

    def test_full_gtpu_proof_requires_allocated_ue_and_three_distinct_echoes(self):
        with tempfile.TemporaryDirectory() as directory:
            path=Path(directory)/'n3.pcap'
            nonce=b'private-random-echo'
            path.write_bytes(n3_capture(nonce))
            self.assertEqual(gtpu_proof(path,nonce),{'uplink':{'packets':3,'teids':[7]},'downlink':{'packets':3,'teids':[9]}})
            for kwargs in [{'wrong_ue':True},{'duplicate':True}]:
                path.write_bytes(n3_capture(nonce,**kwargs))
                with self.assertRaises(AssertionError): gtpu_proof(path,nonce)

    def test_empty_capture_cannot_pass_user_plane(self):
        with tempfile.TemporaryDirectory() as directory:
            path=Path(directory)/'n3.pcap'; path.write_bytes(b'')
            with self.assertRaises(AssertionError): gtpu_proof(path,b'unique')

if __name__=='__main__': unittest.main()
