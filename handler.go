package tgrpc

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/toukii/bytes"
	"github.com/toukii/goutils"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type invocationEventHandler struct {
	vf Verifier
}

func newInvocationEventHandler(vf Verifier) *invocationEventHandler {
	return &invocationEventHandler{vf: vf}
}

func (i *invocationEventHandler) OnResolveMethod(desc *desc.MethodDescriptor) {
	log.Debugf("OnResolveMethod: %+v", desc.GetName())
}

func (i *invocationEventHandler) OnSendHeaders(md metadata.MD) {
	now := time.Now()
	md["_tgrpc"] = []string{fmt.Sprintf("start_time=%d", now.UnixNano()), fmt.Sprintf("track_id=%s", "TRACKID")}
	log.Debugf("OnSendHeaders: %+v", md)
}

func (i *invocationEventHandler) OnReceiveHeaders(md metadata.MD) {
	log.Debugf("OnReceiveHeaders: %+v", md)
}

func (i *invocationEventHandler) OnReceiveResponse(md metadata.MD, message proto.Message) {
	now := time.Now()
	wr := bytes.NewWriter(make([]byte, 0, 1024))
	err := jsonpbMarshaler.Marshal(wr, message)
	if isErr(err) {
		return
	}
	bs := wr.Bytes()
	if !reflect.ValueOf(i.vf).IsNil() {
		startTime, err := getStartTime(md)
		cost := int64(0)
		if !isErr(err) {
			cost = now.UnixNano() - startTime
		}
		i.vf.Verify(bs, cost)
	}
	log.Debugf("OnReceiveResponse md: %+v %+v", md, now)
	log.WithField("OnReceiveResponse", goutils.ToString(bs)).Info()
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

func (i *invocationEventHandler) OnReceiveTrailers(s *status.Status, _ metadata.MD) {
	if err := s.Err(); err != nil {
		log.WithField("OnReceiveTrailers", err).Error()
	}
}

func Caller(depth int) (string, string, int) {
	pc, file, line, ok := runtime.Caller(1 + depth)
	if !ok {
		return "", "", -1
	}

	return runtime.FuncForPC(pc).Name(), path.Base(file), line
}
