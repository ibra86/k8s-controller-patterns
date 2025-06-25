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

#### 3. Add CI-CD
```bash
make run #default run
make run ARGS="server --log-level debug" #run with arguments
make build
docker run -p 8080:8080 k8s-controller-patterns:latest server --log-level debug

export GITHUB_PAT=<GITHUB_TOKEN>
echo $GITHUB_PAT | docker login ghcr.io -u ibra86 --password-stdin

kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=ibra86 \
  --docker-password=$GITHUB_PAT \
  --docker-email=<user-email> \
  --dry-run=client -o yaml > secret.yaml
kubectl apply -f secret.yaml
helm install k8s-controllers ./charts/app
kubectl port-forward service/k8s-controllers 8080:80
curl http://localhost:8080
```

#### 4. client-go api
```bash
# list deployments
go run main.go list --log-level debug --kubeconfig ~/.kube/config--log-level debug
```