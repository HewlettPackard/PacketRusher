# my5G-RANTester


![GitHub](https://img.shields.io/github/license/LABORA-INF-UFG/my5G-RANTester?color=blue) 
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/LABORA-INF-UFG/my5G-RANTester) ![GitHub commit activity](https://img.shields.io/github/commit-activity/y/LABORA-INF-UFG/my5G-RANTester) 
![GitHub last commit](https://img.shields.io/github/last-commit/LABORA-INF-UFG/my5G-RANTester)
![GitHub contributors](https://img.shields.io/github/contributors/LABORA-INF-UFG/my5G-RANTester)

<img width="20%" src="docs/media/img/my5g-logo.png" alt="my5g-core"/>

- [Description](#description)
- [Installation](#installation)
    - [Recommended environment](#recommended-environment)
    - [Requirements](#requirements)
    - [RAN Tester](#ran-tester)
- [Check](#check)
- [More information](#more-information)

----
# Description

my5G-RANTester is a tool for emulating control and data planes of the UE (user equipment) and gNB (5G base station). my5G-RANTester follows the 3GPP Release 15 standard for NG-RAN (Next Generation-Radio Access Network). Using my5G-RANTester, it is possible to generate different workloads and test several functionalities of a 5G core, including its complaince to the 3GPP standards. Scalability is also a relevant feature of the my5G-RANTester, which is able mimic the behaviour of a large number of UEs and gNBs accessing simultaneously a 5G core. Currently, the wireless channel is not implemented in the tool.

If you want to cite this tool, please use the following information:
```
@misc{???,
    title={???},
    author={???},
    year={2020},
    eprint={???},
    archivePrefix={arXiv},
    primaryClass={cs.NI}
}
```
If you have questions or comments, please email us: [my5G team](mailto:my5G.initiative@gmail.com). 

my5G-RANTester borrows libraries and data structures from the [free5gc project](https://github.com/free5gc/free5gc).

----
# Installation

## Recommended environment

* CPU type: x86-64 (specific model and number of cores only affect performance)
* RAM: 1 GB
* Operating System (OS): Linux Ubuntu 18.04 or 20.04 LTS.

The installation can be done directly in the host OS or inside a virtual machine (VM).


## Requirements

Install GCC:
```
sudo apt update && apt -y install build-essential
```

Install GSL:
```
sudo apt -y install libgsl-dev
```

Install Go:
```
???
```

## RAN Tester

Download the source code:
```
git clone https://github.com/LABORA-INF-UFG/my5G-RANTester.git
```

Install the dependencies:
```
cd my5G-RANTester
go mod download
```
  
Build the binary:
```
cd cmd 
go build app.go
```

Edit configuration file **config/config.yml** and make the following procedures.

1. Configure **amfif** with the AMF IP address, port and core name that you are testing. Currrently, there are three options for **name**: my5g-core, free5gc or open5gs.
```yaml
amfif:
  ip: "127.0.0.1"
  port: 38412
  name: "free5gc"
```

2. Change **upif** with UPF IP address and port (N3 reference point).
```yaml
upfif:
  ip: "10.200.200.102"
  port: 2152
```

3. Check the values in UE (key,opc,amf), as illustrated below. The same values must be registered in the 5G core under evaluation since they are used by the my5G-RANTester during the tests.
```yaml
  key: "70d49a71dd1a2b806a25abe0ef749f1e"
  opc: "6f1bf53d624b3a43af6592854e2444c7"
  amf: "8000"
```

Done! The software is successfully installed.

----
# Check

Make sure that the path to GSL dynamic library is part of environment variable LD_LIBRARY_PATH:
```
LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib/gsl/
export LD_LIBRARY_PATH
```

Make also sure that you are in directory **my5G-RANTester/cmd**, then run the most basic test:
```
./app ue
```

The result should be similar to the following:
```
???
```

# More information
   
If my5G-RANTester was successfuly installed, you may find useful take a look in the USAGE manual that also presents several examples.

If you found problems during the installation or during the checking, you may find a solution in the TROUBLESHOOTING document.

If you do not have a 5G core for testing, you can easily deploy and run the [my5G-core](https://github.com/my5G/my5G-core).
