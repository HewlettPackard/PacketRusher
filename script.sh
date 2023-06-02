sudo ip link delete gtp0
sudo rm /tmp/9*.sock
go build cmd/app.go && sudo ./app ue
