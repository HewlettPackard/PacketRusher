We have now different types of test for testing some kinds of behaviors from Core, that are show below.

Keep attention about imsi because some tests was automatized(load-tests) and will not permit change. If you are using free5gcore you can use script in /dev/includes_ues.sh to added imsi and other information by webconsole. As show below:
```
   ./included_ues.sh -n <number of UEs that you want to test in load tests>  
```

- Load-test with UEs in queue*
    - You can use the command to test with number of UEs:
            ```
              ./app load-test -n <number of UEs>
            ```
    - For example for testing with 3 UEs:
            ```
              ./app load-test -n 3
            ```

- Load-test with UEs attached at the same time(concurrency) using a unique GNB*
    - You can use the command to test with number of UEs:
                ```
                 ./app load-test -t -n <number of UEs>
                ```
    - You can use the command to test with 3 UEs:
              ```
                ./app load-test -t -n 3
              ```

- Load-test with UEs attached at the same time(concurrency) using a GNB per UE*
    - You can use the command to test with number of UEs:
             ``` ./app load-test -g -n <number of ues> ```
    - You can use the command to test with 3 GNBs, each one with an UE:
             ``` ./app load-test -g -n 3 ```

- Load test with UEs attached at the same time based a poisson and exponential distribution*
  - You can use the command to test:
           ```
           ./app nlinear-tests -s <samples> -mu <mean> -se <seed>
           ```
  - Each samples is a random number generate by poisson distribution and means number of UEs attached to Core at the same time.
  - Between each sample has a time duration in seconds defined by Exponential Distribution
  - Mean and seed is used by Poisson and exponential distribution for generate random numbers.
  
- Load-test with GNBs
    - You can use the command to test 10 GNBs attached to core:
              ``` ./app gnb -n 10  ```
    - Configurations in config/config.yml.

- Test with an UE and a GNB.
    - You can use command to test:
             ``` ./app ue ```
    - Configurations in config/config.yml.
    
## Additional comments


For tests with * imsi UEs was automatized and you have to include them in web UI of test core as show below.

Example: if you want to test 10 UEs you have to included imsi UE range to 2089300000001 from 2089300000010 in web UI. You can change other values in config/config.yml for example opc,k as you interest and used them in testing but imsi and hplmn will follow the range above.
  <p align="">
     <img src="docs/media/img/ues10.png"/>
  </p>
  
Example: if you want to test 2 UEs you have to include imsi 2089300000001 and 2089300000002 in web UI of test core.
 <p align="">
     <img src="docs/media/img/ue_configuration.png"/>
 </p>
