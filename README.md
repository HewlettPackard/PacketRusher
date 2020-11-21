# my5G-RANTester: NGAP/NAS Test tool.

My5g-RANTester is a GNB/UE simulator for testing 3GPP standard and stress in 5G core.

## Getting started

### Prerequisites

- my5gRANTester is built and tested with Go 1.14.4.

### Install

1- Downloading source code:
  - ```git clone https://github.com/LABORA-INF-UFG/my5G-RANTester.git ```
  - ```cd my5g-RANTester ```
  - ```go mod download ```
  
2- Build binary:
  - ```cd cmd ```
  - ```go build app.go```
  
3- Edit configuration file in config/config.yml:

  - Change amfif with AMF ip and port(N2).
  - Change amif with core name that you are testing. In the moment we have two options: free5gcore or open5gs.
  - Change upif with UPF ip, port(N3).
  - Check the values in ue(opc,key,amf). This values must be registered by webconsole core and my5gRANTester will use them in all tests.
  - Keep attention about imsi because some tests was automatized(load-tests) and will not permit change. Read more below.
  

## How can you use my5G-RANTester ?

### Running with template:
  - ```cd cmd ```
  - ```./app <with flag that identify type of test>```

### We have now different types of test for testing some kinds of behaviors from Core, that are show below:

  1- Load-test with UEs in queue. You can use the command to test: 
            ```./app load-test -n <number of ues> ```
        
    - Important in testing, the imsi UEs was automatized and you have to include them in web UI of test core as show below:
    * Example: if you want to test 10 ues you have to included imsi UE range to 2089300000001 from 2089300000010 in web UI. You can change
    other values in config/config.yml for example opc,k as you interest and used them in testing but imsi and hplmn will follow the range above.
    * Example: if you want to test 2 ues you have to include imsi 2089300000001 and 2089300000002 in web UI of test core.
   
  2- Load-test with UEs attached at the same time(concurrency) using a gnb. You can use the command to test with 3 ues: 
              ```./app load-test -t -n 3 ```
    
    - Important in testing, the imsi UEs was automatized and you have to include them in web UI of test core as show below:
    * Example: if you want to test 2 ues you have to include imsi 2089300000001 and 2089300000002 in web UI of test core. You can change
    other values in config/config.yml for example opc,k. But the imsi and hplm will follow the range above.
    
  3- Load-test with ues attached at the same time(concurrency) using a gnb per ue. You can use the command to test with 3 ues with 3 gnbs: 
              ```./app load-test -g -n 3  ```
     
     - You can change plmn list in config/config.yml and tester will use your values within all emulate gnbs.
     - Important in testing, the imsi UEs was automatized and you have to include them in web UI from the test core as show below:
     * Example: if you want to test 2 ues you have to include imsi 2089300000001 and 2089300000002 in web UI of test core.  You can change
    other values in config/config.yml. For example opc,k. But the imsi and will follow the range above.
    
  4- Load-test with gnbs. You can use the command to test 10 gnbs attach to core: 
                ```./app gnb -g 10 ```
    
    - You can change all configurations in config/config.yml as your interest.
    
  5- Test ue and gnb. You can use command to test: 
                  ```./app ue ```
     
     - You can change all configurations in config/config.yml and test with an ue and a gnb.
   


