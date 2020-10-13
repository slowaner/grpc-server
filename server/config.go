package server

// Config configuration struct for grpc server wrapper
type Config struct {
	// Enables listening for grpc
	GrpcEnable bool
	// Grpc listening host
	GrpcHost string
	// Grpc listening port
	GrpcPort uint8
	// Enable grpc listening via tls
	GrpcEnableTls bool

	// Enables listening for grpc-web
	WebEnable bool
	// Grpc-web listening host
	WebHost string
	// Grpc-web listening port
	WebPort uint8
	// Enable grpc-web listening via tls
	WebEnableTls bool
}
