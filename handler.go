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

type invocationEventHandler struct {
	vf     Verifier
	method string
	ivk    *Invoke
}

func newInvocationEventHandler(vf Verifier, method string, ivk *Invoke) *invocationEventHandler {
	if !reflect.ValueOf(vf).IsNil() {
		return &invocationEventHandler{vf: vf, method: method, ivk: ivk}
	}
	return &invocationEventHandler{vf: nil, method: method, ivk: ivk}
}

// md 是本次grpc请求的header，非请求的meta
func (i *invocationEventHandler) OnResolveMethod(desc *desc.MethodDescriptor) {
	// log.Debugf("OnResolveMethod: %+v", desc.GetName())
}

func (i *invocationEventHandler) OnSendHeaders(md metadata.MD) {
	now := time.Now()
	trackId := uuid.New()
	md["_tgrpc"] = []string{fmt.Sprintf("start_time=%d", now.UnixNano()), fmt.Sprintf("track_id=%s", trackId)}
	// log.Debugf("OnSendHeaders: %+v", md)
}

func (i *invocationEventHandler) OnReceiveHeaders(md metadata.MD) {
	// log.Debugf("OnReceiveHeaders: %+v", md)
	i.verifyCost(md)
}

func (i *invocationEventHandler) OnReceiveResponse(md metadata.MD, message proto.Message) {
	wr := bytes.NewWriter(make([]byte, 0, 1024))
	err := jsonpbMarshaler.Marshal(wr, message)
	if isErr(err) {
		return
	}
	bs := wr.Bytes()
	if i.ivk != nil {
		i.ivk.preResp <- bs
	}
	if i.vf != nil {
		i.vf.Verify(bs)
	}
	// log.Debugf("OnReceiveResponse md: %+v", md)
	log.WithField(i.method, goutils.ToString(bs)).Info()
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
		log.WithField("OnReceiveTrailers", err).Error()
	}
}

func (i *invocationEventHandler) verifyCost(md metadata.MD) {
	if i.vf != nil {
		startTime, err := getStartTime(md)
		if !isErr(err) {
			i.vf.VerifyCost(time.Now().UnixNano() - startTime)
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
