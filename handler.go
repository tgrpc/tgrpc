package tgrpc

import (
	"path"
	"runtime"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type invocationEventHandler struct{}

func newInvocationEventHandler() *invocationEventHandler {
	return &invocationEventHandler{}
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
	s, err := jsonpbMarshaler.MarshalToString(message)
	if isErr(err) {
		return
	}
	log.Debugf("OnReceiveResponse md: %+v", md)
	trackId := md["trackid"]
	log.WithField("OnReceiveResponse", s).Infof("trackId: %+v", trackId)
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
