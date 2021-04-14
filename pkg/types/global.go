package types

import (
	"github.com/labstack/echo/v4/middleware"
	"os"
)

// generic response message used to send response to the client,
type RespMessage struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data interface{} `json:"data"`
}

// defines the schema for logging the helpers output while the system is running
var EchoLoggerConfig = middleware.LoggerConfig{
	Skipper: middleware.DefaultSkipper,
	Format: `{"time":"${time_rfc3339_nano}","remote_ip":"${remote_ip}","host":"${host}",` +
		`"method":"${method}","uri":"${uri}","status":${status},"error":"${error}",` +
		`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
		`"bytes_out":${bytes_out}}` + "\n",
	Output: os.Stdout,
}

// server protection against cross-site scripting (XSS) attack, content type sniffing,
// clickjacking, insecure connection and other code injection attacks
var DefaultSecureConfig = middleware.SecureConfig{
	Skipper:            middleware.DefaultSkipper,
	XSSProtection:      "1; mode=block",
	ContentTypeNosniff: "nosniff",
	XFrameOptions:      "SAMEORIGIN",
}

