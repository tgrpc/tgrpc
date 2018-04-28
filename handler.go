package tgrpc

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	uuid "github.com/dchest/uniuri"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/toukii/bytes"
	"github.com/toukii/goutils"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	Silence bool
)

type invocationEventHandler struct {
	vf           Verifier
	method       string
	ivk, nextIvk *Invoke
}

func newInvocationEventHandler(vf Verifier, method string, ivk, nextIvk *Invoke) *invocationEventHandler {
	handler := &invocationEventHandler{
		method:  method,
		ivk:     ivk,
		nextIvk: nextIvk,
	}
	ivk.Do(ivk.Init)
	if !reflect.ValueOf(vf).IsNil() {
		handler.vf = vf
	}
	return handler
}

// md 是本次grpc请求的header，非请求的meta
func (i *invocationEventHandler) OnResolveMethod(desc *desc.MethodDescriptor) {
}

func (i *invocationEventHandler) OnSendHeaders(md metadata.MD) {
	now := time.Now()
	trackId := uuid.New()
	md["_tgrpc"] = []string{fmt.Sprintf("start_time=%d", now.UnixNano()), fmt.Sprintf("track_id=%s", trackId)}
}

func (i *invocationEventHandler) OnReceiveHeaders(md metadata.MD) {
	i.verifyCost(md)
}

func (i *invocationEventHandler) OnReceiveResponse(md metadata.MD, message proto.Message) {
	wr := bytes.NewWriter(make([]byte, 0, 1024))
	err := jsonpbMarshaler.Marshal(wr, message)
	if isErr(err) {
		return
	}
	bs := wr.Bytes()
	if i.nextIvk != nil {
		i.nextIvk.preResp <- bs
	}
	if i.vf != nil {
		i.vf.Verify(bs)
	}
	// log.WithField(i.method, goutils.ToString(bs)).Info()
	if !Silence {
		fmt.Printf("%s ==> %s\n", i.method, goutils.ToString(bs))
	}
}

func getStartTime(md metadata.MD) (int64, error) {
	mds := md["_tgrpc"]
	for _, v := range mds {
		if strings.HasPrefix(v, "start_time=") {
			return strconv.ParseInt(strings.Split(v, "start_time=")[1], 10, 64)
		}
	}
	return 0, fmt.Errorf("start_time is not set.")
}

func (i *invocationEventHandler) OnReceiveTrailers(s *status.Status, md metadata.MD) {
	if err := s.Err(); err != nil {
		log.WithField(fmt.Sprintf("%s--OnReceiveTrailers", i.method), err).Error()
	}
}

func (i *invocationEventHandler) verifyCost(md metadata.MD) {
	if i.vf != nil {
		startTime, err := getStartTime(md)
		if !isErr(err) {
			ns := time.Now().UnixNano() - startTime
			i.vf.VerifyCost(ns, i.ivk.Costch)
		}
	}
}

func Caller(depth int) (string, string, int) {
	pc, file, line, ok := runtime.Caller(1 + depth)
	if !ok {
		return "", "", -1
	}

	return runtime.FuncForPC(pc).Name(), path.Base(file), line
}
