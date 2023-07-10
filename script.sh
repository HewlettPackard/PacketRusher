sudo rm /tmp/9*.sock
go mod download
go build cmd/app.go && sudo ./app ue
