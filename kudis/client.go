package kudis

import (
	"context"

	pb "bitbucket.org/linkernetworks/aurora/src/kubernetes/kudis/pb"
	"google.golang.org/grpc"
)

// Client is a wrapper for container log piper gRPC client
type Client struct {
	serverAddr string // <host>:<port> of gRPC server
	enableSSL  bool
	conn       *grpc.ClientConn
	client     pb.SubscriptionServiceClient // protobuf generated interface
}

// NewInsecure creates a Job gPRC client (not secure)
func NewInsecure(serverAddr string) (*Client, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Client{
		serverAddr: serverAddr,
		enableSSL:  false,
		conn:       conn,
		client:     pb.NewSubscriptionServiceClient(conn),
	}, nil
}

func (c *Client) SubscribePodLogs(req *pb.PodLogSubscriptionRequest) (*pb.SubscriptionResponse, error) {
	return c.client.SubscribePodLogs(context.Background(), req)
}

// Close connection
func (c *Client) Close() error {
	return c.conn.Close()
}
