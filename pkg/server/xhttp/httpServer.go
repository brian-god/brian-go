package xhttp

import (
	"context"
	"github.com/brian-god/brian-go/pkg"
	"github.com/brian-god/brian-go/pkg/logger"
	"github.com/brian-god/brian-go/pkg/server"
	"github.com/brian-god/brian-go/pkg/xcodec"
	"github.com/labstack/echo/v4"
	"log"
	"net"
	"net/http"
	"os"
)

// http Server struct
type Server struct {
	*echo.Echo
	config   *Config
	listener net.Listener
}

func newServer(config *Config) *Server {
	listener, err := net.Listen("tcp", config.Address())
	if err != nil {
		config.logger.Panic("new xecho server err", logger.FieldErrKind(xcodec.ErrKindListenErr), logger.FieldErr(err))
	}
	config.Port = listener.Addr().(*net.TCPAddr).Port
	return &Server{
		Echo:     echo.New(),
		config:   config,
		listener: listener,
	}
}

// Server implements server.Server interface.
func (s *Server) Serve() error {
	s.Echo.Logger.SetOutput(os.Stdout)
	s.Echo.Debug = s.config.Debug
	s.Echo.HideBanner = true
	//TODO std日志
	s.Echo.StdLogger = log.New(os.Stdout, "hugo", 1)
	for _, route := range s.Echo.Routes() {
		//输出地址信息和处理的方法
		s.config.logger.Info("add route", logger.FieldMethod(route.Method), logger.String("path", route.Path))
	}
	s.Echo.Listener = s.listener
	err := s.Echo.Start("")
	if err != http.ErrServerClosed {
		return err
	}

	s.config.logger.Info("close echo", logger.FieldAddr(s.config.Address()))
	return nil
}

// Stop implements server.Server interface
// it will terminate echo server immediately
//停止具体服务。服务接口
//将立即停止echo服务
func (s *Server) Stop() error {
	return s.Echo.Close()
}

// GracefulStop implements server.Server interface
// it will stop echo server gracefully
//优雅的停止服务，服务接口
//将优雅的停止echo服务
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.Echo.Shutdown(ctx)
}

// Info returns server info, used by governor and consumer balancer
// 初始化服务信息
func (s *Server) Info() *server.ServiceInfo {
	return &server.ServiceInfo{
		Name:      pkg.Name(),
		Scheme:    "http",
		IP:        s.config.Host,
		Port:      s.config.Port,
		Weight:    0.0,
		Enable:    false,
		Healthy:   false,
		Metadata:  map[string]string{},
		Region:    "",
		Zone:      "",
		GroupName: "",
	}
}

//向服务中注册控制器
func (s *Server) UseController(con Controller) {
	//调用controller的注册方法将接口注册到系统中
	con.Register(s)
}
