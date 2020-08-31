package xgrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brian-god/brian-go/pkg/discover"
	"github.com/brian-god/brian-go/pkg/server/xgrpc"
	"github.com/brian-god/brian-go/pkg/xcodec"
	"google.golang.org/grpc"
	"reflect"
	"strings"
	"sync"
)

/**
 *
 * Copyright (C) @2020 hugo network Co. Ltd
 * grpc client
 * @description
 * @updateRemark
 * @author               hugo
 * @updateUser
 * @createDate           2020/8/17 9:51 上午
 * @updateDate           2020/8/17 9:51 上午
 * @version              1.0
**/

// 定义一个存放server链接的集合
var conn sync.Map

// GrpcClient grpc 客户端
type GrpcClient interface {
	//调用远程的服务
	Call(serverName string, int interface{}, ctx context.Context, method string, result interface{}, params ...interface{}) error
}

// ConnGrpcClient 需要传入构建好的链接进行服务的调用
type ConnGrpcClient struct {
	cc *grpc.ClientConn
	//服务发现
	discover *discover.Discover
}

// ServerGrpcClient 服务使用的客户端
type ServerGrpcClient struct {
}

//创建客户端
func NewConnGrpcClient(dis *discover.Discover) *ConnGrpcClient {
	return &ConnGrpcClient{discover: dis}
}

//创建客户端
func NewServerGrpcClient() *ServerGrpcClient {
	return &ServerGrpcClient{}
}

//int 接口
//ctx上下文
//method 调用的方法
//result 返回值 该调用方式暂时只支持单个返回
//params 请求参数
func (c *ServerGrpcClient) Call(serverName string, int interface{}, ctx context.Context, method string, result interface{}, params ...interface{}) error {
	//判断是否有链接了
	serConn, ok := conn.Load(serverName)
	if !ok {
		//获取根据serverName 获取链接
		serConn = getConnByServerName()
		//并将数据存储进集合中
		conn.Store(serverName, serConn)
	}
	return invoke(serConn.(*grpc.ClientConn), int, ctx, method, result, params...)
}

//int 接口
//ctx上下文
//method 调用的方法
//result 返回值 该调用方式暂时只支持单个返回
//params 请求参数
func (c *ConnGrpcClient) Call(serverName string, int interface{}, ctx context.Context, method string, result interface{}, params ...interface{}) error {
	return invoke(c.cc, int, ctx, method, result, params...)
}

// getConnByServerName根据服务名称获取链接，主要用于服务的发现
func getConnByServerName() *grpc.ClientConn {
	conn, err := grpc.Dial("localhost:9092", grpc.WithInsecure())
	if err != nil {
		panic(err.Error())
	}
	return conn
}

// Invoke 具体调用
func invoke(cc *grpc.ClientConn, int interface{}, ctx context.Context, method string, result interface{}, params ...interface{}) error {
	out := new(xgrpc.HugoResponse)
	par := make([]string, 0)
	//request data
	data, err := json.Marshal(params)
	if err != nil {
		return errors.New(err.Error())
	}
	par = append(par, string(data))
	request := &xgrpc.HugoRequest{MethodName: method, Parameters: par}
	//获取类型
	elem := reflect.TypeOf(int).Elem()
	//定义服务调用的地址
	stype := fmt.Sprintf("%s.%s.Grpc/HugoGrpc", elem.PkgPath(), elem.Name())
	err = cc.Invoke(ctx, stype, request, out)
	//如果调用报错直接抛出
	if err != nil {
		return err
	}
	if err := xgrpc.ChickResponse(out); err != nil {
		return err
	}
	//判断返回的是否有数据
	if len(out.Data) == 0 {
		return nil
	}
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("输出参数必须是指针类型")
	}
	//对返回数据进行解码操作
	var resData []interface{}
	if err := json.Unmarshal([]byte(out.Data), &resData); err != nil {
		return err
	}
	//返回值的长度
	lenth := len(resData)
	//定义一个 error
	var reSerr error
	//判断是否有返回
	if lenth > 0 {
		//获取最后一位是返回值类型
		resType := resData[lenth-1]
		//获取真正的返回数据
		resData = resData[:lenth-1]
		//将数据类型转成string
		strResType := resType.(string)
		//将string截取长切片
		strResTypes := strings.Split(strResType, ",")
		for i, v := range strResTypes {
			//返回的数据
			resDataValue := resData[i]
			if "error" == v {
				if nil != resDataValue {
					reSerr = errors.New(resDataValue.(string))
				}
			} else {
				resVale, rerr := xcodec.UnmarshalByType(resDataValue, rv.Elem().Type())
				if nil != rerr {
					return rerr
				}
				//给返回结果赋值
				rv.Elem().Set(resVale)
			}
		}
	}
	//如果结果为nil,这里返回null
	return reSerr
}
