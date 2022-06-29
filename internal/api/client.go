package api

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	apipb "github.com/tcfw/didem/api"
	"google.golang.org/grpc"
)

type Client struct {
	cc *grpc.ClientConn
}

func (a *Client) Close() error {
	return a.cc.Close()
}

func (a *Client) P2P() apipb.P2PServiceClient {
	return apipb.NewP2PServiceClient(a.cc)
}

func (a *Client) Em() apipb.EmServiceClient {
	return apipb.NewEmServiceClient(a.cc)
}

func NewClient() (*Client, error) {
	cc, err := grpc.Dial(viper.GetString("daemon_addr"), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "connecting to daemon")
	}

	return &Client{cc: cc}, nil
}
