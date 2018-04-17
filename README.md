# tgrpc

## tg: tgrpc in termial

```
go get github.com/tgrpc/tgrpc/tg
tg -i
tg [-c tgrpc.toml]
```

## test

```
# 开启服务
go run $GOPATH/src/github.com/tgrpc/ngrpc/server.go &
# 开启nginx代理grpc
$GOPATH/src/github.com/tgrpc/ngrpc/nginx.sh &
# 测试
go test -v -test.run TestInvokeGRpcGreeter
```

## origin

部分代码参考[github.com/uber/prototool](https://github.com/uber/prototool)
