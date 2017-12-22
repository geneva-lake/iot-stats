package server

import (
	"iot-stats/service"
	"net"

	"github.com/gin-gonic/gin"
)

// Config Server configuration parameters
type Config struct {
	Port       string
	Host       string
	ApiKey     string
	Expiration int
}

func (c Config) GetAddr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

// Server aggregate data and method
type Server struct {
	config *Config
	ms     service.MongoInterface
}

// NewServer return new instance of Server
func NewServer(c *Config, ms service.MongoInterface) *Server {
	return &Server{config: c, ms: ms}
}

func (s *Server) Serve() error {
	api := newApi(s.config.ApiKey, s.ms)
	web := newWeb(s.config.Expiration, s.ms)
	login := newLogin(s.config.Expiration, s.ms)
	router := gin.Default()
	router.Use(gin.Recovery())
	router.POST("/login", login.loginHandler)
	a := router.Group("/api")
	a.Use(api.checkApiKey)
	a.POST("/register", api.registerDevice)
	a.POST("/error", api.errorReport)
	a.StaticFile("/firmware", "././build")
	w := router.Group("/web")
	w.Use(web.checkSession)
	w.GET("/list/:skip/:limit", web.getDevices)
	err := router.RunTLS(s.config.GetAddr(), "server.pem", "server.key")
	if err != nil {
		return err
	}
	return nil
}
