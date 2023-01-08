package rest_api_gin_server

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/api_server"
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/gin-gonic/gin"

	finish "github.com/evgeniums/go-finish-service"
)

type ServerConfig struct {
	API_VERSION  string `validate:"required"`
	MULTITENANCY bool
	common.WithNameBaseConfig

	HOST            string `validate:"omitempty,ip" default:"127.0.0.1"`
	PORT            uint16 `validate:"required"`
	TRUSTED_PROXIES []string
}

type Server struct {
	ServerConfig
	app_context.WithAppBase

	tenanciesById   map[string]*Tenancy
	tenanciesByPath map[string]*Tenancy

	ginEngine *gin.Engine
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Config() interface{} {
	return &s.ServerConfig
}

func (s *Server) ApiVersion() string {
	return s.API_VERSION
}

func (s *Server) IsMultiTenancy() bool {
	return s.MULTITENANCY
}

func (s *Server) Tenancy(id string) (api_server.Tenancy, error) {
	tenancy, ok := s.tenanciesById[id]
	if !ok {
		return nil, errors.New("unknown tenancy")
	}
	return tenancy, nil
}

func (s *Server) AddTenancy(id string) error {
	return errors.New("not implemented yet")
}

func (s *Server) RemoveTenancy(id string) error {
	return errors.New("not implemented yet")
}

func (s *Server) address() string {
	a := fmt.Sprintf("%s:%d", s.HOST, s.PORT)
	return a
}

func (s *Server) apiPath() string {
	p := fmt.Sprintf("/%s", s.API_VERSION)
	return p
}

func ginDefaultLogger(log logger.Logger, notLogged ...string) gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknow"
	}

	var skip map[string]struct{}

	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, p := range notLogged {
			skip[p] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		if _, ok := skip[path]; ok {
			return
		}

		msg := "Unknown GIN handler"
		fields := logger.Fields{
			"hostname":   hostname,
			"statusCode": statusCode,
			"latency":    latency, // time to process
			"clientIP":   clientIP,
			"method":     c.Request.Method,
			"path":       path,
			"referer":    referer,
			"dataLength": dataLength,
			"userAgent":  clientUserAgent,
		}

		if len(c.Errors) > 0 {
			log.Error(msg, errors.New(c.Errors.ByType(gin.ErrorTypePrivate).String()), fields)
		} else {
			if statusCode >= http.StatusInternalServerError {
				log.Error(msg, errors.New("internal server error"), fields)
			} else if statusCode >= http.StatusBadRequest {
				log.Warn(msg, fields)
			} else {
				log.Info(msg, fields)
			}
		}
	}
}

func (s *Server) Init(ctx app_context.Context, configPath ...string) error {

	ctx.Logger().Info("Init REST API gin server")

	s.WithAppBase.Init(ctx)

	// load configuration
	err := object_config.LoadLogValidate(ctx.Cfg(), ctx.Logger(), ctx.Validator(), s, "api_server", configPath...)
	if err != nil {
		return ctx.Logger().Fatal("failed to load server configuration", err, logger.Fields{"name": s.Name()})
	}

	// init gin router
	s.ginEngine = gin.New()
	// trusted proxies are needed for correct logging of client IP address
	s.ginEngine.SetTrustedProxies(s.TRUSTED_PROXIES)
	// use default logger for unhandled paths and then use recovery middleware
	s.ginEngine.Use(ginDefaultLogger(ctx.Logger()), gin.Recovery())

	// done
	return nil
}

func (s *Server) Run(fin *finish.Finisher) {

	srv := &http.Server{Addr: s.address(), Handler: s.ginEngine}
	fin.Add(srv)

	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			msg := "failed to start HTTP server"
			fmt.Printf("%s %s: %s\n", msg, s.Name(), err)
			s.App().Logger().Fatal(msg, err, logger.Fields{"name": s.Name()})
		}
	}()
}

func (s *Server) AddService(id string) error {

	return errors.New("not implemented yet")
}
