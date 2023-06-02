#!/bin/bash

sudo ip link delete gtp0
sudo ip link delete gtp5g1
#sudo killall -9 gogtp5g-link
sudo gogtp5g-link add gtp5g1 30.103.11.21 --ran &
sleep 1
echo "1"
sudo gogtp5g-tunnel add far gtp5g1 1 --action 2
echo 2
sudo gogtp5g-tunnel add far gtp5g1 2 --action 2 --hdr-creation 0 1978933400 30.103.11.22 2152
echo 3
sudo gogtp5g-tunnel add pdr gtp5g1 1 --pcd 1 --hdr-rm 1 --ue-ipv4 30.103.12.23 --f-teid 1 30.103.11.21 --far-id 1
echo 4
sudo gogtp5g-tunnel add pdr gtp5g1 2 --pcd 2 --ue-ipv4 30.103.12.23 --far-id 2

echo 5
sudo ip r add 30.103.12.3/32 dev gtp5g1
sudo ip a add 30.103.12.23/32 dev gtp5g1
