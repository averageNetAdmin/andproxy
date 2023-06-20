package main

import (
	"context"
	"fmt"
	"log"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/andproto/models"
	"github.com/averageNetAdmin/andproxy/webint/config"
	"github.com/gin-gonic/gin"
)

func main() {

	config.ReadFile("./config.yml")

	config.StartLogging()

	config.ConnectAndProxy(config.AndConfig.AndProxyUrl)
	fmt.Println("start webint")

	ctx := context.Background()

	srv1, err := config.ProxyClient.CreateServer(ctx, &andproto.CreateServerRequest{
		Srv: &models.Server{
			Address:  "172.16.0.10",
			Port:     80,
			Weight:   1,
			MaxFails: 10,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	srv2, err := config.ProxyClient.CreateServer(ctx, &andproto.CreateServerRequest{
		Srv: &models.Server{
			Address:  "172.16.0.20",
			Port:     80,
			Weight:   1,
			MaxFails: 10,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	srv3, err := config.ProxyClient.CreateServer(ctx, &andproto.CreateServerRequest{
		Srv: &models.Server{
			Address:  "172.16.0.30",
			Port:     80,
			Weight:   1,
			MaxFails: 10,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	pl, err := config.ProxyClient.CreateServerPool(ctx, &andproto.CreateServerPoolRequest{
		Pool: &models.ServerPool{
			Servers: []int64{
				srv1.Id,
				srv2.Id,
				srv3.Id,
			},
			BalancingMethod: "roundrobin",
			Name:            "test_pool",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	pf, err := config.ProxyClient.CreateProxyFilter(ctx, &andproto.CreateProxyFilterRequest{
		Fltr: &models.ProxyFilter{
			ServerPool: pl.Id,
			TargetNet:  "0.0.0.0/0",
			Name:       "test_fltr",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	pth, err := config.ProxyClient.CreateHttpPath(ctx, &andproto.CreateHttpPathRequest{
		Path: &models.HttpPath{
			Path:         ".*",
			ProxyFilters: []int64{pf.Id},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	hst, err := config.ProxyClient.CreateHttpHost(ctx, &andproto.CreateHttpHostRequest{
		Host: &models.HttpHost{
			Address: ".*",
			Paths:   []int64{pth.Id},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	hndlr, err := config.ProxyClient.CreateHttpHandler(ctx, &andproto.CreateHttpHandlerRequest{
		Hndlr: &models.HttpHandler{
			Hosts:  []int64{hst.Id},
			Name:   "test_hndlr",
			Status: models.HttpHandler_Listen,
			Port:   80,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	_ = hndlr

	router := gin.Default()

	err = router.Run(fmt.Sprintf("localhost:%d", config.AndConfig.Port))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("as")
}
