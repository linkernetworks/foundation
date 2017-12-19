package grpc

import (
	"context"

	pb "bitbucket.org/linkernetworks/aurora/src/grpc/messages"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// JobGrpcClient is a wrapper for Job gRPC client
type JobGrpcClient struct {
	serverAddr string // <host>:<port> of gRPC server
	enableSSL  bool
	conn       *grpc.ClientConn
	grpcClient pb.JobClient // protobuf generated interface
	// add cert when SSL is needed
	sslCrt string // cert path
}

// NewInsecureJobClient creates a Job gPRC client (not secure)
func NewInsecureJobClient(serverAddr string) (*JobGrpcClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &JobGrpcClient{
		serverAddr: serverAddr,
		enableSSL:  false,
		conn:       conn,
		grpcClient: pb.NewJobClient(conn),
	}, nil
}

// NewSSLJobClient creates a Job gRPC client with SSL
func NewSSLJobClient(serverAddr string, sslCrt string) (client *JobGrpcClient, err error) {
	// gRPC SSL Reference: https://bbengfort.github.io/programmer/2017/03/03/secure-grpc.html
	creds, err := credentials.NewClientTLSFromFile(sslCrt, "")
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}
	return &JobGrpcClient{
		serverAddr: serverAddr,
		enableSSL:  true,
		sslCrt:     sslCrt,
		conn:       conn,
		grpcClient: pb.NewJobClient(conn),
	}, nil
}

// EnqueueJob call job enqueue
func (c *JobGrpcClient) EnqueueJob(req *pb.JobRequest) (*pb.JobResponse, error) {
	return c.grpcClient.EnqueueJob(context.Background(), req) // TODO context with
}

func (c *JobGrpcClient) StopJob(req *pb.StopJobRequest) (*pb.JobResponse, error) {
	return c.grpcClient.StopJob(context.Background(), req) // TODO context with
}

// Close connection
func (c *JobGrpcClient) Close() error {
	return c.conn.Close()
}
