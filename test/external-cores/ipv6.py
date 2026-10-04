#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Actual Open5GS IPv6 RA evidence; never invent a prefix from the NAS IID."""
import ipaddress
import socket
import struct
from pathlib import Path
from prepare import CORE_IP, RAN_IP


def require(condition, message):
    if not condition:
        raise AssertionError(message)


def inner_packets(path):
    """Decode the fixture's Ethernet/IPv4 N3 capture, including GTP extensions."""
    with Path(path).open('rb') as capture:
        header=capture.read(24)
        endian={b'\xd4\xc3\xb2\xa1':'<',b'\xa1\xb2\xc3\xd4':'>',b'\x4d\x3c\xb2\xa1':'<',b'\xa1\xb2\x3c\x4d':'>'}.get(header[:4])
        require(len(header)==24 and endian is not None and struct.unpack(endian+'I',header[20:24])[0]==1,'RA requires Ethernet PCAP')
        while record:=capture.read(16):
            require(len(record)==16,'truncated RA PCAP record')
            size=struct.unpack(endian+'IIII',record)[2]
            require(size<=65535,'unexpected RA capture size')
            frame=capture.read(size)
            require(len(frame)==size,'truncated RA capture frame')
            if len(frame)<42 or frame[12:14]!=b'\x08\x00':
                continue
            ip=frame[14:]; ihl=(ip[0]&15)*4
            if ip[0]>>4!=4 or ip[9]!=17 or ihl<20 or len(ip)<ihl+16:
                continue
            peers=(socket.inet_ntoa(ip[12:16]),socket.inet_ntoa(ip[16:20]))
            if peers not in {(CORE_IP,RAN_IP),(RAN_IP,CORE_IP)}:
                continue
            total=struct.unpack('!H',ip[2:4])[0]
            require(ihl+16<=total<=len(ip),'truncated N3 IPv4 body')
            udp=ip[ihl:total]
            if struct.unpack('!HH',udp[:4])!=(2152,2152):
                continue
            require(struct.unpack('!H',udp[4:6])[0]==len(udp),'invalid N3 UDP length')
            gtp=udp[8:]
            if (gtp[0]&0xf0)!=0x30 or gtp[1]!=255:
                continue
            require(struct.unpack('!H',gtp[2:4])[0]==len(gtp)-8,'invalid GTP length')
            offset=8
            if gtp[0]&7:
                require(len(gtp)>=12,'truncated GTP optional fields')
                offset,extension=12,gtp[11]
                require(gtp[0]&4 or not extension,'GTP extension flag absent')
                for _ in range(32):
                    if not extension:
                        break
                    require(offset<len(gtp),'truncated GTP extension')
                    length=gtp[offset]*4
                    require(length>=4 and offset+length<=len(gtp),'invalid GTP extension bounds')
                    extension,offset=gtp[offset+length-1],offset+length
                require(not extension,'GTP extension chain too long')
            yield peers,struct.unpack('!I',gtp[4:8])[0],gtp[offset:]


def checksum(data):
    data+=b'\0' if len(data)%2 else b''
    total=sum(struct.unpack('!'+str(len(data)//2)+'H',data))
    while total>>16:
        total=(total&65535)+(total>>16)
    return (~total)&65535


def router_advertisement_proof(path, expected_ue):
    expected=ipaddress.IPv6Address(expected_ue)
    network=ipaddress.IPv6Network((expected,64),strict=False)
    linklocal=ipaddress.IPv6Address(int(ipaddress.IPv6Address('fe80::'))|(int(expected)&((1<<64)-1)))
    for peers,teid,inner in inner_packets(path):
        if peers!=(CORE_IP,RAN_IP) or len(inner)<56 or inner[0]>>4!=6 or inner[6]!=58 or inner[40]!=134:
            continue
        require(teid>0 and len(inner)==40+struct.unpack('!H',inner[4:6])[0],'invalid captured RA length/TEID')
        source,destination=ipaddress.IPv6Address(inner[8:24]),ipaddress.IPv6Address(inner[24:40])
        require(source.is_link_local and destination==linklocal and inner[7]==255,'RA must use actual NAS IID link-local destination/source and hop255')
        body=inner[40:]
        pseudo=inner[8:40]+struct.pack('!I3xB',len(body),58)
        require(body[1]==0 and checksum(pseudo+body)==0,'invalid RA ICMPv6 code/checksum')
        lifetime=struct.unpack('!H',body[6:8])[0]
        require(lifetime>0,'RA has no usable default-router lifetime')
        offset=16
        found=None
        while offset<len(body):
            require(offset+2<=len(body),'truncated RA option')
            kind,length=body[offset],body[offset+1]*8
            require(length>0 and offset+length<=len(body),'invalid RA option length')
            option=body[offset:offset+length]
            if kind==3:
                require(length==32,'invalid RA PIO length')
                valid,preferred=struct.unpack('!II',option[4:12])
                prefix=ipaddress.IPv6Address(option[16:32])
                if option[2]==64 and option[3]&64 and prefix==network.network_address:
                    require(0<preferred<=valid,'invalid learned prefix lifetimes')
                    found={'prefix':str(network),'valid_lifetime':valid,'preferred_lifetime':preferred}
            offset+=length
        require(found is not None,'actual RA does not delegate the expected autonomous /64')
        return {'source':str(source),'destination':str(destination),'teid':teid,'hop_limit':255,'router_lifetime':lifetime,**found}
    raise AssertionError('capture lacks an actual Open5GS router advertisement')
