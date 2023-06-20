package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/dbmng/config"
	"github.com/averageNetAdmin/andproxy/dbmng/grpcsrv"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func main() {

	config.ReadFile("./config.yml")

	config.StartLogging()

	err := config.OpenDB()
	if err != nil {
		err = errors.WithMessage(err, "config.OpenDB()")
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", config.AndConfig.Port))
	if err != nil {
		err = errors.WithMessage(err, "net.Listen()")
		log.Fatal(err)
	}

	gs := grpc.NewServer()
	andproto.RegisterDbmngServer(gs, &grpcsrv.DbSrv{})

	go func() {
		err = gs.Serve(l)
		if err != nil {
			err = errors.WithMessage(err, "gs.Serve()")
			log.Fatal(err)
		}
	}()

	defer gs.GracefulStop()
	config.Glogger.Info("start db")

	stopChan := make(chan os.Signal, 5)

	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	<-stopChan

}
