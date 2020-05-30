package server

import (
	"context"
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/testdata"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Registrar interface {
	Register(ctx context.Context, grpcServer *grpc.Server) error
}

type Server struct {
	*grpc.Server
}

func NewGrpcForServers(servers []Registrar) (*Server, error) {
	var opts []grpc.ServerOption
	// TODO: make gRPC via tls
	//if *tls {
	//	if *certFile == "" {
	//		*certFile = testdata.Path("server1.pem")
	//	}
	//	if *keyFile == "" {
	//		*keyFile = testdata.Path("server1.key")
	//	}
	//	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	//creds, err := credentials.NewServerTLSFromFile(testdata.Path("server1.pem"), testdata.Path("server1.key"))
	//if err != nil {
	//	log.Fatalf("Failed to generate credentials %v", err)
	//}
	//opts = append(opts, grpc.Creds(creds))
	//}
	grpcServer := grpc.NewServer(opts...)
	for _, srv := range servers {
		err := srv.Register(nil, grpcServer)
		if err != nil {
			return nil, err
		}
	}

	return &Server{grpcServer}, nil
}

func (srv *Server) Serve() chan error {
	errChan := make(chan error, 1)

	enableTls := true
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 24560))
	if err != nil {
		errChan <- err
		close(errChan)
		return errChan
	}

	wrappedServer := grpcweb.WrapServer(
		srv.Server,
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}),
		grpcweb.WithWebsockets(true),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool {
			return true
		}),
		grpcweb.WithWebsocketPingInterval(10),
	)

	handler := func(resp http.ResponseWriter, req *http.Request) {
		if wrappedServer.IsGrpcWebSocketRequest(req) {
			wrappedServer.HandleGrpcWebsocketRequest(resp, req)
		} else {
			wrappedServer.ServeHTTP(resp, req)
		}
	}

	httpServer := http.Server{
		Addr:    fmt.Sprintf(":%d", 9090),
		Handler: http.HandlerFunc(handler),
	}

	signalChan := make(chan os.Signal, 1)
	grpcSignalChan := make(chan os.Signal, 1)
	grpcErrChan := make(chan error, 1)
	srvErrChan := make(chan error, 1)
	srvStopErrChan := make(chan error, 1)

	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	signal.Notify(grpcSignalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			srvStopErrChan <- err
		}
		close(srvStopErrChan)
	}()
	go func() {
		<-grpcSignalChan
		srv.Server.Stop()
	}()
	go func() {
		for err := range srvErrChan {
			errChan <- err
		}
		for err := range grpcErrChan {
			errChan <- err
		}
		for err := range srvStopErrChan {
			errChan <- err
		}
		close(errChan)
	}()

	if enableTls {
		go func() {
			err := httpServer.ListenAndServeTLS(testdata.Path("server1.pem"), testdata.Path("server1.key"))
			if err != http.ErrServerClosed {
				srvErrChan <- err
			}
			close(srvErrChan)
		}()
	} else {
		go func() {
			err := httpServer.ListenAndServe()
			if err != http.ErrServerClosed {
				srvErrChan <- err
			}
			close(srvErrChan)
		}()
	}

	go func() {
		grpcErrChan <- srv.Server.Serve(lis)
		close(grpcErrChan)
	}()

	return errChan
}
