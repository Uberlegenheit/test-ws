package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

var upg = websocket.Upgrader{
	ReadBufferSize:  102400,
	WriteBufferSize: 102400,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (api *API) MarketPrice(c *gin.Context) {
	uid := c.Query("user_id")
	if len(uid) == 0 {
		handleError(c, fmt.Errorf("you have to provide user_id"), http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(uid, 10, 64)
	if err != nil {
		handleError(c, fmt.Errorf("invalid user_id value"), http.StatusBadRequest)
		return
	}

	conn, err := upg.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		handleError(c, fmt.Errorf("could not open websocket connection"), http.StatusBadRequest)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	go func() {
		for {
			_, _, err = conn.ReadMessage()
			if err != nil {
				log.Printf("Connection closed: %v\n ", err)
				break
			}
			time.Sleep(time.Millisecond * 500)
		}
		cancel()
	}()

	api.services.AddConnectionToMarketDataStream(ctx, userID, conn)
}

func handleError(c *gin.Context, err error, statusCode int) {
	log.Println("[api] Error ", err)
	c.JSON(statusCode, gin.H{"error": err.Error()})
}
