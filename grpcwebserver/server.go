package grpcwebserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RunGrpcWebServer run grpc-web server.
func RunGrpcWebServer(ctx context.Context, s *grpc.Server, listenAddress string, allowedHeaders []string) {
	zap.L().Info("GrpcWeb server starting.")
	headers := []string{
		"x-grpc-web",
		"content-type",
		"content-length",
		"accept-encoding",
	}

	headers = append(headers, allowedHeaders...)

	opts := []grpcweb.Option{
		// gRPC-Web compatibility layer with CORS configured to accept on every request
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithWebsockets(true),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool {
			return true
		}),
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}),
	}
	wrappedGrpc := grpcweb.WrapServer(s, opts...)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if wrappedGrpc.IsAcceptableGrpcCorsRequest(req) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ", "))
			return
		}
		if wrappedGrpc.IsGrpcWebSocketRequest(req) || wrappedGrpc.IsGrpcWebRequest(req) {
			wrappedGrpc.ServeHTTP(w, req)
			return
		}

		http.DefaultServeMux.ServeHTTP(w, req)
	})

	grpcweb := &http.Server{Addr: listenAddress, Handler: handler}
	zap.L().Info("GRPCWeb listen address", zap.String("address", listenAddress))
	go func() {
		if err := grpcweb.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Error("GrpcWeb Serve error.", zap.Error(err))
		}
	}()

	<-ctx.Done()

	if err := grpcweb.Close(); err != nil {
		zap.L().Error("GrpcWeb server close error.", zap.Error(err))
	} else {
		zap.L().Info("GrpcWeb server stopped.")
	}
}
