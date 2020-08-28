package nacos_discover

import (
	"context"
	"github.com/brian-god/brian-go/pkg/client/xnacos_client"
	"github.com/brian-god/brian-go/pkg/discover"
	"github.com/brian-god/brian-go/pkg/server"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

/**
 * Copyright (C) @2020 hugo network Co. Ltd
 * nacos 的服务发现
 * @author: hugo
 * @version: 1.0
 * @date: 2020/8/27
 * @time: 21:35
 * @description:
 */

// nacos的实体
type NacoseDiscover struct {
	naming_client.INamingClient
}

func CreateNacoseDiscover(client *xnacos_client.NacosClient) *NacoseDiscover {
	return &NacoseDiscover{client.GetNamingClient()}
}

// 服务发现
func (n NacoseDiscover) GetServerInstance(con context.Context, param *discover.ServerInstancesParam) (*server.ServiceInfo, error) {
	// SelectOneHealthyInstance将会按加权随机轮训的负载均衡策略返回一个健康的实例
	// 实例必须满足的条件：health=true,enable=true and weight>0
	instance, err := n.SelectOneHealthyInstance(vo.SelectOneHealthInstanceParam{
		ServiceName: param.ServiceName,
		GroupName:   param.GroupName, // 默认值DEFAULT_GROUP
		Clusters:    param.Clusters,  // 默认值DEFAULT
	})
	if nil != err {
		return nil, err
	}
	ser := server.ServiceInfo{
		Name:        instance.ServiceName,
		Scheme:      instance.ServiceName,
		IP:          instance.Ip,
		Port:        int(instance.Port),
		Weight:      instance.Weight,
		Enable:      instance.Enable,
		Healthy:     instance.Healthy,
		Ephemeral:   instance.Ephemeral,
		Metadata:    instance.Metadata,
		Region:      "",
		Zone:        "",
		GroupName:   "",
		ClusterName: "",
	}
	return &ser, nil
}
