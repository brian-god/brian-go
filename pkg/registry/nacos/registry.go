package nacos

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brian-god/brian-go/pkg/logger"
	"github.com/brian-god/brian-go/pkg/server"
	"github.com/brian-god/brian-go/pkg/xcodec"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

/**
 * nacos注册中心
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/21 9:25 上午
 * @updateDate           2020/8/21 9:25 上午
 * @version              1.0
**/

type nacosRegistry struct {
	//服务的client
	namingClient *naming_client.INamingClient
	//客户端配置
	conf *constant.ClientConfig
	kvs  sync.Map
	//服务配置
	serConf *constant.ServerConfig
	//logger
	log *logrus.Logger
}

func newETCDRegistry(config *constant.ClientConfig, serverConfig *constant.ServerConfig) *nacosRegistry {
	// Create naming client for service discovery
	namingClient, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfig,
		"clientConfig":  config,
	})
	if nil != err {
		logrus.Panic(logger.FieldMod(xcodec.ErrKindRegisterErr), err.Error())
	}
	res := &nacosRegistry{
		conf:         config,
		serConf:      serverConfig,
		namingClient: &namingClient,
		log:          logrus.New(),
	}
	res.log.Info(logger.FieldMod(xcodec.ModRegistryNacos))
	return res
}

// RegisterService ...
func (e *nacosRegistry) RegisterService(ctx context.Context, info *server.ServiceInfo) error {

	return nil
}

// DeregisterService ...
func (e *nacosRegistry) DeregisterService(ctx context.Context, info *server.ServiceInfo) error {

}

func (e *nacosRegistry) deregister(ctx context.Context, key string) error {

}

// Close ...
func (e *nacosRegistry) Close() error {

}
