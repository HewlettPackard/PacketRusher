# SPDX-License-Identifier: Apache-2.0
import ipaddress
import json
import socket
import struct
import tempfile
import unittest
from pathlib import Path
from ipv6 import checksum, router_advertisement_proof
from prepare import CORE_IP, RAN_IP, UE_IP, DN_IP, UE_IPV6, DN_IPV6, generate, profile_options
from probe import family_nonce, gtpu_proof
from probe_test import n3_capture, packet


def ip6(source,destination,body,protocol=17,hop=64):
    peers=socket.inet_pton(socket.AF_INET6,source)+socket.inet_pton(socket.AF_INET6,destination)
    return b'\x60\0\0\0'+struct.pack('!HBB',len(body),protocol,hop)+peers+body


def udp6(source,destination,payload,uplink=True):
    udp=struct.pack('!HHHH',40000 if uplink else 9000,9000 if uplink else 40000,len(payload)+8,0)+payload
    peers=socket.inet_pton(socket.AF_INET6,source)+socket.inet_pton(socket.AF_INET6,destination)
    value=checksum(peers+struct.pack('!I3xB',len(udp),17)+udp) or 65535
    return ip6(source,destination,udp[:6]+struct.pack('!H',value)+udp[8:])


def ra(source='fe80::1',destination='fe80::2',prefix='2001:db8:cafe:1::',hop=255,valid=100,preferred=60,pio=True):
    body=struct.pack('!BBHBBHII',134,0,0,64,0,60,0,0)
    if pio:
        body+=struct.pack('!BBBBIII',3,4,64,192,valid,preferred,0)+socket.inet_pton(socket.AF_INET6,prefix)
    peers=socket.inet_pton(socket.AF_INET6,source)+socket.inet_pton(socket.AF_INET6,destination)
    value=checksum(peers+struct.pack('!I3xB',len(body),58)+body)
    return ip6(source,destination,body[:2]+struct.pack('!H',value)+body[4:],58,hop)


def n3(records):
    data=struct.pack('<IHHIIII',0xa1b2c3d4,2,4,0,0,65535,1)
    for uplink,inner in records:
        # A valid E+S optional sequence header with PDU Session Container.
        gtp=b'\x36\xff'+struct.pack('!HI',len(inner)+8,7 if uplink else 9)+b'\x12\x34\x00\x85\x01\x10\x01\x00'+inner
        frame=b'\0'*12+b'\x08\x00'+packet(RAN_IP if uplink else CORE_IP,CORE_IP if uplink else RAN_IP,gtp,2152,2152)
        data+=struct.pack('<IIII',0,0,len(frame),len(frame))+frame
    return data


class ActualIPv6Evidence(unittest.TestCase):
    def test_only_explicit_native_open5gs_portable_profiles_allow_ipv6(self):
        for family in ('IPv6','IPv4v6'):
            for backend in ('userspace','ebpf'):
                self.assertEqual(profile_options('open5gs',backend=backend,native=True,pdu_type=family),(1,backend,'open5gs'))
            for options in [dict(core='free5gc',sessions=1,native=True),dict(core='open5gs'),dict(core='open5gs',native=True,backend='gtp5g')]:
                with self.assertRaises(ValueError):profile_options(pdu_type=family,**options)

    def test_real_pools_subscriber_types_and_requested_nas_family_match(self):
        for family,kind,versions in [('IPv6',2,[6]),('IPv4v6',3,[4,6])]:
            with self.subTest(family=family),tempfile.TemporaryDirectory() as directory:
                generate('open5gs',directory,native=True,pdu_type=family,backend='ebpf')
                root=Path(directory)
                profile=json.loads((root/'profile.json').read_text())
                self.assertEqual([row['version'] for row in profile['user_plane_families']],versions)
                smf=json.loads((root/'config/smf.yaml').read_text())['smf']
                upf=json.loads((root/'config/upf.yaml').read_text())['upf']
                self.assertEqual([ipaddress.ip_network(row['subnet']).version for row in smf['session']],versions)
                self.assertEqual([ipaddress.ip_network(row['subnet']).version for row in upf['session']],versions)
                subscriber=json.loads((root/'subscriber.js').read_text().split('insertOne(',1)[1].rsplit(');',1)[0])
                session=subscriber['slice'][0]['session'][0]
                self.assertEqual(session['type'],kind)
                self.assertEqual(session['ue'],({'ipv4':UE_IP} if 4 in versions else {})|{'ipv6':UE_IPV6})
                self.assertEqual(json.loads((root/'config/packetrusher.yaml').read_text())['ue']['pdusessiontype'],family)

    def test_dual_family_capture_requires_each_allocated_pair_and_real_udp_checksum(self):
        with tempfile.TemporaryDirectory() as directory:
            path=Path(directory)/'n3.pcap'
            base=b'unique-dual-family-identifier'
            nonce4,nonce6=family_nonce(base,4),family_nonce(base,6)
            self.assertNotIn(nonce4,nonce6); self.assertNotIn(nonce6,nonce4)
            records=[(up,udp6(UE_IPV6 if up else DN_IPV6,DN_IPV6 if up else UE_IPV6,nonce6+struct.pack('!I',i),up)) for up in (True,False) for i in range(3)]
            path.write_bytes(n3_capture(nonce4)+n3(records)[24:])
            self.assertEqual(gtpu_proof(path,nonce4)['uplink']['packets'],3)
            self.assertEqual(gtpu_proof(path,nonce6,family=6,ue_ip=UE_IPV6,dn_ip=DN_IPV6)['downlink']['packets'],3)
            for bad_ue in ['2001:db8:cafe:1::3','2001:db8:cafe:2::2']:
                with self.assertRaises(AssertionError):gtpu_proof(path,nonce6,family=6,ue_ip=bad_ue,dn_ip=DN_IPV6)
            corrupt=bytearray(records[0][1]);corrupt[-1]^=1
            path.write_bytes(n3([(True,bytes(corrupt)),*records[1:]]))
            with self.assertRaises(AssertionError):gtpu_proof(path,nonce6,family=6,ue_ip=UE_IPV6,dn_ip=DN_IPV6)

    def test_actual_ra_requires_hop_source_iid_prefix_checksum_and_valid_lifetimes(self):
        with tempfile.TemporaryDirectory() as directory:
            path=Path(directory)/'n3.pcap';path.write_bytes(n3([(False,ra())]))
            proof=router_advertisement_proof(path,UE_IPV6)
            self.assertEqual(proof['prefix'],'2001:db8:cafe:1::/64')
            self.assertEqual(proof['destination'],'fe80::2')
            for options in [dict(hop=254),dict(source=DN_IPV6),dict(destination='fe80::3'),dict(prefix='2001:db8:cafe:2::'),dict(valid=0),dict(preferred=101),dict(pio=False)]:
                path.write_bytes(n3([(False,ra(**options))]))
                with self.assertRaises(AssertionError):router_advertisement_proof(path,UE_IPV6)
            bad=bytearray(ra());bad[42]^=1;path.write_bytes(n3([(False,bytes(bad))]))
            with self.assertRaises(AssertionError):router_advertisement_proof(path,UE_IPV6)

    def test_ipv6_missing_dn_needs_uplink_and_rejects_response(self):
        with tempfile.TemporaryDirectory() as directory:
            path=Path(directory)/'n3.pcap';nonce=b'unique-negative-family'
            ul=udp6(UE_IPV6,DN_IPV6,nonce+struct.pack('!I',3))
            dl=udp6(DN_IPV6,UE_IPV6,nonce+struct.pack('!I',3),False)
            path.write_bytes(n3([(True,ul)]))
            self.assertEqual(gtpu_proof(path,nonce,(3,),False,6,UE_IPV6,DN_IPV6)['downlink']['packets'],0)
            path.write_bytes(n3([(True,ul),(False,dl)]))
            with self.assertRaises(AssertionError):gtpu_proof(path,nonce,(3,),False,6,UE_IPV6,DN_IPV6)


if __name__=='__main__':unittest.main()
