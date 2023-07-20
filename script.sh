sudo rm /tmp/9*.sock
go mod download
go build cmd/app.go && sudo ./app c --scenario /home/athomgmt/my5G-RANTester/scenarios/scenario1/scenario1.wasm
