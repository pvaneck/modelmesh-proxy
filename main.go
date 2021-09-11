package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	gw "github.com/pvaneck/modelmesh-proxy/gen"
)

var (
	// command-line options:
	// gRPC server endpoint
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:8033", "gRPC server endpoint")
	logger             = zap.New()
	listenPort         = 8008
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	marshaller := &CustomJSONPb{}
	marshaller.EmitUnpopulated = false
	marshaller.DiscardUnknown = false

	// Register gRPC server endpoint
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaller),
	)
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}
	logger.Info("Registering gRPC Inference Service Handler...")
	err := gw.RegisterGRPCInferenceServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	if port, ok := os.LookupEnv("LISTEN_PORT"); ok {
		listenPort, err = strconv.Atoi(port)
		if err != nil {
			logger.Error(err, "unable to parse LISTEN_PORT environment variable")
			os.Exit(1)
		}
	}
	logger.Info(fmt.Sprintf("Listening on port %d", listenPort))

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(fmt.Sprintf(":%d", listenPort), mux)
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		logger.Error(err, "unable to start gRPC REST proxy")
		os.Exit(1)
	}
}
