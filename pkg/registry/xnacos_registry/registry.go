package xnacos_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brian-god/brian-go/pkg/client/xnacos_client"
	"github.com/brian-god/brian-go/pkg/logger"
	"github.com/brian-god/brian-go/pkg/server"
	"github.com/brian-god/brian-go/pkg/xcodec"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
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

// nacos的实体
type nacosRegister struct {
	naming_client.INamingClient
}

func newNacosRegister(client *xnacos_client.NacosClient) *nacosRegister {
	return &nacosRegister{client.GetNamingClient()}
}

//DefaultClientConfig创建一个默认的sever配置
func DefaultServerConfigs() []constant.ServerConfig {
	// 至少一个ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "127.0.0.1",
			ContextPath: "/nacos",
			Port:        80,
		},
	}
	return serverConfigs
}

// RegisterService ... 服务注册
func (e *nacosRegister) RegisterService(ctx context.Context, info *server.ServiceInfo) error {
	ok, err := e.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          info.IP,
		Port:        uint64(info.Port),
		ServiceName: info.Name,
		Weight:      info.Weight,
		Enable:      info.Enable,
		Healthy:     info.Healthy,
		Ephemeral:   info.Ephemeral,
		Metadata:    info.Metadata,
		ClusterName: info.ClusterName, // 默认值DEFAULT
		GroupName:   info.GroupName,   // 默认值DEFAULT_GROUP
	})
	if !ok {
		return err
	}
	return nil
}

// DeregisterService ... 注销服务
func (e *nacosRegister) DeregisterService(ctx context.Context, info *server.ServiceInfo) error {
	ok, err := e.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          info.IP,
		Port:        uint64(info.Port),
		ServiceName: info.Name,
		Ephemeral:   info.Ephemeral,
		Cluster:     info.ClusterName, // 默认值DEFAULT
		GroupName:   info.GroupName,   // 默认值DEFAULT_GROUP
	})
	if !ok {
		return err
	}
	return nil
}

// Close ...
func (e *nacosRegister) Close() error {

}
