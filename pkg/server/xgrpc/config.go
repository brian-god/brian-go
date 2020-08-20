package xgrpc

import (
	"fmt"
	"github.com/brian-god/brian-go/pkg/conf"
	"github.com/brian-god/brian-go/pkg/xcast"
	"github.com/labstack/gommon/color"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/4 3:17 下午
 * @updateDate           2020/8/4 3:17 下午
 * @version              1.0
**/
// TODO 日志需要单独处理
//Grpc 的配置结构体
type Config struct {
	Host string
	Port int
	// Network network type, tcp4 by default
	Network string `json:"network" toml:"network"`
	// DisableTrace disbale Trace Interceptor, false by default
	//禁用跟踪器默认为true
	DisableTrace bool
	// DisableMetric disable Metric Interceptor, false by default
	//禁用监听器默认为true
	DisableMetric bool
	// SlowQueryThresholdInMilli, request will be colored if cost over this threshold value
	SlowQueryThresholdInMilli int64
	serverOptions             []grpc.ServerOption
	streamInterceptors        []grpc.StreamServerInterceptor
	unaryInterceptors         []grpc.UnaryServerInterceptor
	colorer                   *color.Color
	//TODO 日志
	logger *logrus.Logger
}

// DefaultConfig represents default config
// User should construct config base on DefaultConfig
//grpc默认的配置
//用户不做调整则使用默认的配置
func DefaultConfig() *Config {
	return &Config{
		Network:                   "tcp4",
		Host:                      "127.0.0.1",
		Port:                      9090,
		DisableMetric:             true,
		DisableTrace:              true,
		SlowQueryThresholdInMilli: 500,
		logger:                    logrus.New(),
		colorer:                   color.New(),
		serverOptions:             []grpc.ServerOption{},
		//流方法，流拦截器
		streamInterceptors: []grpc.StreamServerInterceptor{},
		//grpc中使用一元拦截器
		unaryInterceptors: []grpc.UnaryServerInterceptor{},
	}
}

// hugo Standard HTTP Server config
func StdConfig() *Config {
	return RawConfig()
}

// RawConfig ...
func RawConfig() *Config {
	var config = DefaultConfig()
	//协议
	if v := conf.Get("brian.rpc.server.Network"); v != nil {
		if v, err := xcast.ToStringE(v); nil == err {
			config.Network = v
		}
	}
	//端口
	if v := conf.Get("brian.rpc.server.port"); v != nil {
		if intValue, err := xcast.ToIntE(v); nil == err {
			config.Port = intValue
		}
	}
	//ip
	if v := conf.Get("brian.rpc.server.host"); v != nil {
		if v, err := xcast.ToStringE(v); nil == err {
			config.Host = v
		}
	}
	//监听
	if v := conf.Get("brian.rpc.server.DisableMetric"); v != nil {
		if v, err := xcast.ToBoolE(v); nil == err {
			config.DisableMetric = v
		}
	}
	//追踪
	if v := conf.Get("brian.rpc.server.DisableTrace"); v != nil {
		if v, err := xcast.ToBoolE(v); nil == err {
			config.DisableTrace = v
		}
	}
	//超时
	if v := conf.Get("brian.rpc.server.timeout"); v != nil {
		if v, err := xcast.ToInt64E(v); nil == err {
			config.SlowQueryThresholdInMilli = v
		}
	}
	//日志级别
	if v := conf.Get("brian.rpc.server.log.level"); v != nil {
		if v, err := xcast.ToStringE(v); nil == err {
			if level, err := logrus.ParseLevel(v); nil == err {
				config.logger.Level = level
			}
		}
	}
	return config
}

// Build ...
func (config *Config) Build() *Server {
	//TODO
	if !config.DisableTrace {
		//config.unaryInterceptors = append(config.unaryInterceptors, traceUnaryServerInterceptor)
		//config.streamInterceptors = append(config.streamInterceptors, traceStreamServerInterceptor)
	}

	if !config.DisableMetric {
		//config.unaryInterceptors = append(config.unaryInterceptors, prometheusUnaryServerInterceptor)
		//config.streamInterceptors = append(config.streamInterceptors, prometheusStreamServerInterceptor)
	}

	return newServer(config)
}

// Address ...
//用来获取组装完成的服务地址
func (config Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
