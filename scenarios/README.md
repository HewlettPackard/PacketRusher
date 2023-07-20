# Custom scenarios
## Introduction
PacketRusher's custom scenarios are built using WebAssembly, all WebAssembly language that supports WASI should work for the purpose of writing custom scenarios.
For now, only a single UE can be managed using a Custom Scenario, and very few functions are exposed to the custom scenario.
Custom scenarios are highly WIP, and function used in the scenario WILL change.

## Usage
You can run the `./build.sh` to build custom scenarios in Go using tinygo. Each custom scenario must be in their dedicated folder.

You can see the sample directory for a sample scenario.
Most of the Go standard library can be used in a custom scenario, but issues may arise depending on the functions used.

Once the scenario has been built into a .wasm file, it can be run using PacketRusher's custom-scenario CLI:
```bash
./app custom-scenario --scenario sample.go.wasm
```

You can also reduce log level from 4 to 3 in config.yml if you are unable to see your fmt.Println() because there are too much logs :D

## State

Custom scenarios are WIP, and function names will change.