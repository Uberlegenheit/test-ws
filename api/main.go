package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"
	"websocket-test/conf"
	"websocket-test/helpers/null"
	"websocket-test/service"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/schema"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
	"log"
)

type (
	API struct {
		cfg          conf.Config
		router       *gin.Engine
		server       *http.Server
		services     service.Service
		queryDecoder *schema.Decoder
		auth         *auth.Client
	}

	// Route stores an API route data
	Route struct {
		Path       string
		Method     string
		Func       func(http.ResponseWriter, *http.Request)
		Middleware []negroni.HandlerFunc
	}
)

func NewAPI(c conf.Config, s service.Service) (*API, error) {
	queryDecoder := schema.NewDecoder()
	queryDecoder.IgnoreUnknownKeys(true)
	queryDecoder.RegisterConverter(null.Time{}, func(s string) reflect.Value {
		timestamp, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return reflect.Value{}
		}
		t := null.NewTime(time.Unix(timestamp, 0))
		return reflect.ValueOf(t)
	})

	api := &API{
		cfg:          c,
		services:     s,
		queryDecoder: queryDecoder,
	}

	api.initialize()
	return api, nil
}

func (api *API) Run() error {
	return api.startServe()
}

func (api *API) Stop() error {
	return api.server.Shutdown(context.Background())
}

func (api *API) Title() string {
	return "API"
}

func (api *API) initialize() {
	api.router = gin.Default()

	api.router.Use(gin.Logger())
	api.router.Use(gin.Recovery())

	api.router.Use(cors.New(cors.Config{
		AllowOrigins:     api.cfg.API.CORSAllowedOrigins,
		AllowCredentials: true,
		AllowMethods: []string{
			http.MethodPost, http.MethodHead, http.MethodGet, http.MethodOptions, http.MethodPut, http.MethodDelete,
		},
		AllowHeaders: []string{
			"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token",
			"Authorization", "User-Env", "Access-Control-Request-Headers", "Access-Control-Request-Method",
		},
	}))

	// public routes
	api.router.GET("/", api.Index)
	api.router.GET("/health", api.Health)

	api.router.GET("/market-price", api.MarketPrice)

	api.server = &http.Server{Addr: fmt.Sprintf(":%d", 9000 /*api.cfg.API.ListenOnPort*/), Handler: api.router}
}

func (api *API) startServe() error {
	log.Println("Start listening server on port", zap.Uint64("port", api.cfg.API.ListenOnPort))
	err := api.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("API server was closed")
		return nil
	}
	if err != nil {
		return fmt.Errorf("cannot run API service: %s", err.Error())
	}
	return nil
}
