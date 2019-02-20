package proxy

import (
	"context"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Config struct {
	LocalPort string `json:"local_port"`
	RemoteHost string `json:"remote_host"`
	RemotePort string `json:"remote_port"`
}


var DefaultConfig = &Config{
	LocalPort: "8080",
	RemoteHost: "",
	RemotePort: "",
}

var config = DefaultConfig

func NewProxy(c *Config) {
	config = DefaultConfig
}

func Start() {
	fmt.Println("********************** Proxy Starting **********************")
	e := echo.New()
	e.Debug = true
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 9,
	}))
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		ExposeHeaders: []string{"Authorization"},
	}))

	e.Logger.SetLevel(log.DEBUG)
	e.HideBanner = true

	// set endpoints
	e.GET("/*", request)
	e.POST("/*", request)

	// start server
	go func() {
		if err := e.Start(config.LocalPort); err != nil {
			e.Logger.Info(err)
			e.Logger.Info("shutting down the server")
		}
	}()

	// wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	fmt.Println("********************** Proxy Shutdown **********************")
}

func request(c echo.Context) error {
	return c.String(http.StatusBadGateway, "request failed")
}
