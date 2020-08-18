package xhttp

import (
	"fmt"
	"github.com/brian-god/brian-go/pkg/conf"
	"github.com/brian-god/brian-go/pkg/logger"
	"github.com/brian-god/brian-go/pkg/xcodec"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

/**
 * Copyright (C) @2020 hugo network Co. Ltd
 *
 * @author: hugo
 * @version: 1.0
 * @date: 2020/8/2
 * @time: 11:56
 * @description:
 */

// HTTP 服务配置类

// HTTP config
type Config struct {
	Host          string
	Port          int
	Debug         bool
	DisableMetric bool
	DisableTrace  bool

	SlowQueryThresholdInMilli int64
	//TODO 日志
	logger *logrus.Logger
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Host:                      "127.0.0.1",
		Port:                      8080,
		Debug:                     false,
		SlowQueryThresholdInMilli: 500, // 500ms
		logger:                    logrus.New(),
	}
}

// hugo Standard HTTP Server config
func StdConfig(name string) *Config {
	return RawConfig("Hugo.server." + name)
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil &&
		errors.Cause(err) != conf.ErrInvalidKey {
		config.logger.Panic("http server parse config panic", logger.FieldErrKind(xcodec.ErrKindUnmarshalConfigErr), logger.FieldErr(err), logger.FieldKey(key))
	}
	return config
}

// 修改日志配置 ...
func (config *Config) WithLogger(logger *logrus.Logger) *Config {
	config.logger = logger
	return config
}

// WithHost ...
func (config *Config) WithHost(host string) *Config {
	config.Host = host
	return config
}

// WithPort ...
func (config *Config) WithPort(port int) *Config {
	config.Port = port
	return config
}

// Build create server instance, then initialize it with necessary interceptor
func (config *Config) Build() *Server {
	server := newServer(config)
	//TODO 中间件注册
	//server.Use(recoverMiddleware(config.logger, config.SlowQueryThresholdInMilli))

	if !config.DisableMetric {
		//	server.Use(metricServerInterceptor())
	}

	if !config.DisableTrace {
		//server.Use(traceServerInterceptor())
	}
	return server
}

// Address ...
func (config *Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
