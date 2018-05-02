# tgrpc

tg: tgrpc in termial
============

为了方便测试，可开启grpc服务代理，参考[github.com/tgrpc/ngrpc](https://github.com/tgrpc/ngrpc)。

使用方法，请看[tgrpc/doc](https://github.com/tgrpc/doc)

## usage

```
tg -c tgrpc.toml
```

```
DEBU[0000] dial tcp:localhost:2080 ...
DEBU[0000] use exist desc                                FileDescriptorSet=.helloworld.Greeter.pbin
ERRO[0000] time cost: 3.41066ms more than 2 ms;
DEBU[0000] dial tcp:localhost:2080 ...
DEBU[0000] use exist desc                                FileDescriptorSet=.helloworld.LangService.pbin
DEBU[0000] time cost: 1.007471ms / 2 ms;
DEBU[0000] time cost: 1.008415ms / 2 ms;
WARN[0000] time cost: 1.849676ms nearby 2 ms;
WARN[0000] time cost: 1.921683ms nearby 2 ms;
WARN[0000] time cost: 1.967108ms nearby 2 ms;
WARN[0001] time cost: 1.545166ms nearby 2 ms;
WARN[0001] time cost: 1.77252ms nearby 2 ms;
WARN[0001] time cost: 1.63466ms nearby 2 ms;
helloworld.Greeter/SayHello size: 3
 avg: 1.863101 ms
 max: 1.967108 ms
 min: 1.772520 ms
```

```
tg -s
```

```
DEBU[0000] dial tcp:localhost:2080 ...
DEBU[0000] use exist desc                                FileDescriptorSet=.helloworld.Greeter.pbin
ERRO[0000] time cost: 4.05723ms more than 2 ms;
helloworld.Greeter/SayHello ==> {
	"message": "Hello tgrpc-tg1"
}
```

## todo

1. 调用链状态显示

## roadmap

1. 验证接口返回结果
2. 验证接口响应时间
3. 方法调用闭环，支持利用上次返回结果构建参数
4. 接口响应时间统计：最大、最小、平均耗时

## origin

部分代码参考[github.com/uber/prototool](https://github.com/uber/prototool)
