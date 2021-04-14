package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"os"
	"uwdiscorwb/v1/pkg/config"
	"uwdiscorwb/v1/pkg/handlers"
	"uwdiscorwb/v1/pkg/log"
	"uwdiscorwb/v1/pkg/types"
)

func init() {
	properties, err := config.GetConfig()
	if err != nil {
		log.Error("failed to initialized server, exit with error - ", err)
		os.Exit(1)
	}
	if properties.DiscordWebhookUrl == "" {
		log.Error("discord webhook url is required, you can pass an environment variable to discord_webhook_url")
		os.Exit(1)
	}
}

func main() {
	// server properties for starting the server
	var properties, _ = config.GetConfig()

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("5M"))
	e.Use(middleware.LoggerWithConfig(types.EchoLoggerConfig))
	if properties.SecureServer {
		e.Use(middleware.SecureWithConfig(types.DefaultSecureConfig))
	}

	// Routes
	e.GET("/health-check", handlers.HealthCheck)
	e.POST("/flowroute", handlers.FlowrouteCallbackHandler)
	e.GET("/voipms", handlers.VoipmsCallbackHandler)

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", properties.Port)))
}
