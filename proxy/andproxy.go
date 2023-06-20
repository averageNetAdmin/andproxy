package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/grpcsrv"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func main() {
	// Create handler from file and run it
	// TODO: add code for reading all handlers from directory

	config.ReadFile("./config.yml")

	config.StartLogging()

	// err := config.OpenDB()
	// if err != nil {
	// 	err = errors.WithMessage("config.OpenDB()/%s", err)
	// 	log.Fatal(err)
	// }

	config.ConnectDbMng(config.AndConfig.DbMngUrl)
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", config.AndConfig.Port))
	if err != nil {
		err = errors.WithMessage(err, "net.Listen()")
		config.Glogger.Fatal(err)
		return
	}

	gs := grpc.NewServer()
	andproto.RegisterProxyServer(gs, &grpcsrv.Proxy{})

	go func() {
		err = gs.Serve(l)
		if err != nil {
			err = errors.WithMessage(err, "net.Listen()")
			config.Glogger.Fatal(err)
			return
		}
	}()

	defer gs.GracefulStop()

	config.Glogger.Info("start proxy")

	stopChan := make(chan os.Signal, 5)

	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	<-stopChan

}
