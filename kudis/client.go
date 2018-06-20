package kudis

import (
	pb "github.com/linkernetworks/kube/kudis/pb"
	"google.golang.org/grpc"
)

// Client is a wrapper for container log piper gRPC client
type Client struct {
	pb.SubscriptionServiceClient        // protobuf generated interface
	serverAddr                   string // <host>:<port> of gRPC server
	enableSSL                    bool
	conn                         *grpc.ClientConn
}

// NewInsecure creates a Job gPRC client (not secure)
func NewInsecure(serverAddr string) (*Client, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Client{
		SubscriptionServiceClient: pb.NewSubscriptionServiceClient(conn),
		serverAddr:                serverAddr,
		enableSSL:                 false,
		conn:                      conn,
	}, nil
}

// Close connection
func (c *Client) Close() error {
	return c.conn.Close()
}
