package registry

import (
	"context"
	"github.com/brian-god/brian-go/pkg/server"
	"io"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/20 6:03 下午
 * @updateDate           2020/8/20 6:03 下午
 * @version              1.0
**/
// ServerInstance ...
type ServerInstance struct {
	Scheme string
	IP     string
	Port   int
	Labels map[string]string
}

// Registry register/deregister service
// registry impl should control rpc timeout
type Registry interface {
	RegisterService(context.Context, *server.ServiceInfo) error
	DeregisterService(context.Context, *server.ServiceInfo) error
	io.Closer
}
