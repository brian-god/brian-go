package xnacos_client

import (
	"github.com/brian-god/brian-go/pkg/logger"
	"github.com/brian-god/brian-go/pkg/xcodec"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/sirupsen/logrus"
	"sync"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/27 3:33 下午
 * @updateDate           2020/8/27 3:33 下午
 * @version              1.0
**/
// nacos客户端
type NacosClient struct {
	//服务的client
	namingClient *naming_client.INamingClient
	//配置的client
	configClient *config_client.IConfigClient
	//客户端配置
	clientConf *constant.ClientConfig
	kvs        sync.Map
	//服务配置
	serConfigs []constant.ServerConfig
	//logger
	log *logrus.Logger
}

// 获取INamingClient
func (client *NacosClient) GetNamingClient() naming_client.INamingClient {
	return *client.namingClient
}

//DefaultClientConfig创建一个默认的client配置
func DefaultClientConfig() *constant.ClientConfig {
	clientConfig := constant.ClientConfig{
		NamespaceId:         "e525eafa-f7d7-4029-83d9-008937f9d468", // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "info",
	}
	return &clientConfig
}

//创建nacosclient
func newClient(client *constant.ClientConfig, server []constant.ServerConfig) *NacosClient {
	// 创建服务发现客户端
	namingClient, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": *client,
		"clientConfig":  server,
	})
	if nil != err {
		logrus.Panic("create naming client error ", logger.FieldMod(xcodec.ModConfig), logger.FieldErr(err))
	}

	// 创建动态配置客户端
	configClient, err1 := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": *client,
		"clientConfig":  server,
	})
	if nil != err1 {
		logrus.Panic("create config client error ", logger.FieldMod(xcodec.ModConfig), logger.FieldErr(err))
	}
	nacosClient := &NacosClient{
		namingClient: &namingClient,
		configClient: &configClient,
		clientConf:   client,
		serConfigs:   server,
		log:          logrus.New(),
	}
	return nacosClient
}
