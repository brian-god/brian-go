package xnacos_registry

import (
	"github.com/brian-god/brian-go/pkg/conf"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/26 4:23 下午
 * @updateDate           2020/8/26 4:23 下午
 * @version              1.0
**/

// RawConfig ...
func RawServerConfig() *constant.ServerConfig {
	var config = DefaultServerConfigs()
	// 解析最外层配置
	if err := conf.UnmarshalKey(key, &config); err != nil {
		xlog.Panic("unmarshal key", xlog.FieldMod("registry.etcd"), xlog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), xlog.FieldErr(err), xlog.String("key", key), xlog.Any("config", config))
	}
	// 解析嵌套配置
	if err := conf.UnmarshalKey(key, &config.Config); err != nil {
		xlog.Panic("unmarshal key", xlog.FieldMod("registry.etcd"), xlog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), xlog.FieldErr(err), xlog.String("key", key), xlog.Any("config", config))
	}
	return config
}
