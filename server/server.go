package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

type Server struct {
	*grpc.Server
	config   Config
	stopChan chan bool
}

func (s *Server) Serve() (ec <-chan error) {
	errChan := make(chan error, 1)
	ec = errChan

	if s.stopChan != nil {
		errChan <- errors.New("already started")
		close(errChan)
		return
	}

	s.stopChan = make(chan bool)

	grpcErrChan := s.serveGrpc(s.stopChan)
	httpErrChan := s.serveHttp(s.stopChan)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for err := range httpErrChan {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		for err := range grpcErrChan {
			errChan <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	return
}

func (s *Server) serveGrpc(stopChan <-chan bool) (ec <-chan error) {
	errChan := make(chan error, 2)
	ec = errChan

	if !s.config.GrpcEnable {
		close(errChan)
		return
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.GrpcHost, s.config.GrpcPort))
	if err != nil {
		errChan <- err
		close(errChan)
		return
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer close(signalChan)

		select {
		case <-signalChan:
		case <-stopChan:
		}

		s.Server.GracefulStop()
	}()

	go func() {
		defer close(errChan)

		errChan <- s.Server.Serve(lis)
	}()

	return
}

func (s *Server) serveHttp(stopChan <-chan bool) (ec <-chan error) {
	errChan := make(chan error, 2)
	ec = errChan

	if !s.config.WebEnable {
		close(errChan)
		return
	}

	listenErrChan := make(chan error, 1)
	stopErrChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		for err := range listenErrChan {
			errChan <- err
		}
		for err := range stopErrChan {
			errChan <- err
		}
	}()

	wrappedServer := grpcweb.WrapServer(
		s.Server,
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
		Addr:    fmt.Sprintf("%s:%d", s.config.WebHost, s.config.WebPort),
		Handler: http.HandlerFunc(handler),
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer close(signalChan)
		defer close(stopErrChan)

		select {
		case <-signalChan:
		case <-stopChan:
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			stopErrChan <- err
		}
	}()

	if s.config.WebEnableTls {
		go func() {
			defer close(listenErrChan)

			// TODO: tls
			err := httpServer.ListenAndServeTLS(testdata.Path("server1.pem"), testdata.Path("server1.key"))
			if err != http.ErrServerClosed {
				listenErrChan <- err
			}
		}()
	} else {
		go func() {
			defer close(listenErrChan)

			err := httpServer.ListenAndServe()
			if err != http.ErrServerClosed {
				listenErrChan <- err
			}
		}()
	}

	return
}

func NewGrpcForServers(ctx context.Context, cfg Config, servers []Registrar) (s *Server, err error) {
	var opts []grpc.ServerOption
	// TODO: tls
	if cfg.GrpcEnableTls {
		//if *certFile == "" {
		//	*certFile = testdata.Path("server1.pem")
		//}
		//if *keyFile == "" {
		//	*keyFile = testdata.Path("server1.key")
		//}
		//creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		creds, creadErr := credentials.NewServerTLSFromFile(testdata.Path("server1.pem"), testdata.Path("server1.key"))
		if creadErr != nil {
			err = errors.Wrap(creadErr, "failed to generate credentials")
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}
	grpcServer := grpc.NewServer(opts...)
	for _, srv := range servers {
		err = srv.Register(ctx, grpcServer)
		if err != nil {
			return
		}
	}

	s = &Server{
		Server: grpcServer,
		config: cfg,
	}
	return
}
