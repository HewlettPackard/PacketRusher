# my5G-RANTester


![GitHub](https://img.shields.io/github/license/LABORA-INF-UFG/my5G-RANTester?color=blue) 
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/LABORA-INF-UFG/my5G-RANTester) ![GitHub commit activity](https://img.shields.io/github/commit-activity/y/LABORA-INF-UFG/my5G-RANTester) 
![GitHub last commit](https://img.shields.io/github/last-commit/LABORA-INF-UFG/my5G-RANTester)
![GitHub contributors](https://img.shields.io/github/contributors/LABORA-INF-UFG/my5G-RANTester)

<img width="20%" src="docs/media/img/my5g-logo.png" alt="my5g-core"/>

----

my5G-RANTester is a tool that simulates User Equipment (UE) and Next Generation-Radio Access Networks (NG-RANs) for testing any 5G cores on the 3GPP standard (i.e., Release 15). This tool provides stress tests for the control plane and data plane. Moreover, the tests can be played individually for analyzing only one UE or in scale for analyzing hundreds of UE simultaneously.

----
## Installation

**Requirements**

The software requirement:
* Go 1.14.4
* GSL 2.6
* GCC

The installation can be done directly over the host operating system (OS) or inside a virtual machine (VM). System requirements:
* CPU type: x86-64 (specific model and number of cores only affect performance)
* RAM: 1 GB
* Ubuntu 18.04/20.04 LTS.

**Steps**

Install GSL-2.6 and GCC
```
sudo apt update
sudo apt install build-essential
sudo apt install install libgsl-dev
```

Downloading source code:
```
git clone https://github.com/my5G/my5G-RANTester.git
```

Install dependencies:
```
cd my5g-RANTester
go mod download
```
  
Build binary:
```
cd cmd 
go build app.go
```
  
Edit configuration file in config/config.yml:

Change amfif with AMF ip, port and core name that you are testing. In the moment we have two options: free5gcore or open5gs
```yaml
amfif:
  ip: "127.0.0.1"
  port: 38412
  name: "free5gc"
```

Check the values in UE(opc,key,amf). This values must be registered by webconsole core and my5gRANTester will use them in all tests.
```yaml
  key: "70d49a71dd1a2b806a25abe0ef749f1e"
  opc: "6f1bf53d624b3a43af6592854e2444c7"
  amf: "8000"
```

Keep attention about imsi because some tests was automatized(load-tests) and will not permit change. If you are using free5gcore you can using script in /dev/includes_ues.sh to added imsi and other information by webconsole. As show below:
```
   ./included_ues.sh -n <number of UEs that you want to test in load tests>  
```

  
## Tests

**Running with template:**
```
cd cmd
./app ue
```

**Check interface uetun1 that will test data plane with ping or other tool.**

# Questions
 
For questions and support please send a e-mail message to [my5G team](mailto:my5G.initiative@gmail.com). 

