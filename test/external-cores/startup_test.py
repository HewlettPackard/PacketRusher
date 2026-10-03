# SPDX-License-Identifier: Apache-2.0
import json
import socket
import struct
import tempfile
import unittest
from unittest.mock import patch
from pathlib import Path
from startup import accepted_pfcp, open_registered_and_associated, free_registered_and_associated
from prepare import CORE_IP


def capture(cause=1, source=CORE_IP, destination='127.0.0.4'):
    pfcp = b'\x20\x06' + struct.pack('!H',9) + b'\x00\x00\x03\x00' + struct.pack('!HHB',19,1,cause)
    udp = struct.pack('!HHHH',8805,8805,len(pfcp)+8,0) + pfcp
    ip = b'\x45\x00' + struct.pack('!H',20+len(udp)) + b'\x00'*4 + b'\x40\x11\x00\x00' + socket.inet_aton(source) + socket.inet_aton(destination) + udp
    frame = b'\x00'*12+b'\x08\x00'+ip
    return struct.pack('<IHHIIII',0xa1b2c3d4,2,4,0,0,65535,1)+struct.pack('<IIII',0,0,len(frame),len(frame))+frame


class StartupEvidence(unittest.TestCase):
    def test_acceptance_requires_real_response_cause_and_configured_peers(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory)/'pfcp.pcap'
            for data in [capture(cause=64),capture(source='127.0.0.9'),capture()[:-1],capture()[:24]]:
                path.write_bytes(data)
                self.assertIsNone(accepted_pfcp(path))
            path.write_bytes(capture())
            self.assertEqual(accepted_pfcp(path),dict(source=CORE_IP,destination='127.0.0.4',sequence=3,cause=1))

    def test_associated_log_without_packet_and_deassociated_state_cannot_pass(self):
        with tempfile.TemporaryDirectory() as directory:
            state=Path(directory); (state/'core').mkdir()
            for name in ['udr','udm','ausf','bsf','pcf','nssf','smf','amf']:
                (state/'core'/f'{name}.log').write_text('NF registered [Heartbeat:10s]\n')
            for name in ['smf','upf']:
                with (state/'core'/f'{name}.log').open('a') as file: file.write('PFCP associated [peer]:8805\n')
            self.assertIsNone(open_registered_and_associated(state))
            (state/'pfcp-startup.pcap').write_bytes(capture())
            self.assertIsNotNone(open_registered_and_associated(state))
            with (state/'core'/'smf.log').open('a') as file: file.write('PFCP de-associated [peer]:8805\n')
            self.assertIsNone(open_registered_and_associated(state))

    def test_free5gc_current_association_requires_accepted_configured_peer_packet(self):
        with tempfile.TemporaryDirectory() as directory, patch('startup.free_registered',return_value={'smf':['actual-id']}):
            state=Path(directory); (state/'core').mkdir()
            profile={'sessions':1,'pfcp_smf_ip':'127.0.0.18','pfcp_upf_ip':CORE_IP}
            log=state/'core'/'smf-stdout.log'
            log.write_text('Received PFCP Association Setup Accepted Response from UPF\n')
            self.assertIsNone(free_registered_and_associated(state,profile))
            (state/'pfcp-startup.pcap').write_bytes(capture(destination='127.0.0.4'))
            self.assertIsNone(free_registered_and_associated(state,profile))
            (state/'pfcp-startup.pcap').write_bytes(capture(destination='127.0.0.18'))
            self.assertIsNotNone(free_registered_and_associated(state,profile))
            for event in ['PFCP Heartbeat error: expired','Canceled association to UPF[peer]','Canceled SMF PFCP context']:
                log.write_text('Received PFCP Association Setup Accepted Response from UPF\n'+event+'\n')
                self.assertIsNone(free_registered_and_associated(state,profile))

if __name__ == '__main__': unittest.main()
