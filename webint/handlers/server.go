package handlers

import (
	"context"
	"net/http"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/andproto/models"
	"github.com/averageNetAdmin/andproxy/webint/config"
	"github.com/gin-gonic/gin"
)

func CreateServer(c *gin.Context) {

	ctx := context.Background()

	var server *models.Server
	if err := c.BindJSON(server); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp, err := config.ProxyClient.CreateServer(ctx, &andproto.CreateServerRequest{
		Srv: server,
	})
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, resp.Id)
}

func UpdateServer(c *gin.Context) {

	ctx := context.Background()

	var server *models.Server
	if err := c.BindJSON(server); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	_, err := config.ProxyClient.UpdateServer(ctx, &andproto.UpdateServerRequest{
		Srv: server,
	})
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, true)
}

func DeleteServer(c *gin.Context) {

	ctx := context.Background()

	id := c.GetInt64("id")

	_, err := config.ProxyClient.DeleteServer(ctx, &andproto.DeleteServerRequest{
		Id: id,
	})
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, true)
}

func ReadServers(c *gin.Context) {

	ctx := context.Background()

	limit := c.GetInt64("limit")
	offset := c.GetInt64("offset")
	sortBy := c.GetString("sortby")

	_, err := config.ProxyClient.ReadServers(ctx, &andproto.ReadServersRequest{
		Limit:  limit,
		Offset: offset,
		SortBy: sortBy,
	})
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.IndentedJSON(http.StatusCreated, true)
}
