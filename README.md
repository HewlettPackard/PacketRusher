# my5G-RANTester: NGAP/NAS Test tool.


My5g-RANTester is a GNB/UE simulator for testing 3GPP standard and stress in 5G core.

# How can you use my5G-RANTester ?

We have now different types of tests for testing some kinds of behaviors from Core, that are show below:

  - Load-test with ues in queue. You can use the command to test with 10 ues: 
      * ./app load-test -n 10
    - Important in this testing that imsi ues was fixed and you have to included them in web UI from the test core as show below:
    Example: if you want to test 10 ues you have to subscribe imsi range to 2089300000001 from 2089300000010. You can change
    other values in config/config.yml for example opc,k. But the imsi and hplm will not change.
    Example: if you want to test 2 ues you have to subscribe imsi 2089300000001 and 2089300000002 in web UI from the test core.
   
  - Load-test with ues attach at the same time(concurrency) using one gnb. You can use the command to test with 3 ues:
      * ./app load-test -t -n 10
    - Important in this testing that imsi ues was fixed and you have to included them in web UI from the test core as show below:
    Example: if you want to test 2 ues you have to subscribe imsi 2089300000001 and 2089300000002 in web UI from the test core. You can change
    other values in config/config.yml for example opc,k. But the imsi and hplm does not change.
    
  - Load-test with ues attach at the same time(concurrency) using a gnb to ue. You can use the command to test with 3 ues with 3 gnbs:
      * ./app load-test -g -n 3
     - You can change plm list in config/config.yml and tester will use this values with all emulate gnbs.
     - Important in this testing that imsi ues was fixed and you have to included them in web UI from the test core as show below:
     Example: if you want to test 2 ues you have to subscribe imsi 2089300000001 and 2089300000002 in web UI from the test core.  You can change
    other values in config/config.yml for example opc,k. But the imsi and hplm does not change.
    
  - Load-test with gnbs. You can use the command to test 10 emulate gnbs attach to core:
    * ./app gnb -g 3
    
  - Test ue and gnb. You can use command to test:
     * ./app ue
     - You can change all configurations in config/config.yml and test with an ue and a gnb.
   


