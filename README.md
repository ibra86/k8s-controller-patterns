# k8s-controller-patterns

#### 0. Set-up control plane in Codespace via script
```bash
./setup-amd64.sh start
./setup-amd64.sh stop
./setup-amd64.sh cleanup
```

#### 1. Add go basics + test
```bash
go run main.go go-basic
go test ./cmd
go build -o controller
./controller go-basic
```