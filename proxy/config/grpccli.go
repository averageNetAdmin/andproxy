package config

import (
	"fmt"

	"github.com/averageNetAdmin/andproxy/andproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var dbMngConn *grpc.ClientConn
var DbMngClient andproto.DbmngClient

func ConnectDbMng(url string) {

	fmt.Println(url)

	var err error
	dbMngConn, err = grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		Glogger.Fatal("grpc.Dial(): ", err)
	}

	DbMngClient = andproto.NewDbmngClient(dbMngConn)
}
