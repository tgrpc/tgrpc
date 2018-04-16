# tgrpc

## tg: tgrpc in termial

```
go get github.com/tgrpc/tgrpc/tg
tg -i
tg [-c tgrpc.toml]
```

## test

```
ln -s $GOPATH/src/github.com/toukii/ngrpc/helloworld helloworld

$GOPATH/src/github.com/toukii/ngrpc/nginx.sh

go run $GOPATH/src/github.com/toukii/ngrpc/serve.go

go test -v -test.run TestInvokeGRpcGreeter
```

## origin

部分代码参考[github.com/uber/prototool](https://github.com/uber/prototool)
