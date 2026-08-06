package api

import (
	"fmt"
	"net"
	"net/http"

	pb "github.com/kevincouton/standard-tools-go/proto/health"
	"google.golang.org/grpc"
)

func Serve(state *AppState, httpPort, grpcPort int) error {
	errCh := make(chan error, 2)

	go func() {
		r := NewRouter(state)
		errCh <- http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r)
	}()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			errCh <- err
			return
		}
		s := grpc.NewServer()
		pb.RegisterHealthServer(s, &HealthServer{})
		errCh <- s.Serve(lis)
	}()

	return <-errCh
}
