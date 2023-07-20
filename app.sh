sudo rm /tmp/9*.sock
go mod download
go build cmd/app.go && sudo ./app multi-ue -n 1000 --loop -tr 5 -td 5000

