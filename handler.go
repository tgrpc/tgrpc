package tgrpc

import (
	"path"
	"reflect"
	"runtime"

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
	log.Debugf("OnSendHeaders: %+v", md)
}

func (i *invocationEventHandler) OnReceiveHeaders(md metadata.MD) {
	log.Debugf("OnReceiveHeaders: %+v", md)
}

func (i *invocationEventHandler) OnReceiveResponse(md metadata.MD, message proto.Message) {
	wr := bytes.NewWriter(make([]byte, 0, 1024))
	err := jsonpbMarshaler.Marshal(wr, message)
	if isErr(err) {
		return
	}
	bs := wr.Bytes()
	if !reflect.ValueOf(i.vf).IsNil() {
		i.vf.Verify(bs, 0, 0)
	}
	log.Debugf("OnReceiveResponse md: %+v", md)
	trackId := md["trackid"]
	log.WithField("OnReceiveResponse", goutils.ToString(bs)).Infof("trackId: %+v", trackId)
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
