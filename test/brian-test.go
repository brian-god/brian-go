package main

import (
	"cloud.google.com/go/trace/testdata/helloworld"
	"context"
	"fmt"
	"github.com/brian-god/brian-go"
	"github.com/brian-god/brian-go/pkg/server/xgrpc"
	"github.com/brian-god/brian-go/pkg/server/xhttp"
	"github.com/labstack/echo/v4"
	"net/http"
)

var ser = brian.Application{}

//构建一个controller

type TestController struct {
}

func (test *TestController) Register(server *xhttp.Server) {
	server.GET("/index", test.index)
}

//写入一个测试的方法
func (test *TestController) index(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "hello hugo")
}

func main() {
	app := brian.DefaultApplication()
	//注册rpc服务
	app.RegisterRpcServer(new(TestApi), new(TestApiImpl))
	//注册http controller
	app.RegisterController(&TestController{})
	if err := app.Startup(); err != nil {
		fmt.Println("启动有误")
	}
	app.Run()
	/*dir, _ := os.Getwd()
	out := conf.InitConfig(fmt.Sprintf("%s/test/appliction.properties", dir))
	fmt.Println(out)*/
	/*r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	go r.Run() // listen and serve on 0.0.0.0:8080
	server := grpc.NewServer()

	helloworld.RegisterGreeterServer(server, new(Greeter))

	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err.Error())
	}
	server.Serve(lis)*/

	/*if err := ser.Startup(serverHTTP, serveGRPC); err != nil {
		fmt.Println("启动有误")
	}
	ser.Run()*/
}
func Hello() error {
	fmt.Printf("你好")
	return nil
}

//rpc服务
func serveGRPC() error {
	//获取一个grpc服务
	grpcServer := xgrpc.DefaultConfig().Build()
	grpcServer.Register(new(TestApi), new(TestApiImpl))
	//注册服务
	return ser.Serve(grpcServer)
}
func serverHTTP() error {
	httpServer := xhttp.StdConfig("http").Build()
	//使用
	httpServer.UseController(&TestController{})
	//启动服务
	ser.Serve(httpServer)
	return nil
}

type Greeter struct {
	helloworld.GreeterServer
}

//grpc
func (g Greeter) SayHello(context context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{
		Message: "Hello Jupiter",
	}, nil
}

type TestApi interface {
	SayHello(name string) string
}
type TestApiImpl struct {
}

// SayHello
// TODO rpc接口的实现必须使用值接收者不能够使用指针接收者，使用指针接收者会造成结构体是否实现接口的判断出错
func (test TestApiImpl) SayHello(name string) string {
	return fmt.Sprintf("hell %s", name)
}
