#!/bin/sh

HOST=localhost
PORT=5000

BASE_URL="http://$HOST:$PORT/api/subscriber"
K=70d49a71dd1a2b806a25abe0ef749f1e
OPC=6f1bf53d624b3a43af6592854e2444c7

i=0

imsi=2089300000000
plmnid=20893

while [ $i -lt $1 ] ;
do
    imsi=$((imsi+1))
    curl "http://localhost:5000/api/subscriber/imsi-$imsi/$plmnid" \
      -H 'Accept: application/json' \
      -H 'X-Requested-With: XMLHttpRequest' \
      -H 'User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36' \
      -H 'Content-Type: application/json;charset=UTF-8' \
      --data-binary '{"plmnID":"'$plmnid'","ueId":"imsi-'$imsi'","AuthenticationSubscription":{"authenticationManagementField":"8000","authenticationMethod":"5G_AKA","milenage":{"op":{"encryptionAlgorithm":0,"encryptionKey":0,"opValue":""}},"opc":{"encryptionAlgorithm":0,"encryptionKey":0,"opcValue":"'$OPC'"},"permanentKey":{"encryptionAlgorithm":0,"encryptionKey":0,"permanentKeyValue":"'$K'"},"sequenceNumber":"16f3b3f70fc2"},"AccessAndMobilitySubscriptionData":{"gpsis":["msisdn-0900000000"],"nssai":{"defaultSingleNssais":[{"sd":"010203","sst":1},{"sd":"112233","sst":1}],"singleNssais":[{"sd":"010203","sst":1},{"sd":"112233","sst":1}]},"subscribedUeAmbr":{"downlink":"2 Gbps","uplink":"1 Gbps"}},"SessionManagementSubscriptionData":{"singleNssai":{"sst":1,"sd":"010203"},"dnnConfigurations":{"internet":{"sscModes":{"defaultSscMode":"SSC_MODE_1","allowedSscModes":["SSC_MODE_1","SSC_MODE_2","SSC_MODE_3"]},"pduSessionTypes":{"defaultSessionType":"IPV4","allowedSessionTypes":["IPV4"]},"sessionAmbr":{"uplink":"2 Gbps","downlink":"1 Gbps"},"5gQosProfile":{"5qi":9,"arp":{"priorityLevel":8},"priorityLevel":8}}}},"SmfSelectionSubscriptionData":{"subscribedSnssaiInfos":{"01010203":{"dnnInfos":[{"dnn":"internet"}]},"01112233":{"dnnInfos":[{"dnn":"internet"}]}}},"AmPolicyData":{"subscCats":["free5gc"]},"SmPolicyData":{"smPolicySnssaiData":{"01010203":{"snssai":{"sst":1,"sd":"010203"},"smPolicyDnnData":{"internet":{"dnn":"internet"}}},"01112233":{"snssai":{"sst":1,"sd":"112233"},"smPolicyDnnData":{"internet":{"dnn":"internet"}}}}}}' \
      --compressed

    i=$((i+1))
done
