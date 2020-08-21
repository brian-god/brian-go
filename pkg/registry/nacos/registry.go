package nacos

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brian-god/brian-go/pkg/server"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
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
	client *naming_client.INamingClient
	//客户端配置
	conf *constant.ClientConfig
	kvs  sync.Map
	//服务配置
	serConf *constant.ServerConfig
}

func newETCDRegistry(config *Config) *nacosRegistry {

}

// RegisterService ...
func (e *nacosRegistry) RegisterService(ctx context.Context, info *server.ServiceInfo) error {

	return err
}

// DeregisterService ...
func (e *nacosRegistry) DeregisterService(ctx context.Context, info *server.ServiceInfo) error {

}

func (e *nacosRegistry) deregister(ctx context.Context, key string) error {

}

// Close ...
func (e *nacosRegistry) Close() error {

}
