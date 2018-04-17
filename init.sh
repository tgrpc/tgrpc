$GOPATH/src/github.com/toukii/ngrpc/nginx.sh &

go run $GOPATH/src/github.com/toukii/ngrpc/serve.go &

go test -v -test.run TestInvokeGRpcGreeter
