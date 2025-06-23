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

#### 2. Add logging
```bash
go build -o controller
./controller --log-level info
./controller --log-level debug
./controller --log-level trace
```

#### 2. Add FastHTTP server
```bash
go build -o controller
./controller server --log-level debug
```