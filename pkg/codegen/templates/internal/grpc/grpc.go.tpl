package {{ .manifest.Name }}

import (
	"context"
	"net"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	apiv1 "github.com/tritonmedia/{{ .manifest.Name }}/api/v1"
)

type GRPCService struct {
	lis net.Listener
	srv *grpc.Server
}

func NewGRPCService() *GRPCService {
	return &GRPCService{}
}

// Run starts a grpc server with the internal server handler
func (g *GRPCService) Run(ctx context.Context, log logrus.FieldLogger) error {
	listAddr := ":" + strconv.Itoa(8000)
	l, err := net.Listen("tcp", listAddr)
	if err != nil {
		return err
	}
	g.lis = l

	h, err := NewServiceHandler(ctx, log)
	if err != nil {
		return err
	}

	g.srv = grpc.NewServer()
	reflection.Register(g.srv)
	apiv1.RegisterAPIService(g.srv, apiv1.NewAPIService(h))

	// handle closing the server
	go func() {
		<-ctx.Done()
		log.Info("shutting down server")
		g.srv.GracefulStop()
	}()

	// One day Serve() will accept a context?
	log.Infof("starting GRPC server on %s", listAddr)
	return errors.Wrap(g.srv.Serve(g.lis), "unexpected grpc.Serve error")
}
