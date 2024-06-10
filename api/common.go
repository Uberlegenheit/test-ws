package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (api *API) Index(c *gin.Context) {
	c.String(http.StatusOK, "This is a service 'nuhai bebru'")
}

func (api *API) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
