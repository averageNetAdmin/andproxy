package config

import (
	"github.com/averageNetAdmin/andproxy/andproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var proxyConn *grpc.ClientConn
var ProxyClient andproto.ProxyClient

func ConnectAndProxy(url string) {

	var err error
	proxyConn, err = grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		Logger.Fatal(err)
	}

	ProxyClient = andproto.NewProxyClient(proxyConn)
}
